VERSION := 0.0.2

dev:
	@xcaddy run v2.8.1 -c Caddyfile.local

fmt:
	@go fmt

image:
	@podman build -t ghcr.io/d7z-project/caddy-gitea-pages:$(VERSION) -f Dockerfile .