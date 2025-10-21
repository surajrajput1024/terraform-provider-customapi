package types

import (
	"encoding/json"
	"fmt"
)

type CustomAPIRequest struct {
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        json.RawMessage   `json:"body,omitempty"`
	QueryParams map[string]string `json:"query_params,omitempty"`
}

type CustomAPIResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       json.RawMessage   `json:"body"`
	Success    bool              `json:"success"`
	Error      string            `json:"error,omitempty"`
}

type UserProfile struct {
	ID                   int                    `json:"id"`
	Email                string                 `json:"email"`
	Token                *string                `json:"token"`
	Role                 string                 `json:"role"`
	Name                 string                 `json:"name"`
	CreatedAt            string                 `json:"createdAt"`
	UpdatedAt            string                 `json:"updatedAt"`
	UserID               *int                   `json:"userId"`
	RoleID               *int                   `json:"role_id"`
	LastLoginAt          string                 `json:"lastLoginAt"`
	PasswordChangedAt    *string                `json:"passwordChangedAt"`
	MustChangePassword   bool                   `json:"mustChangePassword"`
	FailedLoginAttempts  int                    `json:"failedLoginAttempts"`
	LockedUntil          *string                `json:"lockedUntil"`
	PasswordSecurity     PasswordSecurity       `json:"passwordSecurity"`
}

type PasswordSecurity struct {
	MustChangePassword        bool   `json:"mustChangePassword"`
	PasswordChangeRequired    bool   `json:"passwordChangeRequired"`
	DaysSincePasswordChange   int    `json:"daysSincePasswordChange"`
	PasswordChangedAt         string `json:"passwordChangedAt"`
	IsPasswordExpired         bool   `json:"isPasswordExpired"`
	Reason                    string `json:"reason"`
}

type Organization struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	Status      string `json:"status"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API Error %d: %s", e.Code, e.Message)
}
