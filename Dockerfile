FROM alpine:latest as builder

ARG GO_BINARY_NAME="go1.22.0.linux-amd64.tar.gz"

WORKDIR /build

RUN apk update && apk add wget
RUN wget "https://go.dev/dl/$GO_BINARY_NAME" && tar -C /usr/local -xzf $GO_BINARY_NAME

COPY go.mod .
COPY go.sum .
RUN /usr/local/go/bin/go mod download

COPY main.go .
COPY stepin stepin
RUN /usr/local/go/bin/go build -o stepin .

FROM alpine:latest as base

WORKDIR /app

RUN apk update && apk add step-cli

COPY templates templates
COPY --from=builder /build/stepin .

CMD [ "/app/stepin" ]
