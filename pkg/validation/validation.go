package validation

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	secretsantav1alpha1 "github.com/logicIQ/secret-santa/api/v1alpha1"
	"github.com/logicIQ/secret-santa/pkg/generators"
	tmplpkg "github.com/logicIQ/secret-santa/pkg/template"
)

var (
	genericMaskPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(\w+)\s*[:=]\s*([^\s\n]+)`),
		regexp.MustCompile(`"([^"]{8,})"`),
		regexp.MustCompile(`'([^']{8,})'`),
	}
	separatorPattern = regexp.MustCompile(`[:=]`)
	yamlPattern      = regexp.MustCompile(`^\s*[a-zA-Z_][a-zA-Z0-9_-]*\s*:\s*.+$`)

	dangerousPatterns = []*regexp.Regexp{
		regexp.MustCompile(`{{\.}}`),
		regexp.MustCompile(`{{\s*range\s+\.\s*}}`),
		regexp.MustCompile(`{{\s*with\s+\.\s*}}`),
		regexp.MustCompile(`{{.*call.*}}`),
		regexp.MustCompile(`{{.*js.*}}`),
		regexp.MustCompile(`{{.*urlquery.*}}`),
	}
	dangerousPatternMessages = []string{
		"direct root context access is not allowed",
		"ranging over root context is not allowed",
		"with root context is not allowed",
		"call function is not allowed",
		"js function is not allowed",
		"urlquery function is not allowed",
	}
)

func ValidateTemplate(tmplStr string) error {
	if tmplStr == "" {
		return fmt.Errorf("template cannot be empty")
	}

	// Check for dangerous patterns that could lead to code injection
	for i, pattern := range dangerousPatterns {
		if pattern.MatchString(tmplStr) {
			return fmt.Errorf("template validation failed: %s", dangerousPatternMessages[i])
		}
	}

	// Validate template syntax with restricted function map
	// Create a new template with only safe functions to prevent code injection
	safeFuncs := template.FuncMap{}
	for name, fn := range tmplpkg.FuncMap() {
		// Only allow explicitly safe functions
		safeFuncs[name] = fn
	}
	// Remove potentially dangerous functions
	delete(safeFuncs, "call")
	delete(safeFuncs, "js")
	delete(safeFuncs, "urlquery")

	tmpl := template.New("validation").Option("missingkey=error").Funcs(safeFuncs)
	_, err := tmpl.Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("template syntax error: %w", err)
	}

	return nil
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

	if isYAMLFormat(data) {
		return maskYAMLData(data)
	}

	return maskGenericData(data)
}

func isYAMLFormat(data string) bool {
	lines := strings.Split(data, "\n")
	yamlLineCount := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if yamlPattern.MatchString(line) {
			yamlLineCount++
		}
	}
	return yamlLineCount > 0
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
				trimmed := strings.TrimLeft(line, " \t")
				indent := line[:len(line)-len(trimmed)]
				lines[i] = fmt.Sprintf("%s%s: <MASKED>", indent, strings.TrimSpace(parts[0]))
			}
		}
	}
	return strings.Join(lines, "\n")
}

func maskGenericData(data string) string {
	result := data
	for _, re := range genericMaskPatterns {
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			if strings.Contains(match, ":") || strings.Contains(match, "=") {
				parts := separatorPattern.Split(match, 2)
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
		default:
			result[i] = "<MASKED>"
		}
	}
	return result
}
