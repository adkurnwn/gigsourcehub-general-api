run: 
	@go run main.go

test:
	@go test -v -cover -covermode=atomic ./...

testcoverage:
	go test -v -cover -covermode=atomic ./... --coverprofile=coverage.out ./... && \
	go tool cover -func=coverage.out

generate-mocks:
	@rm -rf mocks && \
	mockery --all

build: 
	@go build -ldflags="-s -w" -o build-app main.go

run-swagger:
	@swag fmt && swag init ./main.go -o ./docs

migration:
	atlas migrate diff ${name} --env hcl -c file://database/atlas.hcl

migrate-up:
	atlas migrate apply --env hcl -c file://database/atlas.hcl

migrate-down:
	atlas migrate down --env hcl -c file://database/atlas.hcl