FROM golang:1.24-alpine

WORKDIR /app

# Install air for hot-reload (pinned to last version supporting Go 1.24)
RUN go install github.com/air-verse/air@v1.61.7

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

# Use air for hot-reload in development
CMD ["air", "-c", ".air.toml"]
