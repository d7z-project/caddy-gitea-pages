# Gitea Pages

[English (Google TR)](./README_en.md) | 中文

> 参照 Github Pages 实现的 Gitea Pages

## 安装说明

此处需要用到 `xcaddy` 工具，使用如下命令生成 Caddy 执行文件

```bash
xcaddy build --with git.d7z.net/d7z-project/caddy-gitea-pages
# 列出当前模块
./caddy list-modules | grep gitea
```

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

## TODO

- [x] 支持 CNAME
- [x] 支持内容缓存
- [ ] 支持 oauth2 登录访问私有页面

## 致谢

此项目参考了 [42wim/caddy-gitea](https://github.com/42wim/caddy-gitea)

## LICENSE

此项目使用 [Apache-2.0](./LICENSE)
