package validators

import (
	"errors"
	"regexp"
	"strconv"
	"unicode"
)

// PasswordValidator contains methods to validate password strength
type PasswordValidator struct {
	MinLength          int
	RequireUppercase   bool
	RequireLowercase   bool
	RequireDigit       bool
	RequireSpecial     bool
	DisallowCommon     bool
	commonPasswords    map[string]bool
	specialCharPattern *regexp.Regexp
}

// NewPasswordValidator creates a new password validator with default settings
func NewPasswordValidator() *PasswordValidator {
	return &PasswordValidator{
		MinLength:        8,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
		RequireSpecial:   true,
		DisallowCommon:   true,
		// Initialize with a small set of commonly used passwords
		commonPasswords: map[string]bool{
			"password": true, "123456": true, "qwerty": true,
			"admin": true, "welcome": true, "abc123": true,
		},
		specialCharPattern: regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`),
	}
}

// Validate checks a password against the defined rules
func (pv *PasswordValidator) Validate(password string) error {
	if len(password) < pv.MinLength {
		return errors.New("password must be at least " + strconv.Itoa(pv.MinLength) + " characters long")
	}

	if pv.DisallowCommon {
		if pv.commonPasswords[password] {
			return errors.New("password is too common; please choose a more secure password")
		}
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case pv.specialCharPattern.MatchString(string(char)):
			hasSpecial = true
		}
	}

	if pv.RequireUppercase && !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	if pv.RequireLowercase && !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}

	if pv.RequireDigit && !hasDigit {
		return errors.New("password must contain at least one digit")
	}

	if pv.RequireSpecial && !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}
