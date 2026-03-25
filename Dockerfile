FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /bin/marketing ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/marketing /bin/marketing
COPY migrations /migrations
EXPOSE 8083
CMD ["/bin/marketing"]
