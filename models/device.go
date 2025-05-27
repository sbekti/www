package models

// Device represents a network device with its associated VLAN
type Device struct {
	MAC         string `json:"mac"`
	Description string `json:"description"`
	VLAN        string `json:"vlan"`
}

// ValidateVLAN checks if the VLAN is one of the allowed values
func (d *Device) ValidateVLAN() bool {
	switch d.VLAN {
	case "trusted", "iot", "guest":
		return true
	default:
		return false
	}
}

// ValidateMAC checks if the MAC address is in the correct format
func (d *Device) ValidateMAC() bool {
	// Simple validation for now - just check length and hex characters
	if len(d.MAC) != 12 {
		return false
	}
	for _, c := range d.MAC {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
} 