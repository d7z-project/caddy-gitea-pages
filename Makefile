VERSION := 0.0.3

dev:
	@xcaddy run v2.10.0 -c Caddyfile.local

fmt:
	@go fmt

image:
	@podman build -t ghcr.io/d7z-project/caddy-gitea-pages:$(VERSION) -f Dockerfile .