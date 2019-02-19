test:
	go test ./...
testc:
	go test ./... -coverprofile=c.out && go tool cover -html=c.out