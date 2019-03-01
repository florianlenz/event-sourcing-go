package es

type ESEvent struct {
	occurredAt int64
	version    uint8
}

func (e ESEvent) OccurredAt() int64 {
	return e.occurredAt
}

func (e ESEvent) Version() uint8 {
	return e.version
}
