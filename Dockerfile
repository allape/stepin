FROM node:22 AS ui_builder

WORKDIR /build

COPY ui/package.json        .
COPY ui/package-lock.json   .

RUN npm i --no-audit

COPY ui .

RUN npm run build

FROM golang:1.23.3-alpine3.20 AS builder

WORKDIR /build

COPY go.mod go.mod
COPY go.sum go.sum
RUN /usr/local/go/bin/go mod download

RUN apk update && apk add build-base

COPY . .
RUN /usr/local/go/bin/go build -o app

FROM alpine:3.20

WORKDIR /app

RUN apk update && apk add step-cli

COPY --from=ui_builder /build/dist ui/dist
COPY --from=builder /build/app app

EXPOSE 8080

CMD [ "/app/app" ]

### build ###
# export docker_http_proxy=http://host.docker.internal:1080
# docker build --platform linux/amd64 --build-arg http_proxy=$docker_http_proxy --build-arg https_proxy=$docker_http_proxy -f Dockerfile -t allape/stepin:latest .

### run ###
# docker compose -f docker.compose.yaml up -d

