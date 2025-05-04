package marshalling

import (
	"encoding/json"
	"fmt"
	"time"
)

// Helper function used to find the right time
func MarshalTimeToDateTime(v *time.Time) ([]byte, error) {
	dateStr := (*v).UTC().Format(time.RFC3339)
	return []byte(fmt.Sprintf("\"%s\"", dateStr)), nil
}

func UnmarshalStringToDateTime(src json.RawMessage, dst *time.Time) error {
	dateStr := string(src)
	dateStr = dateStr[1 : len(dateStr)-1]
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		t, err = time.Parse(time.DateOnly, dateStr)
		if err != nil {
			return fmt.Errorf("failed to parse date: %w", err)
		}
	}
	*dst = t
	return nil
}
