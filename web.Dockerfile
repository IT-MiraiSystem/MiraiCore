FROM node:lts-alpine AS build
WORKDIR /tmp
RUN apk update && apk upgrade && apk add --no-cache git
RUN git config --global http.sslVerify false
RUN git clone https://gitlab.joker.f5.si/appdeveloper/miraigate-for-teacher.git
WORKDIR /tmp/miraigate-for-teacher
RUN npm install
RUN npm run build

FROM nginx:alpine
ENV TZ="Asia/Tokyo"
RUN echo $TZ > /etc/timezone
COPY --from=build /tmp/miraigate-for-teacher/dist /var/www/html/teacher
COPY nginx/nginx.conf /etc/nginx/conf.d/default.conf
CMD ["nginx", "-g", "daemon off;"]