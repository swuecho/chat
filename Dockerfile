FROM node:16 as frontend_builder

# Set the working directory to /app
WORKDIR /app

# Copy the package.json and package-lock.json files to the container
COPY web/package*.json ./

# Install dependencies
RUN npm install

# Copy the remaining application files to the container
COPY web/ .
# Build the application
RUN npm run build

FROM golang:1.19-alpine3.16 AS builder

WORKDIR /app

COPY api/go.mod api/go.sum ./
RUN go mod download

COPY api/ .
# web/dist/* api/static/
COPY --from=frontend_builder /app/dist/ ./static/
COPY --from=frontend_builder /app/dist/assets ./static/

RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -installsuffix cgo -o /app/app

FROM alpine:3.16

WORKDIR /app

COPY --from=builder /app/app /app
# for go timezone work
COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /app/zoneinfo.zip
ENV ZONEINFO=/app/zoneinfo.zip 

EXPOSE 8080

ENTRYPOINT ["/app/app"]