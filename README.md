# Gitea Pages

> 参照 Github Pages 实现的 Gitea Pages

## 安装说明

此处需要用到 `xcaddy` 工具，使用如下命令生成 Caddy 执行文件

```bash
xcaddy build --with git.d7z.net/d7z-project/caddy-gitea-pages
# 列出当前模块
./caddy list-modules | grep gitea
```

## 配置说明

安装后 Caddy 后写入如下配置，在 DNS 控制台绑定域名和 IP 即可。

例如： `https://gitea.com/owner/repo.git` 对应如下配置中的 `owner.example.com/repo`

注意： 需要仓库存在 `gh-pages` 分支和分支内存在 `CNAME` 文件才可访问，如果配置后仍无法访问可重启 Caddy 来清理缓存。

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
   # CNAME 配置缓存地址
   alias path/to/file
   # 默认 404 页面,可填写路径或者 URL
   error40x path/to/file
   # 默认 50x 页面,可填写路径或者 URL
   error50x path/to/file
}
```

其中，token 需要如下权限：

- organization:read
- repository:read
- user:read

## 致谢

此项目参考了 [42wim/caddy-gitea](https://github.com/42wim/caddy-gitea)

## LICENSE

此项目使用 [Apache-2.0](./LICENSE)