package validation

import (
	"testing"
)

func TestValidateTemplate(t *testing.T) {
	tests := []struct {
		name      string
		template  string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "empty template",
			template:  "",
			wantError: true,
			errorMsg:  "template cannot be empty",
		},
		{
			name:      "valid template",
			template:  `password: {{ .pass.value }}`,
			wantError: false,
		},
		{
			name:      "direct root context access",
			template:  `{{.}}`,
			wantError: true,
			errorMsg:  "direct root context access is not allowed",
		},
		{
			name:      "range over root context",
			template:  `{{ range . }}{{ end }}`,
			wantError: true,
			errorMsg:  "ranging over root context is not allowed",
		},
		{
			name:      "with root context",
			template:  `{{ with . }}{{ end }}`,
			wantError: true,
			errorMsg:  "with root context is not allowed",
		},
		{
			name:      "call function",
			template:  `{{ call .func }}`,
			wantError: true,
			errorMsg:  "call function is not allowed",
		},
		{
			name:      "js function",
			template:  `{{ js .value }}`,
			wantError: true,
			errorMsg:  "js function is not allowed",
		},
		{
			name:      "urlquery function",
			template:  `{{ urlquery .value }}`,
			wantError: true,
			errorMsg:  "urlquery function is not allowed",
		},
		{
			name:      "valid nested access",
			template:  `{{ .user.password }}`,
			wantError: false,
		},
		{
			name:      "valid with nested context",
			template:  `{{ with .user }}{{ .name }}{{ end }}`,
			wantError: false,
		},
		{
			name:      "valid range over nested",
			template:  `{{ range .items }}{{ .name }}{{ end }}`,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTemplate(tt.template)
			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateTemplate() expected error but got none")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("ValidateTemplate() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateTemplate() unexpected error = %v", err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
