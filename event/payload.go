package event

import "fmt"

// event payload
type Payload map[string]interface{}

// get string from payload
func (sep Payload) GetString(key string) (string, error) {

	str, k := sep[key].(string)
	if !k {
		return "", fmt.Errorf("'%s' of event payload is not a string", str)
	}

	return str, nil

}
