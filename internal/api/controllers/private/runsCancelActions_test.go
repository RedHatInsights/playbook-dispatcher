package private

import (
	"errors"
	"net/http"
	"testing"

	"playbook-dispatcher/internal/api/dispatch"
)

func TestHandleRunCancelError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "RunNotFoundError returns 404",
			err:      &dispatch.RunNotFoundError{},
			expected: http.StatusNotFound,
		},
		{
			name:     "RunOrgIdMismatchError returns 400",
			err:      &dispatch.RunOrgIdMismatchError{},
			expected: http.StatusBadRequest,
		},
		{
			name:     "RecipientNotFoundError returns 409",
			err:      &dispatch.RecipientNotFoundError{},
			expected: http.StatusConflict,
		},
		{
			name:     "RunCancelNotCancelableError returns 409",
			err:      &dispatch.RunCancelNotCancelableError{},
			expected: http.StatusConflict,
		},
		{
			name:     "RunCancelTypeError returns 400",
			err:      &dispatch.RunCancelTypeError{},
			expected: http.StatusBadRequest,
		},
		{
			name:     "Unknown error returns 500",
			err:      errors.New("some other error"),
			expected: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleRunCancelError(tt.err)
			if result.Code != tt.expected {
				t.Errorf("handleRunCancelError(%T) = %d, want %d", tt.err, result.Code, tt.expected)
			}
		})
	}
}
