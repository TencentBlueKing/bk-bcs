package gokong

import (
	"encoding/json"
	"fmt"
)

type Id string

func (c *Id) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte{}, nil
	}

	return []byte(fmt.Sprintf(`{"id":"%s"}`, *c)), nil
}

func (c *Id) UnmarshalJSON(data []byte) error {

	m := map[string]string{}
	err := json.Unmarshal(data, &m)
	if err != nil || m == nil {
		return nil
	}

	if val, ok := m["id"]; ok {
		id := Id(val)
		*c = id
		return nil
	}

	return nil
}
