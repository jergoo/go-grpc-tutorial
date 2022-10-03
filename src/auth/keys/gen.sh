#! bin/sh

openssl req -newkey rsa:2048 -x509 -nodes -sha256 -days 3650 \
    -keyout server.key -new -out server.crt \
    -subj /CN=grpc.server -reqexts SAN -extensions SAN \
    -config <(cat /System/Library/OpenSSL/openssl.cnf \
        <(printf '[SAN]\nsubjectAltName=DNS:grpc.server'))
    
