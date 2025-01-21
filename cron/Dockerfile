FROM willfarrell/crontab:latest

COPY ./timetable/ /usr/bin/timetable
COPY ../config /usr/bin/timetable/config
RUN apt update && apt install -y python3-pip python3-venv
RUN pip3 install -r /usr/bin/timetable/requirements.txt
COPY ./config.json /opt/crontab

RUN chmod +x /scripts/backup.sh