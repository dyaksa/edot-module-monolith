run:
	go run main.go

build:
	go build -o ./bin/main main.go

clean:
	go mod tidy && rm -rf bin/*.

generate-mocks: 
	go generate ./...
clean-mocks:
	rm -rf mocks/*.go

# Test with coverage
test-coverage:
	go test -v -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Test specific package with mocks
test-unit:
	go test -v ./usecase/... ./repository/... -tags=unit

create-migrations:
	goose create -dir ./migrations -s $(name) sql

migrate:
	goose -dir ./migrations up

rollback:
	goose -dir ./migrations down


.PHONY: run build clean test install-mockery generate-mocks clean-mocks regenerate-mocks test-coverage test-unit create-migrations migrate rollback