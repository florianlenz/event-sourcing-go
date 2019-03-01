package event

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

func NewESEvent(occurredAt int64, version uint8) ESEvent {
	return ESEvent{
		occurredAt: occurredAt,
		version:    version,
	}
}
