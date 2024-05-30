FROM docker.io/library/caddy:2.8-builder-alpine as builder
RUN mkdir -p /usr/local/src
COPY go.mod go.sum caddyfile.go /usr/local/src/
COPY pages /usr/local/src/pages
WORKDIR /usr/local/src
RUN ls && xcaddy build \
    --with github.com/d7z-project/caddy-gitea-pages=./
FROM docker.io/library/caddy:2.8
COPY --from=builder  /usr/local/src/caddy /usr/bin/caddy