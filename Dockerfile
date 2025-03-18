FROM golang:alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app ./cmd/service-propusk/main.go

FROM nginx:alpine

WORKDIR /app

COPY --from=builder /app/app /app/app

RUN apk add --no-cache openssl

RUN chmod +x /app/app

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

COPY nginx.conf.template /etc/nginx/nginx.conf.template

CMD [ "/entrypoint.sh" ]