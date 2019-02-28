package es

type ESEvent struct {
	occurredAt int64
}

func (e ESEvent) OccurredAt() int64 {
	return e.occurredAt
}
