.DEFAULT_GOAL:=build

fmt:
	go fmt ./...

.PHONY: fmt

lint: fmt
	golint ./...

.PHONY: lint

vet: fmt
	go vet ./...

.PHONY: vet

build: vet
	go build -v ./...
.PHONY: build

test:
	go test -v ./...
.PHONY: test

install:
	go install github.com/air-verse/air@latest
	go mod tidy
 
serve:
	@echo "Starting server..."
	echo 'sudo lsof -i -P -n | grep 8080'
	echo $(OPENAI_API_KEY)
	echo $(PG_HOST)
	air
       


