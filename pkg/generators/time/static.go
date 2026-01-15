package time

import (
	"fmt"
	"strconv"
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
		"unix":    strconv.FormatInt(t.Unix(), 10),
		"year":    strconv.Itoa(t.Year()),
		"month":   strconv.Itoa(int(t.Month())),
		"day":     strconv.Itoa(t.Day()),
		"hour":    strconv.Itoa(t.Hour()),
		"minute":  strconv.Itoa(t.Minute()),
		"second":  strconv.Itoa(t.Second()),
	}, nil
}
