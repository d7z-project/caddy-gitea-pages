# Gitea Pages Caddy Plugin

English (Google TR) | [中文](./README_zh.md)

> Gitea Pages implemented with reference to Github Pages.

## Installation Instructions

`xcaddy` utility is required to generate the Caddy executable with the following command

```bash
xcaddy build --with git.d7z.net/d7z-project/caddy-gitea-pages
# List the current modules
. /caddy list-modules | grep gitea
```

## Configuration Notes

After installing Caddy, write the following configuration in ``Caddyfile``.

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

`https://gitea.com/owner/repo.git` corresponds to `owner.example.com/repo` in the example configuration; if you want to access the domain name configured by `CNAME`, you need to access the corresponding `<owner>.example.com/<repo>` domain of the repository to establish the link first. relationship , when accessing `owner.example.com` it will hit `gita.com/owner/owner.example.com` repo

**Note**: You need to have `gh-pages` branch and `CNAME` file in the branch to access the repository, if you still can't access it, you can restart Caddy to clear the cache.

## Acknowledgments

This project is referenced in [42wim/caddy-gitea](https://github.com/42wim/caddy-gitea).

## LICENSE

uses [Apache-2.0](./LICENSE)
