FROM golang:1.24.3-alpine

WORKDIR /go/src
COPY go.mod go.sum .
RUN go mod download

COPY main.go .
ARG KEY
RUN go build -o /go/bin/main -ldflags="-X 'main.Key=${KEY}'" main.go

ENTRYPOINT [ "/go/bin/main" ]
