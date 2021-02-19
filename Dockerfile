FROM golang:alpine AS builder


ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o lango_quick .
WORKDIR /dist
RUN cp /build/.env .env
RUN cp /build/lango_quick .

FROM alpine

COPY --from=builder /dist/lango_quick /
ENTRYPOINT ["/lango_quick"]
