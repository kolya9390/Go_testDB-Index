FROM golang:1.22 as builder

WORKDIR /test_garantex

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /test_garantex/main ./cmd/app

FROM alpine:latest

RUN apk --no-cache add libc6-compat

COPY --from=builder /test_garantex/main /main

WORKDIR /test_garantex

CMD ["/main"]