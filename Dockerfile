FROM golang:1.23.4
WORKDIR /usr/bin/MiraiCore/src
RUN apt update && apt upgrade -y && apt autoremove -y
RUN apt install ufw -y
RUN ufw allow 80 && ufw allow 443 && ufw allow 22
RUN go mod init gitlab.joker.f5.si/appdeveloper/miraicore.git
RUN go mod tidy
CMD ["go","run","main.go"]