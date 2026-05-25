FROM golang:1.26-alpine3.22 AS goose-builder

RUN apk add --no-cache git

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

WORKDIR /app/migrations
COPY migrations/ ./

FROM alpine:3.22

COPY --from=goose-builder /go/bin/goose /goose
COPY --from=goose-builder /app/migrations /migrations

WORKDIR /migrations

ENTRYPOINT ["/bin/sh", "-c", "exec /goose -dir /migrations -table migrations postgres \"$POSTGRES_DSN\" up"]
