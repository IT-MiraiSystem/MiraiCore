FROM node:lts-alpine AS build
WORKDIR /tmp
RUN apk update && apk upgrade && apk add --no-cache git
RUN git config --global http.sslVerify false
RUN git clone https://gitlab.joker.f5.si/appdeveloper/miraigate-for-teacher.git
WORKDIR /tmp/miraigate-for-teacher
RUN npm install
RUN npm run build
WORKDIR /tmp
RUN git clone https://github.com/swagger-api/swagger-ui.git
WORKDIR /tmp/swagger-ui/dist
RUN sed -i '/<\/body>/i <script>\n    window.onload = () => {\n        window.ui = SwaggerUIBundle({\n            url: '\''./openapi.yml'\'',\n            dom_id: '\''#swagger-ui'\'',\n            presets: [\n                SwaggerUIBundle.presets.apis,\n                SwaggerUIStandalonePreset\n            ],\n            layout: "StandaloneLayout",\n        });\n    };\n</script>' /tmp/miraigate-for-teacher/dist/index.html

FROM nginx:alpine
ENV TZ="Asia/Tokyo"
RUN echo $TZ > /etc/timezone
COPY --from=build /tmp/miraigate-for-teacher/dist /var/www/html/teacher
COPY --from=build /tmp/swagger-ui/dist /var/www/html/document/api/
COPY openapi.yml /var/www/html/document/api/openapi.yml
COPY web/nginx.conf /etc/nginx/conf.d/default.conf
CMD ["nginx", "-g", "daemon off;"]