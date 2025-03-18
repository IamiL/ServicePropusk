#!/bin/sh

# Генерация самоподписанного SSL-сертификата
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/ssl/private/cert.key \
  -out /etc/ssl/certs/cert.crt \
  -subj "/C=US/ST=State/L=City/O=Organization/OU=Org/CN=localhost"

# Замена маркеров в конфиге Nginx
envsubst < /etc/nginx/nginx.conf.template > /etc/nginx/nginx.conf

# (Отладочная строка) Проверка содержимого nginx.conf
cat /etc/nginx/nginx.conf

# Запуск Go приложения в фоновом режиме
/app &

# Запуск Nginx
nginx -g 'daemon off;'