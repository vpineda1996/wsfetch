package marshalling

import (
	"fmt"
	"time"
)

// Helper function used to find the right time
func MarshalTimeToDateTime(v *time.Time) ([]byte, error) {
	dateStr := (*v).UTC().Format(time.RFC3339)
	return []byte(fmt.Sprintf("\"%s\"", dateStr)), nil
}
