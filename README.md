# Gitea Pages

[English (Google TR)](./README_en.md) | 中文

> 参照 Github Pages 实现的 Gitea Pages

## 安装说明

此处需要用到 `xcaddy` 工具，使用如下命令生成 Caddy 执行文件，
如果 `xcaddy` 不存在，需先前往 [caddyserver/xcaddy](https://github.com/caddyserver/xcaddy/releases) 安装 `xcaddy`,
同时安装好 Golang 1.22

```bash
xcaddy build v2.8.4 --with github.com/d7z-project/caddy-gitea-pages
# 列出当前模块
./caddy list-modules | grep gitea
```

当前项目也提供 `linux/amd64` 和 `linux/arm64` 的镜像:

```bash
docker pull ghcr.io/d7z-project/caddy-gitea-pages:nightly
```

具体配置说明参考 `docker.io/library/caddy` 镜像。

## 配置说明

安装后 Caddy 后, 在 `Caddyfile` 写入如下配置:

```conf
{
    order gitea before file_server
}

:80
gitea {
   # Gitea 服务器地址
   server https://gitea.com
   # Gitea Token
   token please-replace-it
   # 默认域名，类似于 Github 的 github.io
   domain example.com
}
```

其中，token 需要如下权限：

- `organization:read`
- `repository:read`
- `user:read`

更详细的配置可查看 [Caddyfile](./Caddyfile)

## 使用说明

仓库 `https://gitea.com/owner/repo.git` 对应示例配置中的 `owner.example.com/repo`

如需访问 `CNAME` 配置的域名，则需要先访问仓库对应的 `<owner>.example.com/<repo>` 域名, 此操作只需完成一次。

**注意**： 需要仓库存在 `gh-pages` 分支和分支内存在 `index.html` 文件才可访问，如果配置后仍无法访问可重启 Caddy 来清理缓存。

### 文件回退策略

- URL 末尾为 `/` 时将自动追加 `index.html`
- 未找到文件时，如果存在 `404.html` 将使用此文件，响应 404 状态码
- 如果仓库带有 `routes-history` 和 `routes-hash` 标签时，默认回退使用 `index.html`, 同时返回 200 状态码

## TODO

- [x] 支持 CNAME 自定义路径 (仅适用于 HTTP 模式，不处理 acme 相关的内容)
- [x] 支持内容缓存
- [ ] 优化并发模型和处理竞争问题
- [ ] 支持 Http Range 断点续传
- [ ] 支持 oauth2 登录访问私有页面

## 致谢

此项目参考了 [42wim/caddy-gitea](https://github.com/42wim/caddy-gitea)

## LICENSE

此项目使用 [Apache-2.0](./LICENSE)
