test:
	- GIN_MODE=release go test -v -coverprofile=coverage.out ./handlers
	- go tool cover -func coverage.out
	- go tool cover -html=coverage.out
