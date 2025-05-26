FROM golang:1.24.3@sha256:4c0a1814a7c6c65ece28b3bfea14ee3cf83b5e80b81418453f0e9d5255a5d7b8 AS builder

WORKDIR /go/src
COPY go.mod go.sum .
RUN go mod download

COPY . .
RUN go build -o /go/bin/webserver main.go



FROM bitnami/kubectl:1.32.3@sha256:0ed36a7f6a94c1b82d70811ed03caf3df92b4b2b8f8817d0f726f0f3a26e64af
COPY --from=builder /go/bin/webserver /webserver
COPY static /static
ENTRYPOINT [ "/webserver" ]
