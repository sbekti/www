package util

import (
	"strings"

	"github.com/labstack/echo/v4"
)

// AuthInfo holds the authentication information from reverse proxy headers
type AuthInfo struct {
	Username string   `json:"username"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Groups   []string `json:"groups"`
}

// AuthMiddleware parses authentication headers from reverse proxy and adds them to context
func AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authInfo := &AuthInfo{
				Username: c.Request().Header.Get("Remote-User"),
				Name:     c.Request().Header.Get("Remote-Name"),
				Email:    c.Request().Header.Get("Remote-Email"),
			}

			// Parse groups from comma-separated string
			groupsHeader := c.Request().Header.Get("Remote-Groups")
			if groupsHeader != "" {
				groups := strings.Split(groupsHeader, ",")
				for i, group := range groups {
					groups[i] = strings.TrimSpace(group)
				}
				authInfo.Groups = groups
			}

			// Store auth info in context
			c.Set("auth", authInfo)

			return next(c)
		}
	}
}

// GetAuthInfo retrieves authentication information from context
func GetAuthInfo(c echo.Context) *AuthInfo {
	if auth := c.Get("auth"); auth != nil {
		if authInfo, ok := auth.(*AuthInfo); ok {
			return authInfo
		}
	}
	return &AuthInfo{
		Username: "unknown",
		Name:     "Unknown User",
		Email:    "",
		Groups:   []string{},
	}
} 