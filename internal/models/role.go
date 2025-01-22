package models

import (
	"fmt"
	"strings"
)

type Role string

const (
	RoleAdmin                  Role = "admin"
	RoleHealthcareProfessional Role = "healthcare_professional"
	RolePatient                Role = "patient"
	RoleFamilyMember           Role = "family_member"
	RolePrimaryCaregiver       Role = "primary_caregiver"
)

// ParseRole converts a string to a Role type, returning an error if invalid
func ParseRole(value string) (Role, error) {
	// Convert to lowercase for case-insensitive comparison
	normalizedValue := strings.ToLower(value)

	switch Role(normalizedValue) {
	case RoleAdmin, RoleHealthcareProfessional, RolePatient,
		RoleFamilyMember, RolePrimaryCaregiver:
		return Role(normalizedValue), nil
	default:
		return "", fmt.Errorf("invalid role: %s. Valid roles are: %s",
			value, strings.Join(ValidRoles(), ", "))
	}
}

// ValidRoles returns all valid roles as a slice of strings
func ValidRoles() []string {
	return []string{
		string(RoleAdmin),
		string(RoleHealthcareProfessional),
		string(RolePatient),
		string(RoleFamilyMember),
		string(RolePrimaryCaregiver),
	}
}
