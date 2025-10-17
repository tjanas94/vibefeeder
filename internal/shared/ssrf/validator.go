package ssrf

import (
	"errors"
	"net"
)

var (
	// ErrPrivateIP is returned when URL resolves to a private IP address
	ErrPrivateIP = errors.New("access to private IP addresses is forbidden")

	// ErrLocalhostAccess is returned when URL points to localhost
	ErrLocalhostAccess = errors.New("access to localhost is forbidden")

	// ErrLinkLocalAccess is returned when URL points to link-local address
	ErrLinkLocalAccess = errors.New("access to link-local addresses is forbidden")
)

// ValidateIP checks if an IP address is safe to access.
// It returns an error if the IP is a loopback, link-local, private, unspecified, or multicast address.
func ValidateIP(ip net.IP) error {
	// Check for loopback addresses (127.0.0.0/8 for IPv4, ::1 for IPv6)
	if ip.IsLoopback() {
		return ErrLocalhostAccess
	}

	// Check for link-local addresses (169.254.0.0/16 for IPv4, fe80::/10 for IPv6)
	// This includes cloud metadata endpoints (169.254.169.254, 169.254.169.253)
	if ip.IsLinkLocalUnicast() {
		return ErrLinkLocalAccess
	}

	// Check for private addresses (RFC 1918)
	if ip.IsPrivate() {
		return ErrPrivateIP
	}

	// Check for unspecified address (0.0.0.0 or ::)
	if ip.IsUnspecified() {
		return ErrPrivateIP
	}

	// Check for multicast addresses
	if ip.IsMulticast() {
		return ErrPrivateIP
	}

	return nil
}
