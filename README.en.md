# Gitea Pages

English|[中文](README.md)

> Gitea Pages implementation inspired by Github Pages  
> 
> This project is currently in maintenance mode. You may also check out my other project: [d7z-project/gitea-pages](https://github.com/d7z-project/gitea-pages)

## Installation Guide

The `xcaddy` tool is required for this step. Use the following command to generate the Caddy executable.  
If `xcaddy` is not installed, download it from [caddyserver/xcaddy](https://github.com/caddyserver/xcaddy/releases) first.  
Additionally, ensure Golang 1.24 is installed.

```bash
xcaddy build v2.10.0 --with github.com/d7z-project/caddy-gitea-pages
# List current modules
./caddy list-modules | grep gitea
```

This project also provides Docker images for `linux/amd64` and `linux/arm64`:

```bash
docker pull ghcr.io/d7z-project/caddy-gitea-pages:nightly
```

For configuration details, refer to the `docker.io/library/caddy` image.

## Configuration Guide

After installing Caddy, add the following configuration to your `Caddyfile`:

```conf
{
    order gitea before file_server
}

:80
gitea {
   # Gitea server address
   server https://gitea.com
   # Gitea Token
   token please-replace-it
   # Default domain, similar to Github's github.io
   domain example.com
}
```

The token requires the following permissions:  
- `organization:read`  
- `repository:read`  
- `user:read`  

For more detailed configurations, see [Caddyfile](./Caddyfile).

## Usage Instructions

The repository `https://gitea.com/owner/repo.git` corresponds to `owner.example.com/repo` in the example configuration.  

To access domains configured via `CNAME`, you must first visit the repository's `<owner>.example.com/<repo>` URL. This step only needs to be performed once.  

**Note**: The repository must have a `gh-pages` branch containing an `index.html` file for access. If issues persist after configuration, restart Caddy to clear the cache.  

### Fallback Strategy  
- Appends `index.html` automatically when the URL ends with `/`.  
- If a file is not found and `404.html` exists, it will be served with a 404 status code.  
- For repositories tagged with `routes-history` or `routes-hash`, the fallback uses `index.html` with a 200 status code by default.  

## TODO  
- [x] Support CNAME custom paths (HTTP mode only, no ACME handling)  
- [x] Support content caching  
- [ ] Optimize concurrency model and race condition handling  
- [ ] Support HTTP Range for resumable downloads  
- [ ] Support OAuth2 login for private page access  

## Acknowledgments  
This project references [42wim/caddy-gitea](https://github.com/42wim/caddy-gitea).  

## LICENSE  
This project is licensed under [Apache-2.0](./LICENSE).  