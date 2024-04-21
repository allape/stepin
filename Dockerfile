FROM node:21.7.3 AS ui_builder

RUN npm config --global set proxy "$http_proxy" && \
    npm config --global set https-proxy "$http_proxy"

WORKDIR /build

COPY ui/package.json        .
COPY ui/package-lock.json   .

RUN npm i

COPY ui .

RUN npm run build

FROM alpine:3.19.1 as builder

ARG GO_BINARY_NAME="go1.22.2.linux-amd64.tar.gz"

WORKDIR /build

RUN apk update && apk add wget curl
RUN wget "https://go.dev/dl/$GO_BINARY_NAME" && tar -C /usr/local -xzf $GO_BINARY_NAME

COPY go.mod go.mod
COPY go.sum go.sum
RUN /usr/local/go/bin/go mod download

# GCC
RUN apk update && apk add build-base

COPY . .
RUN /usr/local/go/bin/go build -o app

FROM alpine:3.19.1

WORKDIR /app

COPY --from=ui_builder /build/dist assets
COPY --from=builder /build/app app

EXPOSE 8080

CMD [ "/app/app" ]



