package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sbekti/www/models"
	"github.com/sbekti/www/util"
)

// DeviceHandler handles device-related HTTP requests
type DeviceHandler struct{}

// NewDeviceHandler creates a new device handler
func NewDeviceHandler() *DeviceHandler {
	return &DeviceHandler{}
}

// ListDevices handles GET /devices
func (h *DeviceHandler) ListDevices(c echo.Context) error {
	db := util.GetDB()
	
	rows, err := db.Query(c.Request().Context(), `
		SELECT u.username, u.description, r.groupname
		FROM users u
		JOIN radusergroup r ON u.username = r.username
		ORDER BY u.username
	`)
	if err != nil {
		c.Logger().Errorf("ListDevices: db.Query error: %v", err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error fetching devices: %v", err))
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var d models.Device
		if err := rows.Scan(&d.MAC, &d.Description, &d.VLAN); err != nil {
			c.Logger().Errorf("ListDevices: rows.Scan error: %v", err)
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Error scanning device data: %v", err))
		}
		devices = append(devices, d)
	}
	if err := rows.Err(); err != nil {
		c.Logger().Errorf("ListDevices: rows.Err: %v", err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error iterating over rows: %v", err))
	}

	return c.Render(http.StatusOK, "devices/index.html", devices)
}

// ShowAddForm handles GET /devices/add
func (h *DeviceHandler) ShowAddForm(c echo.Context) error {
	return c.Render(http.StatusOK, "devices/add.html", nil)
}

// AddDevice handles POST /devices/add
func (h *DeviceHandler) AddDevice(c echo.Context) error {
	authInfo := util.GetAuthInfo(c)
	device := models.Device{
		MAC:         c.FormValue("mac"),
		Description: c.FormValue("description"),
		VLAN:        c.FormValue("vlan"),
	}

	if !device.ValidateMAC() || !device.ValidateVLAN() {
		c.Logger().Warnf("AddDevice: Invalid device data by user %s: MAC=%s, VLAN=%s", authInfo.Username, device.MAC, device.VLAN)
		return c.String(http.StatusBadRequest, "Invalid device data")
	}

	db := util.GetDB()
	ctx := c.Request().Context()

	// Start transaction
	tx, err := db.Begin(ctx)
	if err != nil {
		c.Logger().Errorf("AddDevice: db.Begin error for user %s: %v", authInfo.Username, err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error starting transaction: %v", err))
	}
	defer tx.Rollback(ctx) // Rollback is a no-op if Commit has been called

	// Insert into users table
	_, err = tx.Exec(ctx, `
		INSERT INTO users (username, description)
		VALUES ($1, $2)
	`, device.MAC, device.Description)
	if err != nil {
		c.Logger().Errorf("AddDevice: tx.Exec users error for user %s: %v", authInfo.Username, err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error inserting device: %v", err))
	}

	// Insert into radusergroup table
	_, err = tx.Exec(ctx, `
		INSERT INTO radusergroup (username, groupname)
		VALUES ($1, $2)
	`, device.MAC, device.VLAN)
	if err != nil {
		c.Logger().Errorf("AddDevice: tx.Exec radusergroup error for user %s: %v", authInfo.Username, err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error inserting VLAN group: %v", err))
	}

	// Insert into radcheck table
	_, err = tx.Exec(ctx, `
		INSERT INTO radcheck (username, attribute, op, value)
		VALUES ($1, 'Cleartext-Password', ':=', $1)
	`, device.MAC)
	if err != nil {
		c.Logger().Errorf("AddDevice: tx.Exec radcheck error for user %s: %v", authInfo.Username, err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error inserting radcheck: %v", err))
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		c.Logger().Errorf("AddDevice: tx.Commit error for user %s: %v", authInfo.Username, err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error committing transaction: %v", err))
	}

	c.Logger().Infof("Device added by user %s (%s): MAC=%s", authInfo.Username, authInfo.Name, device.MAC)
	return c.Redirect(http.StatusSeeOther, "/devices")
}

// ShowEditForm handles GET /devices/edit/:mac
func (h *DeviceHandler) ShowEditForm(c echo.Context) error {
	mac := c.Param("mac")
	db := util.GetDB()

	var device models.Device
	err := db.QueryRow(c.Request().Context(), `
		SELECT u.username, u.description, r.groupname
		FROM users u
		JOIN radusergroup r ON u.username = r.username
		WHERE u.username = $1
	`, mac).Scan(&device.MAC, &device.Description, &device.VLAN)
	if err != nil {
		c.Logger().Warnf("ShowEditForm: Device not found or db error: MAC=%s, error: %v", mac, err)
		return c.String(http.StatusNotFound, fmt.Sprintf("Device not found: %s", mac))
	}

	return c.Render(http.StatusOK, "devices/edit.html", device)
}

// UpdateDevice handles POST /devices/edit/:mac
func (h *DeviceHandler) UpdateDevice(c echo.Context) error {
	authInfo := util.GetAuthInfo(c)
	mac := c.Param("mac")
	device := models.Device{
		MAC:         mac,
		Description: c.FormValue("description"),
		VLAN:        c.FormValue("vlan"),
	}

	if !device.ValidateVLAN() {
		c.Logger().Warnf("UpdateDevice: Invalid VLAN by user %s: %s for MAC=%s", authInfo.Username, device.VLAN, mac)
		return c.String(http.StatusBadRequest, "Invalid VLAN")
	}

	db := util.GetDB()
	ctx := c.Request().Context()

	// Start transaction
	tx, err := db.Begin(ctx)
	if err != nil {
		c.Logger().Errorf("UpdateDevice: db.Begin error for user %s: %v", authInfo.Username, err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error starting transaction: %v", err))
	}
	defer tx.Rollback(ctx)

	// Update users table
	_, err = tx.Exec(ctx, `
		UPDATE users
		SET description = $1
		WHERE username = $2
	`, device.Description, device.MAC)
	if err != nil {
		c.Logger().Errorf("UpdateDevice: tx.Exec users error for user %s: %v", authInfo.Username, err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error updating device: %v", err))
	}

	// Update radusergroup table
	_, err = tx.Exec(ctx, `
		UPDATE radusergroup
		SET groupname = $1
		WHERE username = $2
	`, device.VLAN, device.MAC)
	if err != nil {
		c.Logger().Errorf("UpdateDevice: tx.Exec radusergroup error for user %s: %v", authInfo.Username, err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error updating VLAN group: %v", err))
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		c.Logger().Errorf("UpdateDevice: tx.Commit error for user %s: %v", authInfo.Username, err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error committing transaction: %v", err))
	}

	c.Logger().Infof("Device updated by user %s (%s): MAC=%s", authInfo.Username, authInfo.Name, device.MAC)
	return c.Redirect(http.StatusSeeOther, "/devices")
}

// DeleteDevice handles POST /devices/delete/:mac
func (h *DeviceHandler) DeleteDevice(c echo.Context) error {
	authInfo := util.GetAuthInfo(c)
	mac := c.Param("mac")
	db := util.GetDB()
	ctx := c.Request().Context()

	// Start transaction
	tx, err := db.Begin(ctx)
	if err != nil {
		c.Logger().Errorf("DeleteDevice: db.Begin error for user %s: %v", authInfo.Username, err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error starting transaction: %v", err))
	}
	defer tx.Rollback(ctx)

	// Delete from all tables
	_, err = tx.Exec(ctx, "DELETE FROM users WHERE username = $1", mac)
	if err != nil {
		c.Logger().Errorf("DeleteDevice: tx.Exec users error for user %s: %v", authInfo.Username, err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error deleting device: %v", err))
	}

	_, err = tx.Exec(ctx, "DELETE FROM radusergroup WHERE username = $1", mac)
	if err != nil {
		c.Logger().Errorf("DeleteDevice: tx.Exec radusergroup error for user %s: %v", authInfo.Username, err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error deleting VLAN group: %v", err))
	}

	_, err = tx.Exec(ctx, "DELETE FROM radcheck WHERE username = $1", mac)
	if err != nil {
		c.Logger().Errorf("DeleteDevice: tx.Exec radcheck error for user %s: %v", authInfo.Username, err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error deleting radcheck: %v", err))
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		c.Logger().Errorf("DeleteDevice: tx.Commit error for user %s: %v", authInfo.Username, err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error committing transaction: %v", err))
	}

	c.Logger().Infof("Device deleted by user %s (%s): MAC=%s", authInfo.Username, authInfo.Name, mac)
	return c.Redirect(http.StatusSeeOther, "/devices")
} 