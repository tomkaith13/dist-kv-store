run-solo:
	cp .env1 .env && go run cmd/main.go
run-duo:
	cp .env2 .env && go run cmd/main.go
run-trio:
	cp .env3 .env && go run cmd/main.go
run-test:
	go test ./...
