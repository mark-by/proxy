#!/bin/sh

mkdir certs/
openssl genrsa -out root.key 2048
openssl req -new -x509 -days 3650 -key root.key -out root.crt -subj "/CN=golang proxy CA"
openssl genrsa -out cert.key 2048
#sudo cp root.crt /usr/local/share/ca-certificates/
sudo cp root.crt /etc/ssl/certs/
sudo update-ca-certificates
