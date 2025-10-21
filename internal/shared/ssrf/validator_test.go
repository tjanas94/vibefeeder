package ssrf

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidateIP_Loopback tests IPv4 loopback address detection
func TestValidateIP_Loopback_IPv4(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
		errType error
	}{
		{
			name:    "127.0.0.1 loopback",
			ip:      "127.0.0.1",
			wantErr: true,
			errType: ErrLocalhostAccess,
		},
		{
			name:    "127.0.0.0 loopback",
			ip:      "127.0.0.0",
			wantErr: true,
			errType: ErrLocalhostAccess,
		},
		{
			name:    "127.255.255.255 loopback",
			ip:      "127.255.255.255",
			wantErr: true,
			errType: ErrLocalhostAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			assert.NotNil(t, ip, "valid IP should parse")

			err := ValidateIP(ip)

			if tt.wantErr {
				assert.Error(t, err, "expected validation error")
				assert.Equal(t, tt.errType, err, "expected specific error type")
			} else {
				assert.NoError(t, err, "expected no validation error")
			}
		})
	}
}

// TestValidateIP_Loopback tests IPv6 loopback address detection
func TestValidateIP_Loopback_IPv6(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
		errType error
	}{
		{
			name:    "::1 IPv6 loopback",
			ip:      "::1",
			wantErr: true,
			errType: ErrLocalhostAccess,
		},
		{
			name:    "0000:0000:0000:0000:0000:0000:0000:0001",
			ip:      "0000:0000:0000:0000:0000:0000:0000:0001",
			wantErr: true,
			errType: ErrLocalhostAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			assert.NotNil(t, ip, "valid IP should parse")

			err := ValidateIP(ip)

			if tt.wantErr {
				assert.Error(t, err, "expected validation error")
				assert.Equal(t, tt.errType, err, "expected specific error type")
			} else {
				assert.NoError(t, err, "expected no validation error")
			}
		})
	}
}

// TestValidateIP_LinkLocal tests link-local address detection (169.254.0.0/16 and cloud metadata)
func TestValidateIP_LinkLocal_IPv4(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
		errType error
	}{
		{
			name:    "169.254.169.254 AWS metadata",
			ip:      "169.254.169.254",
			wantErr: true,
			errType: ErrLinkLocalAccess,
		},
		{
			name:    "169.254.169.253 AWS metadata GCP fallback",
			ip:      "169.254.169.253",
			wantErr: true,
			errType: ErrLinkLocalAccess,
		},
		{
			name:    "169.254.0.0 link-local start",
			ip:      "169.254.0.0",
			wantErr: true,
			errType: ErrLinkLocalAccess,
		},
		{
			name:    "169.254.169.0",
			ip:      "169.254.169.0",
			wantErr: true,
			errType: ErrLinkLocalAccess,
		},
		{
			name:    "169.254.255.255 link-local end",
			ip:      "169.254.255.255",
			wantErr: true,
			errType: ErrLinkLocalAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			assert.NotNil(t, ip, "valid IP should parse")

			err := ValidateIP(ip)

			if tt.wantErr {
				assert.Error(t, err, "expected validation error")
				assert.Equal(t, tt.errType, err, "expected specific error type")
			} else {
				assert.NoError(t, err, "expected no validation error")
			}
		})
	}
}

// TestValidateIP_LinkLocal tests IPv6 link-local address detection (fe80::/10)
func TestValidateIP_LinkLocal_IPv6(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
		errType error
	}{
		{
			name:    "fe80::1 IPv6 link-local",
			ip:      "fe80::1",
			wantErr: true,
			errType: ErrLinkLocalAccess,
		},
		{
			name:    "fe80::ffff:169.254.169.254",
			ip:      "fe80::ffff:169.254.169.254",
			wantErr: true,
			errType: ErrLinkLocalAccess,
		},
		{
			name:    "fe80::0",
			ip:      "fe80::0",
			wantErr: true,
			errType: ErrLinkLocalAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			assert.NotNil(t, ip, "valid IP should parse")

			err := ValidateIP(ip)

			if tt.wantErr {
				assert.Error(t, err, "expected validation error")
				assert.Equal(t, tt.errType, err, "expected specific error type")
			} else {
				assert.NoError(t, err, "expected no validation error")
			}
		})
	}
}

// TestValidateIP_PrivateRFC1918 tests RFC 1918 private address ranges detection
func TestValidateIP_PrivateRFC1918(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
		errType error
	}{
		{
			name:    "10.0.0.0 private range start",
			ip:      "10.0.0.0",
			wantErr: true,
			errType: ErrPrivateIP,
		},
		{
			name:    "10.255.255.255 private range end",
			ip:      "10.255.255.255",
			wantErr: true,
			errType: ErrPrivateIP,
		},
		{
			name:    "10.1.2.3 private address",
			ip:      "10.1.2.3",
			wantErr: true,
			errType: ErrPrivateIP,
		},
		{
			name:    "172.16.0.0 private range start",
			ip:      "172.16.0.0",
			wantErr: true,
			errType: ErrPrivateIP,
		},
		{
			name:    "172.31.255.255 private range end",
			ip:      "172.31.255.255",
			wantErr: true,
			errType: ErrPrivateIP,
		},
		{
			name:    "172.20.10.5 private address",
			ip:      "172.20.10.5",
			wantErr: true,
			errType: ErrPrivateIP,
		},
		{
			name:    "192.168.0.0 private range start",
			ip:      "192.168.0.0",
			wantErr: true,
			errType: ErrPrivateIP,
		},
		{
			name:    "192.168.255.255 private range end",
			ip:      "192.168.255.255",
			wantErr: true,
			errType: ErrPrivateIP,
		},
		{
			name:    "192.168.1.1 private address",
			ip:      "192.168.1.1",
			wantErr: true,
			errType: ErrPrivateIP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			assert.NotNil(t, ip, "valid IP should parse")

			err := ValidateIP(ip)

			if tt.wantErr {
				assert.Error(t, err, "expected validation error")
				assert.Equal(t, tt.errType, err, "expected specific error type")
			} else {
				assert.NoError(t, err, "expected no validation error")
			}
		})
	}
}

// TestValidateIP_PrivateIPv6 tests IPv6 private address detection
func TestValidateIP_PrivateIPv6(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
		errType error
	}{
		{
			name:    "fd00::1 unique local address",
			ip:      "fd00::1",
			wantErr: true,
			errType: ErrPrivateIP,
		},
		{
			name:    "fc00::1 unique local address",
			ip:      "fc00::1",
			wantErr: true,
			errType: ErrPrivateIP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			assert.NotNil(t, ip, "valid IP should parse")

			err := ValidateIP(ip)

			if tt.wantErr {
				assert.Error(t, err, "expected validation error")
				assert.Equal(t, tt.errType, err, "expected specific error type")
			} else {
				assert.NoError(t, err, "expected no validation error")
			}
		})
	}
}

// TestValidateIP_Unspecified tests unspecified addresses (0.0.0.0 and ::)
func TestValidateIP_Unspecified(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
		errType error
	}{
		{
			name:    "0.0.0.0 IPv4 unspecified",
			ip:      "0.0.0.0",
			wantErr: true,
			errType: ErrPrivateIP,
		},
		{
			name:    "::",
			ip:      "::",
			wantErr: true,
			errType: ErrPrivateIP,
		},
		{
			name:    "0000:0000:0000:0000:0000:0000:0000:0000",
			ip:      "0000:0000:0000:0000:0000:0000:0000:0000",
			wantErr: true,
			errType: ErrPrivateIP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			assert.NotNil(t, ip, "valid IP should parse")

			err := ValidateIP(ip)

			if tt.wantErr {
				assert.Error(t, err, "expected validation error")
				assert.Equal(t, tt.errType, err, "expected specific error type")
			} else {
				assert.NoError(t, err, "expected no validation error")
			}
		})
	}
}

// TestValidateIP_Multicast tests multicast address detection
func TestValidateIP_Multicast_IPv4(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
		errType error
	}{
		{
			name:    "224.0.0.0 multicast start",
			ip:      "224.0.0.0",
			wantErr: true,
			errType: ErrPrivateIP,
		},
		{
			name:    "224.0.0.1 all hosts",
			ip:      "224.0.0.1",
			wantErr: true,
			errType: ErrPrivateIP,
		},
		{
			name:    "239.255.255.255 multicast end",
			ip:      "239.255.255.255",
			wantErr: true,
			errType: ErrPrivateIP,
		},
		{
			name:    "225.0.0.5 multicast address",
			ip:      "225.0.0.5",
			wantErr: true,
			errType: ErrPrivateIP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			assert.NotNil(t, ip, "valid IP should parse")

			err := ValidateIP(ip)

			if tt.wantErr {
				assert.Error(t, err, "expected validation error")
				assert.Equal(t, tt.errType, err, "expected specific error type")
			} else {
				assert.NoError(t, err, "expected no validation error")
			}
		})
	}
}

// TestValidateIP_Multicast tests IPv6 multicast address detection
func TestValidateIP_Multicast_IPv6(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
		errType error
	}{
		{
			name:    "ff00::1 IPv6 multicast",
			ip:      "ff00::1",
			wantErr: true,
			errType: ErrPrivateIP,
		},
		{
			name:    "ff02::1 link-local multicast",
			ip:      "ff02::1",
			wantErr: true,
			errType: ErrPrivateIP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			assert.NotNil(t, ip, "valid IP should parse")

			err := ValidateIP(ip)

			if tt.wantErr {
				assert.Error(t, err, "expected validation error")
				assert.Equal(t, tt.errType, err, "expected specific error type")
			} else {
				assert.NoError(t, err, "expected no validation error")
			}
		})
	}
}

// TestValidateIP_PublicAddresses tests that valid public addresses pass validation
func TestValidateIP_PublicAddresses(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
	}{
		{
			name:    "8.8.8.8 Google DNS",
			ip:      "8.8.8.8",
			wantErr: false,
		},
		{
			name:    "1.1.1.1 Cloudflare DNS",
			ip:      "1.1.1.1",
			wantErr: false,
		},
		{
			name:    "208.67.222.222 OpenDNS",
			ip:      "208.67.222.222",
			wantErr: false,
		},
		{
			name:    "198.41.0.4 A root nameserver",
			ip:      "198.41.0.4",
			wantErr: false,
		},
		{
			name:    "2001:4860:4860::8888 IPv6 public",
			ip:      "2001:4860:4860::8888",
			wantErr: false,
		},
		{
			name:    "2001:4860:4860::8844 IPv6 public",
			ip:      "2001:4860:4860::8844",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			assert.NotNil(t, ip, "valid IP should parse")

			err := ValidateIP(ip)

			if tt.wantErr {
				assert.Error(t, err, "expected validation error")
			} else {
				assert.NoError(t, err, "expected no validation error")
			}
		})
	}
}

// TestValidateIP_EdgeCases tests edge cases and boundary conditions
func TestValidateIP_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
		errType error
	}{
		{
			name:    "9.255.255.255 just before private",
			ip:      "9.255.255.255",
			wantErr: false,
		},
		{
			name:    "11.0.0.0 just after private",
			ip:      "11.0.0.0",
			wantErr: false,
		},
		{
			name:    "172.15.255.255 just before second private range",
			ip:      "172.15.255.255",
			wantErr: false,
		},
		{
			name:    "172.32.0.0 just after second private range",
			ip:      "172.32.0.0",
			wantErr: false,
		},
		{
			name:    "192.167.255.255 just before third private range",
			ip:      "192.167.255.255",
			wantErr: false,
		},
		{
			name:    "192.169.0.0 just after third private range",
			ip:      "192.169.0.0",
			wantErr: false,
		},
		{
			name:    "223.255.255.255 just before multicast",
			ip:      "223.255.255.255",
			wantErr: false,
		},
		{
			name:    "240.0.0.0 reserved range - allowed by Go net package but mentioned in RFC",
			ip:      "240.0.0.0",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			assert.NotNil(t, ip, "valid IP should parse")

			err := ValidateIP(ip)

			if tt.wantErr {
				assert.Error(t, err, "expected validation error")
				if tt.errType != nil {
					assert.Equal(t, tt.errType, err, "expected specific error type")
				}
			} else {
				assert.NoError(t, err, "expected no validation error")
			}
		})
	}
}

// TestValidateIP_AllErrorTypes tests that all error types are properly defined and used
func TestValidateIP_AllErrorTypes(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		errType error
	}{
		{
			name:    "ErrLocalhostAccess",
			ip:      "127.0.0.1",
			errType: ErrLocalhostAccess,
		},
		{
			name:    "ErrLinkLocalAccess",
			ip:      "169.254.169.254",
			errType: ErrLinkLocalAccess,
		},
		{
			name:    "ErrPrivateIP",
			ip:      "192.168.1.1",
			errType: ErrPrivateIP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			err := ValidateIP(ip)
			assert.Error(t, err)
			assert.Equal(t, tt.errType, err)
		})
	}
}

// BenchmarkValidateIP benchmarks IP validation performance
func BenchmarkValidateIP(b *testing.B) {
	// Test with a public IP
	publicIP := net.ParseIP("8.8.8.8")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateIP(publicIP)
	}
}

// BenchmarkValidateIP_Private benchmarks private IP validation
func BenchmarkValidateIP_Private(b *testing.B) {
	privateIP := net.ParseIP("192.168.1.1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateIP(privateIP)
	}
}

// BenchmarkValidateIP_Loopback benchmarks loopback IP validation
func BenchmarkValidateIP_Loopback(b *testing.B) {
	loopbackIP := net.ParseIP("127.0.0.1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateIP(loopbackIP)
	}
}
