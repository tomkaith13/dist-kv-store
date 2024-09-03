run1:
	cp .env1 .env && go run cmd/main.go
run2:
	cp .env2 .env && go run cmd/main.go
run3:
	cp .env3 .env && go run cmd/main.go
run-test:
	go test ./...
