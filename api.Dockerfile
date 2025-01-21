FROM golang:1.23.4
WORKDIR /usr/bin/MiraiCore/src
ENV TZ="Asia/Tokyo"
RUN echo $TZ > /etc/timezone
RUN go mod init gitlab.joker.f5.si/appdeveloper/miraicore.git
RUN go mod tidy
ENTRYPOINT [ "/usr/bin/MiraiCore/entrypoint.sh" ]
CMD ["go","run","."]