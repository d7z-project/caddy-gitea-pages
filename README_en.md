# Gitea Pages Caddy Plugin

English (Google TR) | [中文](./README.md)

> Gitea Pages implemented with reference to Github Pages.

## Installation Instructions

`xcaddy` utility is required to generate the Caddy executable with the following command

```bash
xcaddy build --with github.com/d7z-project/caddy-gitea-pages
# List the current modules
. /caddy list-modules | grep gitea
```

We also provides `linux/amd64` and `linux/arm64` images:

```bash
docker pull ghcr.io/d7z-project/caddy-gitea-pages:nightly
```

## Configuration Notes

After installing Caddy, write the following configuration in `Caddyfile`.

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

More detailed configuration can be found in [Caddyfile](./Caddyfile)


## Usage Notes

The repository `https://gitea.com/owner/repo.git` corresponds to `owner.example.com/repo` in the example configuration.

To access the `CNAME` configured domain, you need to access the `<owner>.example.com/<repo>` domain of the repository, which needs to be done only once.

**Note**: You need to have `gh-pages` branch and `index.html` file in the branch to access the repository, if you still can't access it, you can restart Caddy to clear the cache.

## Acknowledgments

This project was inspired by [42wim/caddy-gitea](https://github.com/42wim/caddy-gitea).

## LICENSE

uses [Apache-2.0](./LICENSE)
