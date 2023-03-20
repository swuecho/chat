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
	go build
.PHONY: build

 
serve:
	@echo "Starting server..."
	echo $(OPENAI_API_KEY)
	echo $(PG_HOST)
	air
       


