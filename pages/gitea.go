package pages

import (
	"code.gitea.io/sdk/gitea"
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
)

type GiteaConfig struct {
	Server        string            `json:"server"`
	Token         string            `json:"token"`
	Client        *gitea.Client     `json:"-"`
	Logger        *zap.Logger       `json:"-"`
	CustomHeaders map[string]string `json:"custom_headers"`
	CacheMaxSize  int               `json:"max_cache_size"`
}

func (c *GiteaConfig) FileExists(domain *PageDomain, path string) (bool, error) {
	context, err := c.OpenFileContext(domain, path)
	if context != nil {
		defer context.Body.Close()
	}
	if errors.Is(err, ErrorNotFound) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (c *GiteaConfig) ReadStringRepoFile(domain *PageDomain, path string) (string, error) {
	data, err := c.ReadRepoFile(domain, path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *GiteaConfig) ReadRepoFile(domain *PageDomain, path string) ([]byte, error) {
	context, err := c.OpenFileContext(domain, path)
	if err != nil {
		return nil, err
	}
	defer context.Body.Close()
	all, err := io.ReadAll(context.Body)
	if err != nil {
		return nil, err
	}
	return all, nil
}

func (c *GiteaConfig) OpenFileContext(domain *PageDomain, path string) (*http.Response, error) {
	var (
		giteaURL string
		err      error
	)
	giteaURL, err = url.JoinPath(c.Server+"/api/v1/repos/", domain.Owner, domain.Repo, "media", path)
	if err != nil {
		return nil, err
	}
	giteaURL += "?ref=" + url.QueryEscape(domain.Branch)
	req, err := http.NewRequest(http.MethodGet, giteaURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "token "+c.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	switch resp.StatusCode {
	case http.StatusForbidden:
		return nil, errors.Wrap(ErrorNotFound, "domain file not forbidden")
	case http.StatusNotFound:
		return nil, errors.Wrap(ErrorNotFound, fmt.Sprintf("domain file not found: %s", path))
	case http.StatusOK:
	default:
		return nil, errors.Wrap(ErrorInternal, fmt.Sprintf("unexpected status code '%d'", resp.StatusCode))
	}
	return resp, nil
}
