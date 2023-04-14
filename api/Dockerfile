FROM golang:1.19-alpine3.16 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -installsuffix cgo -o /app/app

FROM alpine:3.16

WORKDIR /app

COPY --from=builder /app/app /app
# for go timezone work
COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /app/zoneinfo.zip
ENV ZONEINFO=/app/zoneinfo.zip 

EXPOSE 8077

ENTRYPOINT ["/app/app"]