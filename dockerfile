# Build stage
FROM golang:1.22.1-alpine3.18 as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

FROM golang:1.22.1-alpine3.18 as builder
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -tags migrate -o /bin/app ./cmd/app


FROM alpine:3.18

RUN apk --no-cache add ca-certificates

COPY --from=builder /bin/app /app
COPY --from=builder /app/config /config
CMD ["/app"]
