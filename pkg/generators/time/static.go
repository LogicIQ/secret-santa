package time

import (
	"fmt"
	"time"
)

type StaticGenerator struct{}

func (g *StaticGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	// Use provided RFC3339 timestamp or current time
	rfc3339 := getStringConfig(config, "rfc3339", "")

	var t time.Time
	var err error

	if rfc3339 != "" {
		t, err = time.Parse(time.RFC3339, rfc3339)
		if err != nil {
			return nil, fmt.Errorf("invalid rfc3339 format: %v", err)
		}
	} else {
		t = time.Now().UTC()
	}

	return map[string]string{
		"rfc3339": t.Format(time.RFC3339),
		"unix":    fmt.Sprintf("%d", t.Unix()),
		"year":    fmt.Sprintf("%d", t.Year()),
		"month":   fmt.Sprintf("%d", t.Month()),
		"day":     fmt.Sprintf("%d", t.Day()),
		"hour":    fmt.Sprintf("%d", t.Hour()),
		"minute":  fmt.Sprintf("%d", t.Minute()),
		"second":  fmt.Sprintf("%d", t.Second()),
	}, nil
}
