package formatters

import (
	"encoding/json"
	"fmt"
)

type JSONFormatter struct {
}

func (jf JSONFormatter) Format(p []byte) ([]byte, error) {
	m := make(map[string]any)

	err := json.Unmarshal(p, &m)
	if err != nil {
		return p, err
	}

	fmt.Println(m)

	return nil, nil
}
