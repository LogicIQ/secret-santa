package validation

import (
	"testing"

	secretsantav1alpha1 "github.com/logicIQ/secret-santa/api/v1alpha1"
	"github.com/logicIQ/secret-santa/pkg/generators"
)

func TestValidateTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantErr  bool
	}{
		{
			name:     "valid template",
			template: `{"password": "{{ .gen.value }}"}`,
			wantErr:  false,
		},
		{
			name:     "empty template",
			template: "",
			wantErr:  true,
		},
		{
			name:     "invalid template syntax",
			template: `{"password": "{{ .gen.value }"}`,
			wantErr:  true,
		},
		{
			name:     "template with functions",
			template: `{"password": "{{ .gen.value | default "test" }}"}`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTemplate(tt.template)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateGeneratorConfigs(t *testing.T) {
	generators.Clear()
	generators.Register("test_generator", nil)
	generators.Register("random_password", nil)

	tests := []struct {
		name    string
		configs []secretsantav1alpha1.GeneratorConfig
		wantErr bool
	}{
		{
			name: "valid configs",
			configs: []secretsantav1alpha1.GeneratorConfig{
				{Name: "test", Type: "random_password"},
				{Name: "test2", Type: "test_generator"},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			configs: []secretsantav1alpha1.GeneratorConfig{
				{Name: "", Type: "random_password"},
			},
			wantErr: true,
		},
		{
			name: "empty type",
			configs: []secretsantav1alpha1.GeneratorConfig{
				{Name: "test", Type: ""},
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			configs: []secretsantav1alpha1.GeneratorConfig{
				{Name: "test", Type: "invalid_type"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGeneratorConfigs(tt.configs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGeneratorConfigs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMaskSensitiveData(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{
		{
			name: "json format",
			data: `{"password": "secret123", "key": "value"}`,
			want: "{\n  \"key\": \"\u003cMASKED\u003e\",\n  \"password\": \"\u003cMASKED\u003e\"\n}",
		},
		{
			name: "yaml format",
			data: "password: secret123\nkey: value",
			want: "password: <MASKED>\nkey: <MASKED>",
		},
		{
			name: "text format",
			data: "password=secret123\nkey: value",
			want: "password= <MASKED>\nkey: <MASKED>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskSensitiveData(tt.data)
			if result != tt.want {
				t.Errorf("MaskSensitiveData() = %v, want %v", result, tt.want)
			}
		})
	}
}