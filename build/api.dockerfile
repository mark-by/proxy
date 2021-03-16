# Build
FROM golang:1.16-buster AS build

WORKDIR /app
ADD . .

ENV CGO_ENABLED=0

RUN go mod tidy
RUN go mod vendor
RUN go build -o bin/proxy ./main.go

# Enviroment
FROM alpine:latest

RUN apk upgrade --update-cache --available && \
    apk add openssl && \
    rm -rf /var/cache/apk/*

WORKDIR /app

COPY /root.crt /etc/ssl/certs/ca.crt
RUN apk update
RUN apk add ca-certificates
RUN update-ca-certificates
RUN mkdir certs/

COPY --from=build /app/bin/proxy .
ADD . .

ENTRYPOINT ["/app/proxy"]

EXPOSE 8080
EXPOSE 8888