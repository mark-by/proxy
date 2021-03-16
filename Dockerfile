# Build
FROM golang:1.16-buster AS build

WORKDIR /app
COPY . .

ENV CGO_ENABLED=0

RUN go mod tidy
RUN go mod vendor
RUN go build -o bin/proxy ./main.go

# Enviroment
FROM alpine:latest

RUN apk upgrade --update-cache --available && \
    apk add openssl && \
    rm -rf /var/cache/apk/*

RUN apk update
RUN apk add ca-certificates

WORKDIR /app

RUN chmod +x /certGen.bash
RUN ./certGen.bash

FROM postgres:alpine

USER postgres

RUN chmod 0700 /var/lib/postgresql/data &&\
    initdb /var/lib/postgresql/data &&\
    echo "host all  all    0.0.0.0/0  md5" >> /var/lib/postgresql/data/pg_hba.conf &&\
    echo "listen_addresses='*'" >> /var/lib/postgresql/data/postgresql.conf &&\
    pg_ctl start &&\
    psql --command "CREATE USER proxy WITH SUPERUSER PASSWORD '123';" &&\
    createdb -O proxy proxy &&\
    pg_ctl stop

COPY --from=build /app/bin/proxy .
ADD . .

ENTRYPOINT ["/app/proxy"]

EXPOSE 8080
EXPOSE 8888

ENV PGPASSWORD 123
CMD service postgresql start