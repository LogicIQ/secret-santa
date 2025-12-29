package validation

import (
	"fmt"
	"encoding/json"
	"regexp"
	"strings"
	"text/template"

	secretsantav1alpha1 "github.com/logicIQ/secret-santa/api/v1alpha1"
	"github.com/logicIQ/secret-santa/pkg/generators"
	tmplpkg "github.com/logicIQ/secret-santa/pkg/template"
)

func ValidateTemplate(tmplStr string) error {
	if tmplStr == "" {
		return fmt.Errorf("template cannot be empty")
	}

	_, err := template.New("validation").Funcs(tmplpkg.FuncMap()).Parse(tmplStr)
	return err
}

func ValidateGeneratorConfigs(configs []secretsantav1alpha1.GeneratorConfig) error {
	for _, config := range configs {
		if config.Name == "" {
			return fmt.Errorf("generator name cannot be empty")
		}
		if config.Type == "" {
			return fmt.Errorf("generator type cannot be empty for generator '%s'", config.Name)
		}
		
		if !generators.IsSupported(config.Type) {
			return fmt.Errorf("unsupported generator type '%s' for generator '%s'", config.Type, config.Name)
		}
	}

	return nil
}

func MaskSensitiveData(data string) string {
	if strings.HasPrefix(strings.TrimSpace(data), "{") {
		return maskJSONData(data)
	}
	
	if strings.Contains(data, ":") && !strings.Contains(data, "{") {
		return maskYAMLData(data)
	}
	
	return maskGenericData(data)
}

func maskJSONData(data string) string {
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(data), &obj); err != nil {
		return maskGenericData(data)
	}
	
	masked := maskMapValues(obj)
	result, err := json.MarshalIndent(masked, "", "  ")
	if err != nil {
		return maskGenericData(data)
	}
	
	return string(result)
}

func maskYAMLData(data string) string {
	lines := strings.Split(data, "\n")
	for i, line := range lines {
		if strings.Contains(line, ":") && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
				indent := strings.Repeat(" ", len(line)-len(strings.TrimLeft(line, " ")))
				lines[i] = fmt.Sprintf("%s%s: <MASKED>", indent, strings.TrimSpace(parts[0]))
			}
		}
	}
	return strings.Join(lines, "\n")
}

func maskGenericData(data string) string {
	patterns := []string{
		`(\w+)\s*[:=]\s*([^\s\n]+)`,
		`"([^"]{8,})"`,
		`'([^']{8,})'`,
	}
	
	result := data
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			if strings.Contains(match, ":") || strings.Contains(match, "=") {
				parts := regexp.MustCompile(`[:=]`).Split(match, 2)
				if len(parts) == 2 {
					separator := ":"
					if strings.Contains(match, "=") {
						separator = "="
					}
					return parts[0] + separator + " <MASKED>"
				}
			}
			return "<MASKED>"
		})
	}
	
	return result
}

func maskMapValues(obj map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range obj {
		switch v := value.(type) {
		case map[string]interface{}:
			result[key] = maskMapValues(v)
		case []interface{}:
			result[key] = maskSliceValues(v)
		case string:
			if len(v) > 0 {
				result[key] = "<MASKED>"
			} else {
				result[key] = v
			}
		default:
			result[key] = "<MASKED>"
		}
	}
	return result
}

func maskSliceValues(slice []interface{}) []interface{} {
	result := make([]interface{}, len(slice))
	for i, item := range slice {
		switch v := item.(type) {
		case map[string]interface{}:
			result[i] = maskMapValues(v)
		case []interface{}:
			result[i] = maskSliceValues(v)
		case string:
			if len(v) > 0 {
				result[i] = "<MASKED>"
			} else {
				result[i] = v
			}
		default:
			result[i] = "<MASKED>"
		}
	}
	return result
}

