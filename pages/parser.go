package pages

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

func (p *PageClient) parseDomain(request *http.Request) (*PageDomain, string, error) {
	// TODO: 处理 IPv6 Host:Port 的问题
	host := strings.Split(request.Host, ":")[0]
	filePath := request.URL.Path
	pathTrim := strings.Split(strings.Trim(filePath, "/"), "/")
	repo := pathTrim[0]
	// 处理 scheme://domain/path 的情况
	if !strings.HasPrefix(filePath, fmt.Sprintf("/%s/", repo)) {
		repo = ""
	}
	if strings.HasSuffix(host, p.BaseDomain) {
		child := strings.Split(strings.TrimSuffix(host, p.BaseDomain), ".")
		result := NewPageDomain(
			child[len(child)-1],
			repo,
			"gh-pages",
		)
		// 处于使用默认 Domain 下
		config, err := p.OwnerCache.GetOwnerConfig(p.GiteaConfig, result.Owner)
		if err != nil {
			return nil, "", err
		}
		ownerRepoName := result.Owner + p.BaseDomain
		if result.Repo == "" && config.Exists(ownerRepoName) {
			// 推导为默认仓库
			result.Repo = ownerRepoName
			return result, filePath, nil
		} else if result.Repo == "" || !config.Exists(result.Repo) {
			if config.Exists(ownerRepoName) {
				result.Repo = ownerRepoName
				return result, filePath, nil
			}
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
