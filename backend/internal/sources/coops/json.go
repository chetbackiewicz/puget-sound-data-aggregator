package coops

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func jsonUnmarshal(body []byte, v any) error {
	return json.Unmarshal(body, v)
}

func stringify(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	default:
		return fmt.Sprint(t)
	}
}
