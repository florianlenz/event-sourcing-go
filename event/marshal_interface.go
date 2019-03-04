package event

type Marshal interface {
	// marshal
	Unmarshal(content interface{}) error
	// unmarshal
	Marshal() (interface{}, error)
}
