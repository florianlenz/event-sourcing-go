package es

import (
	"fmt"
)

// event payload
type Payload map[string]interface{}

// get string from payload
func (p Payload) GetString(key string) (string, error) {

	// exit if key doesn't exist
	if _, exists := p[key]; !exists {
		return "", fmt.Errorf("'%s' does not exist in payload", key)
	}

	// exit if key is not a string
	str, k := p[key].(string)
	if !k {
		return "", fmt.Errorf("value of key '%s' is not a string", key)
	}

	return str, nil

}
