FROM alpine:latest as builder

ARG GO_BINARY_NAME="go1.22.0.linux-amd64.tar.gz"

WORKDIR /build

RUN apk update && apk add wget
RUN wget "https://go.dev/dl/$GO_BINARY_NAME" && tar -C /usr/local -xzf $GO_BINARY_NAME

COPY go.mod .
COPY go.sum .
RUN /usr/local/go/bin/go mod download

COPY main.go .
COPY stepin.go .
RUN /usr/local/go/bin/go build -o stepin .

FROM alpine:latest as base

WORKDIR /app

COPY --from=builder /build/stepin .
COPY templates templates

RUN apk update && apk add step-cli

CMD [ "/app/stepin" ]
