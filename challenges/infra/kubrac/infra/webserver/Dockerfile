FROM golang:1.24.2@sha256:18a1f2d1e1d3c49f27c904e9182375169615c65852ace724987929b910195b2c AS builder

WORKDIR /go/src
COPY go.mod go.sum .
RUN go mod download

COPY . .
RUN go build -o /go/bin/webserver main.go



FROM bitnami/kubectl:1.32.3@sha256:0ed36a7f6a94c1b82d70811ed03caf3df92b4b2b8f8817d0f726f0f3a26e64af
COPY --from=builder /go/bin/webserver /webserver
COPY static /static
ENTRYPOINT [ "/webserver" ]
