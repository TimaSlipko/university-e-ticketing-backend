package utils

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

// Email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail checks if email format is valid
func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// ValidatePassword checks password strength
func ValidatePassword(password string) (bool, []string) {
	var errors []string

	if len(password) < 8 {
		errors = append(errors, "Password must be at least 8 characters long")
	}

	if len(password) > 72 {
		errors = append(errors, "Password must be less than 72 characters long")
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		errors = append(errors, "Password must contain at least one uppercase letter")
	}
	if !hasLower {
		errors = append(errors, "Password must contain at least one lowercase letter")
	}
	if !hasNumber {
		errors = append(errors, "Password must contain at least one number")
	}
	if !hasSpecial {
		errors = append(errors, "Password must contain at least one special character")
	}

	return len(errors) == 0, errors
}

// ValidateUsername checks username format
func ValidateUsername(username string) (bool, []string) {
	var errors []string

	if len(username) < 3 {
		errors = append(errors, "Username must be at least 3 characters long")
	}

	if len(username) > 30 {
		errors = append(errors, "Username must be less than 30 characters long")
	}

	// Username should contain only alphanumeric characters and underscores
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !usernameRegex.MatchString(username) {
		errors = append(errors, "Username can only contain letters, numbers, and underscores")
	}

	return len(errors) == 0, errors
}

// SanitizeString removes extra whitespace and trims string
func SanitizeString(s string) string {
	// Remove extra whitespace and trim
	s = strings.TrimSpace(s)
	// Replace multiple spaces with single space
	spaceRegex := regexp.MustCompile(`\s+`)
	return spaceRegex.ReplaceAllString(s, " ")
}

// NewError creates a new error - helper function
func NewError(message string) error {
	return errors.New(message)
}

// CalculatePagination calculates pagination values
func CalculatePagination(page, limit int, total int64) Pagination {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
