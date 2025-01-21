#!/bin/sh
apt update && apt upgrade -y && apt autoremove -y
apt install -y openssl 
mkdir -p /usr/bin/MiraiCore/certification
cd /usr/bin/MiraiCore/certification
if [ ! -f secret.key ]; then openssl genrsa -out secret.key 4096; fi
if [ ! -f publickey.pem ]; then openssl rsa -in secret.key -pubout -out publickey.pem; fi
echo "Starting MiraiCore"

cd /usr/bin/MiraiCore/src
exec "$@"