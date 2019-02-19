package es

type testLogger struct {
	errorChan chan error
}

func (l *testLogger) Error(error error) {
	l.errorChan <- error
}
