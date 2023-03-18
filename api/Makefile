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


export OPENAI_API_KEY=sk-KltHM7dsS8x2oL0KGJ69XXX
export PG_HOST=192.168.0.1
export PG_DB=hwu
export PG_USER=hwu
export PG_PASS=123456
export PG_PORT=5432
 
serve:
	@echo "Starting server..."
	echo $(OPENAI_API_KEY)
	echo $(PG_HOST)
	air
       


