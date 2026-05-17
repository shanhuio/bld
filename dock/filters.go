package dock

import (
	"encoding/json"
	"fmt"
)

func labelFilters(label string) (string, error) {
	filters := map[string][]string{"label": {label}}
	filterBytes, err := json.Marshal(filters)
	if err != nil {
		return "", fmt.Errorf("marshal filter: %w", err)
	}
	return string(filterBytes), nil
}
