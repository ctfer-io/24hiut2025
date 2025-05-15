FROM alpine:3.19

WORKDIR /app

COPY ./src/ /app/

RUN apk update && apk add --no-cache \
    python3~=3 \
    python3-dev~=3 \
    bash \
    py3-pip~=23 \
    chromium \
    npm \
    chromium-chromedriver && \
pip install --no-cache-dir --break-system-packages \
    -r requirements.txt && \
npm install tinymce@^7


EXPOSE 8080

CMD ["python3", "/app/app.py"]