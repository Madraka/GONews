package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"fmt"
	"strings"
	"time"
)

// TOTPManager handles Two-Factor Authentication using Time-based One-Time Passwords
type TOTPManager struct {
	secretSize int
	digits     int
	period     uint64
	algorithm  string
}

// NewTOTPManager creates a new TOTP manager with default settings
func NewTOTPManager() *TOTPManager {
	return &TOTPManager{
		secretSize: 20,
		digits:     6,
		period:     30,
		algorithm:  "SHA1",
	}
}

// GenerateSecret creates a new random secret for TOTP setup
func (tm *TOTPManager) GenerateSecret() string {
	// Random byte array
	secret := make([]byte, tm.secretSize)
	_, err := rand.Read(secret)
	if err != nil {
		// Fallback to current time based generation if crypto/rand fails
		return fmt.Sprintf("FALLBACK%d", time.Now().UnixNano())
	}

	// Convert to base32 (as per RFC 4648)
	encoder := base32.StdEncoding.WithPadding(base32.NoPadding)
	secretBase32 := encoder.EncodeToString(secret)

	// Add spaces every 4 characters for better readability
	var result strings.Builder
	for i, char := range secretBase32 {
		if i > 0 && i%4 == 0 {
			result.WriteRune(' ')
		}
		result.WriteRune(char)
	}

	return result.String()
}

// GenerateTOTP generates a Time-based One-Time Password
func (tm *TOTPManager) GenerateTOTP(secret string, timestamp time.Time) (string, error) {
	// Remove spaces from secret
	secret = strings.ReplaceAll(secret, " ", "")

	// Decode secret from base32
	decoder := base32.StdEncoding.WithPadding(base32.NoPadding)
	secretBytes, err := decoder.DecodeString(secret)
	if err != nil {
		return "", err
	}

	// Calculate counter (number of time periods since Unix epoch)
	counter := uint64(timestamp.Unix()) / tm.period

	// Generate HOTP value
	otpValue := tm.generateHOTP(secretBytes, counter)

	// Format as string with leading zeros if needed
	otp := fmt.Sprintf("%0*d", tm.digits, otpValue)

	// Truncate to the required number of digits
	if len(otp) > tm.digits {
		otp = otp[len(otp)-tm.digits:]
	}

	return otp, nil
}

// ValidateTOTP validates a Time-based One-Time Password
func (tm *TOTPManager) ValidateTOTP(secret, otp string, timestamp time.Time) bool {
	// Allow for clock skew by checking adjacent time periods
	for i := -1; i <= 1; i++ {
		adjustedTime := timestamp.Add(time.Duration(i) * time.Second * time.Duration(tm.period))
		generatedOTP, err := tm.GenerateTOTP(secret, adjustedTime)
		if err == nil && generatedOTP == otp {
			return true
		}
	}

	return false
}

// generateHOTP generates an HMAC-based One-Time Password
func (tm *TOTPManager) generateHOTP(secret []byte, counter uint64) int {
	// Convert counter to byte array (big-endian)
	counterBytes := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		counterBytes[i] = byte(counter & 0xff)
		counter >>= 8
	}

	// Calculate HMAC-SHA1
	h := hmac.New(sha1.New, secret)
	h.Write(counterBytes)
	hash := h.Sum(nil)

	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0f
	binary := ((int(hash[offset]) & 0x7f) << 24) |
		((int(hash[offset+1]) & 0xff) << 16) |
		((int(hash[offset+2]) & 0xff) << 8) |
		(int(hash[offset+3]) & 0xff)

	// Truncate to the required number of digits
	return binary % pow10(tm.digits)
}

// pow10 calculates 10^n
func pow10(n int) int {
	result := 1
	for i := 0; i < n; i++ {
		result *= 10
	}
	return result
}

// GetQRCodeURL returns a URL for generating QR codes for TOTP setup
func (tm *TOTPManager) GetQRCodeURL(secret, accountName, issuer string) string {
	cleanSecret := strings.ReplaceAll(secret, " ", "")
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=%s&digits=%d&period=%d",
		issuer,
		accountName,
		cleanSecret,
		issuer,
		tm.algorithm,
		tm.digits,
		tm.period)
}
