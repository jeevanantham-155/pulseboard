FROM golang:1.27-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /live-polling ./

FROM alpine:3.22

RUN adduser -D -H appuser && mkdir -p /app && chown -R appuser:appuser /app
WORKDIR /app
USER appuser
COPY --from=builder /live-polling /live-polling
EXPOSE 8080
ENTRYPOINT ["/live-polling"]
