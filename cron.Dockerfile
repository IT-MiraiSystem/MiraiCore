FROM willfarrell/crontab:latest

COPY ./cron/timetable/ /usr/bin/timetable
COPY ./config /usr/bin/timetable/config
RUN apk update && apk upgrade && apk add python3 py3-pip
RUN pip install -r /usr/bin/timetable/requirements.txt
COPY ./config/config.json /opt/crontab