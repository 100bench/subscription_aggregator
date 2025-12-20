package validation

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var dateRegex = regexp.MustCompile(`^(0[1-9]|1[0-2])-\d{4}$`)

// ValidateUUID checks if a string is a valid UUID
func ValidateUUID(userID string) error {
	if userID == "" {
		return fmt.Errorf("user_id is required")
	}
	_, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user_id format: must be a valid UUID")
	}
	return nil
}

// ValidateDate checks if a date string is in MM-YYYY format
func ValidateDate(date, fieldName string) error {
	if date == "" {
		return nil // Optional dates are allowed
	}
	if !dateRegex.MatchString(date) {
		return fmt.Errorf("%s must be in MM-YYYY format", fieldName)
	}
	// Try parsing to ensure it's a valid date
	_, err := time.Parse("01-2006", date)
	if err != nil {
		return fmt.Errorf("%s is not a valid date: %w", fieldName, err)
	}
	return nil
}

// ValidateRequiredDate checks if a required date string is in MM-YYYY format
func ValidateRequiredDate(date, fieldName string) error {
	if date == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return ValidateDate(date, fieldName)
}

// ValidatePrice checks if price is valid (non-negative)
func ValidatePrice(price int) error {
	if price < 0 {
		return fmt.Errorf("price must be non-negative")
	}
	return nil
}

// ValidateServiceName checks if service name is not empty
func ValidateServiceName(serviceName string) error {
	if serviceName == "" {
		return fmt.Errorf("service_name is required")
	}
	return nil
}

// ValidateDateRange checks if end_date is after or equal to start_date
func ValidateDateRange(startDate, endDate string) error {
	if startDate == "" {
		return fmt.Errorf("start_date is required")
	}
	if endDate == "" {
		return nil // end_date is optional
	}

	start, err := time.Parse("01-2006", startDate)
	if err != nil {
		return fmt.Errorf("invalid start_date format")
	}

	end, err := time.Parse("01-2006", endDate)
	if err != nil {
		return fmt.Errorf("invalid end_date format")
	}

	if end.Before(start) {
		return fmt.Errorf("end_date must be after or equal to start_date")
	}

	return nil
}
