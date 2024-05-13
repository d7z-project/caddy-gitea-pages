FROM docker.io/library/caddy:2.7-builder-alpine as builder
RUN mkdir -p /usr/local/src
COPY go.mod go.sum caddyfile.go /usr/local/src/
COPY pages /usr/local/src/pages
ENV GO111MODULE=on GOPROXY=https://goproxy.cn
WORKDIR /usr/local/src
RUN ls && xcaddy build \
    --with git.d7z.net/d7z-project/caddy-gitea-pages=./
FROM docker.io/library/caddy:2.7
COPY --from=builder  /usr/local/src/caddy /usr/bin/caddy