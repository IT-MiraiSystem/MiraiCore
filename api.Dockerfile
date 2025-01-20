FROM golang:1.23.4
WORKDIR /usr/bin/MiraiCore
RUN mkdir certification
WORKDIR /usr/bin/MiraiCore/certification
RUN openssl genrsa -out secret.key 4096
RUN openssl rsa -in secret.key -pubout -out publickey.pem
WORKDIR /usr/bin/MiraiCore/src
RUN apt update && apt upgrade -y && apt autoremove -y
ENV TZ="Asia/Tokyo"
RUN echo $TZ > /etc/timezone
RUN go mod init gitlab.joker.f5.si/appdeveloper/miraicore.git
RUN go mod tidy
CMD ["go","run","."]