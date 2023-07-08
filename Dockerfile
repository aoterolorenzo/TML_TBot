## Build
FROM bitnami/golang:1.20-debian-11 as build

WORKDIR /app
COPY . .

RUN apk update \
        && apk upgrade \
        && apk add --no-cache \
        ca-certificates \
        && update-ca-certificates 2>/dev/null || true

RUN go build -o /bot

## Deploy
FROM chromedp/headless-shell:latest as deploy

WORKDIR /
COPY --from=build /bot /bot
COPY ./assets /assets
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/


USER 1001:1001

ENTRYPOINT ["/bot"]



