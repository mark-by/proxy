FROM postgres:alpine

ADD build/createdb.sh /docker-entrypoint-initdb.d/create.sh
RUN chmod +x /docker-entrypoint-initdb.d/create.sh

EXPOSE 5432