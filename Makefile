.PHONY: test coverage

# Run the app
run:
	make generate-api-doc && go run cmd/server/main.go

test:
	go test ./... -v

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

coverage-html:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "HTML coverage report generated: ./coverage.html"


generate-mocks:
	go generate ./...

generate-api-doc:
	swag init -g cmd/server/main.go