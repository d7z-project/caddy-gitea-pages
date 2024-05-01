package pages

import (
	"code.gitea.io/sdk/gitea"
	"github.com/allegro/bigcache/v3"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io"
	"mime"
	"net/http"
	"path"
	"strings"
)

func (p *PageClient) Route(writer http.ResponseWriter, request *http.Request) error {
	err := p.RouteExists(writer, request)
	if err != nil {
		if errors.Is(err, ErrorNotFound) {
			err := p.ErrorPages.flush40xError(writer)
			if err != nil {
				return err
			}
			return err
		} else if !errors.Is(err, ErrorNotMatches) {
			err := p.ErrorPages.flush50xError(writer)
			if err != nil {
				return err
			}
			return err
		}
	}
	return err
}

func (p *PageClient) RouteExists(writer http.ResponseWriter, request *http.Request) error {
	domain, filePath, err := p.parseDomain(request)
	if err != nil {
		return err
	}
	config, err := p.parseDomainConfig(domain)
	if request.Host != config.Alias && p.AutoRedirect {
		http.Redirect(writer, request, p.ServerProto+"://"+config.Alias, 302)
		return nil
	}
	if err != nil {
		return err
	}
	context, err := p.OpenFileContext(domain, config.RootPath+filePath)
	if errors.Is(err, ErrorNotFound) && config.NotFoundPath != "" {
		context, err = p.OpenFileContext(domain, config.RootPath+config.NotFoundPath)
	}
	if err != nil {
		return err
	}
	contentType := context.Header.Get("Content-Type")
	if contentType != "application/octet-stream" {
		contentType = mime.TypeByExtension(path.Ext(filePath))
	}
	writer.Header().Add("Content-Type", contentType)
	writer.WriteHeader(http.StatusOK)
	defer context.Body.Close()
	_, err = io.Copy(writer, context.Body)
	return err
}

func (p *PageClient) parseDomainConfig(domain *PageDomain) (*DomainConfig, error) {
	unlock := p.pagesConfig.locker.LockAny(domain.key())
	defer unlock()
	cache, err := p.pagesConfig.getDomainConfig(domain)
	if errors.Is(err, bigcache.ErrEntryNotFound) {
		cache = &DomainConfig{
			RootPath:     "",
			NotFoundPath: "index.html",
		}
		defer func() error {
			return p.pagesConfig.setDomainConfig(domain, cache)
		}()
		/////////// 处理 CNAME
		alias, err := p.ReadStringRepoFile(domain, "CNAME")
		if err != nil {
			// 这不是一个可用的仓库
			if errors.Is(err, ErrorNotFound) {
				return nil, err
			}
			return nil, errors.Wrap(err, "unknown error!")
		}
		alias = strings.TrimSpace(alias)
		cache.Alias = alias
		// 添加映射
		if alias != "" {
			p.logger.Info("添加 alias ", zap.String("alias", strings.TrimSpace(alias)))
			p.DomainAlias.add(domain, alias)
		}
		/////////// 检查 404 文件是否存在
		exists, err := p.FileExists(domain, cache.RootPath+"404.html")
		if err != nil {
			return nil, err
		}
		if exists {
			cache.NotFoundPath = "404.html"
		}
		//////////

	} else if err != nil {
		return nil, err
	}
	return cache, nil
}

func (p *PageClient) parseDomain(request *http.Request) (*PageDomain, string, error) {
	// TODO: 处理 IPv6 Host:Port 的问题
	host := strings.Split(request.Host, ":")[0]
	filePath := request.URL.Path
	if strings.HasSuffix(filePath, "/") {
		filePath = filePath + "index.html"
	}
	pathTrim := strings.Split(strings.Trim(filePath, "/"), "/")
	if strings.HasSuffix(host, p.BaseDomain) {
		child := strings.Split(strings.TrimSuffix(host, p.BaseDomain), ".")
		result := NewPageDomain(
			child[len(child)-1],
			pathTrim[0],
			"gh-pages",
		)
		// 处于使用默认 Domain 下
		config, err := p.getOwnerConfig(result.Owner)
		if err != nil {
			return nil, "", err
		}
		ownerRepoName := result.Owner + p.BaseDomain
		if result.Repo == "" && config.Exists(ownerRepoName) {
			// 推导为默认仓库
			result.Repo = ownerRepoName
			return result, filePath, nil
		} else if result.Repo == "" || !config.Exists(result.Repo) {
			// 未指定 repo 或者 repo 不存在，跳过
			return nil, "", errors.Wrap(ErrorNotFound, result.Repo+" not found")
		}
		// 存在子目录且仓库存在
		pathTrim = pathTrim[1:]
		if strings.HasSuffix(filePath, "/") {
			return result, "/" + strings.Join(pathTrim, "/") + "/", nil
		} else {
			return result, "/" + strings.Join(pathTrim, "/"), nil
		}
	} else {
		get, exists := p.DomainAlias.Get(host)
		if exists {
			return &get, filePath, nil
		} else {
			return nil, "", errors.Wrap(ErrorNotFound, "")
		}
	}
}

func (p *PageClient) getOwnerConfig(owner string) (*OwnerConfig, error) {
	unlock := p.pagesConfig.locker.LockAny(owner)
	defer unlock()
	cache, err := p.pagesConfig.GetOwnerCache(owner)
	if errors.Is(err, bigcache.ErrEntryNotFound) {
		// not exists
		cache = NewOwnerConfig()
		defer func() error {
			return p.pagesConfig.setOwnConfig(owner, cache)
		}()
		//////////////// 列出组织下所有仓库
		repos, resp, err := p.client.ListOrgRepos(owner, gitea.ListOrgReposOptions{
			ListOptions: gitea.ListOptions{
				PageSize: 999,
			},
		})
		if err != nil && resp.StatusCode == http.StatusNotFound {
			// 调用用户接口查询
			repos, resp, err = p.client.ListUserRepos(owner, gitea.ListReposOptions{
				ListOptions: gitea.ListOptions{
					PageSize: 999,
				},
			})
			if err != nil {
				return nil, errors.Wrap(err, "")
			}
		} else if err != nil {
			return nil, err
		}
		for _, repo := range repos {
			cache.Repos[repo.Name] = true
			cache.LowerRepos[strings.ToLower(repo.Name)] = true
		}
		////////////////////
	} else if err != nil {
		return nil, err
	}
	return cache, nil
}
