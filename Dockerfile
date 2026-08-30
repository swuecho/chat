FROM node:18-alpine AS frontend_builder
WORKDIR /app
COPY web/package*.json ./
RUN npm ci
COPY web/ .
RUN npm run build

FROM golang:1.24-alpine3.20 AS builder
WORKDIR /app
COPY api/go.mod api/go.sum ./
RUN go mod download
COPY api/ .
COPY --from=frontend_builder /app/dist/ ./static/
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -installsuffix cgo -o /app/app

FROM alpine:3.20
RUN adduser -D -g '' appuser
WORKDIR /app
COPY --from=builder /app/app /app
COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /app/zoneinfo.zip
ENV ZONEINFO=/app/zoneinfo.zip
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1
USER appuser
ENTRYPOINT ["/app/app"]
