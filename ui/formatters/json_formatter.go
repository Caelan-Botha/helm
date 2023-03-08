package formatters

import (
	"encoding/json"
	"fmt"
)

type JSONTestFormatter struct{}

func (jf JSONTestFormatter) Format(p []byte) ([]byte, error) {
	m := make(map[string]any)

	err := json.Unmarshal(p, &m)
	if err != nil {
		return p, err
	}

	out := make([]byte, 0)

	for fieldName, fieldValue := range m {
		fv := fmt.Sprintf("%v", fieldValue)
		str := fieldName + " -> " + fv + "!\n"
		out = append(out, []byte(str)...)
	}

	return out, nil
}
