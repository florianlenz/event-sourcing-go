package event

type Marshal interface {
	// marshal
	Unmarshal(params map[string]interface{}) error
	// unmarshal
	Marshal() (map[string]interface{}, error)
}
