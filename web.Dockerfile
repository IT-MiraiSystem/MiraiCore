FROM node:23-alpine AS build
WORKDIR /tmp
RUN apk update && apk upgrade && apk add --no-cache git
RUN git config --global http.sslVerify false
RUN git clone https://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@github.com/IT-MiraiSystem/MiraiGate_For_Teacher.git
WORKDIR /tmp/MiraiGate_For_Teacher
RUN npm install
RUN npm run build
WORKDIR /tmp
RUN git clone https://github.com/swagger-api/swagger-ui.git
RUN sed -i '/<\/body>/i <script>\n    window.onload = () => {\n        window.ui = SwaggerUIBundle({\n            url: '\''./openapi.yml'\'',\n            dom_id: '\''#swagger-ui'\'',\n            presets: [\n                SwaggerUIBundle.presets.apis,\n                SwaggerUIStandalonePreset\n            ],\n            layout: "StandaloneLayout",\n        });\n    };\n</script>' /tmp/swagger-ui/dist/index.html

FROM nginx:alpine
ENV TZ="Asia/Tokyo"
RUN echo $TZ > /etc/timezone
COPY --from=build /tmp/MiraiGate_For_Teacher/dist /var/www/html/teacher
COPY --from=build /tmp/swagger-ui/dist /var/www/html/document/api/
COPY openapi.yml /var/www/html/document/api/openapi.yml
COPY web/nginx.conf /etc/nginx/conf.d/default.conf
CMD ["nginx", "-g", "daemon off;"]