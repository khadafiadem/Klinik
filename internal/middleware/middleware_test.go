package middleware

import (
	"strings"
	"testing"
)

func TestRoleMatching(t *testing.T) {
	userRoles := []string{"ADMIN", "DOCTOR"}
	requiredRoles := []string{"ADMIN"}

	hasRole := false
	for _, required := range requiredRoles {
		for _, userRole := range userRoles {
			if required == userRole {
				hasRole = true
				break
			}
		}
		if hasRole {
			break
		}
	}

	if !hasRole {
		t.Error("expected to have ADMIN role")
	}
}

func TestRoleMatchingCaseInsensitive(t *testing.T) {
	userRoles := []string{"admin"}
	requiredRoles := []string{"ADMIN"}

	hasRole := false
	for _, required := range requiredRoles {
		for _, userRole := range userRoles {
			if strings.EqualFold(required, userRole) {
				hasRole = true
				break
			}
		}
		if hasRole {
			break
		}
	}

	if !hasRole {
		t.Error("expected case-insensitive role matching to work")
	}
}

func TestPermissionMatching(t *testing.T) {
	userPermissions := []string{"patients.read", "patients.write"}
	requiredPermissions := []string{"patients.read"}

	hasPermission := false
	for _, required := range requiredPermissions {
		for _, userPerm := range userPermissions {
			if required == userPerm {
				hasPermission = true
				break
			}
		}
		if hasPermission {
			break
		}
	}

	if !hasPermission {
		t.Error("expected to have patients.read permission")
	}
}

func TestNoRoleMatch(t *testing.T) {
	userRoles := []string{"DOCTOR"}
	requiredRoles := []string{"ADMIN", "CASHIER"}

	hasRole := false
	for _, required := range requiredRoles {
		for _, userRole := range userRoles {
			if required == userRole {
				hasRole = true
				break
			}
		}
		if hasRole {
			break
		}
	}

	if hasRole {
		t.Error("expected no role match")
	}
}

func TestNoPermissionMatch(t *testing.T) {
	userPermissions := []string{"patients.read"}
	requiredPermissions := []string{"payments.write"}

	hasPermission := false
	for _, required := range requiredPermissions {
		for _, userPerm := range userPermissions {
			if required == userPerm {
				hasPermission = true
				break
			}
		}
		if hasPermission {
			break
		}
	}

	if hasPermission {
		t.Error("expected no permission match")
	}
}
