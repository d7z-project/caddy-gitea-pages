package pages

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
)

func (p *PageClient) FileExists(domain *PageDomain, path string) (bool, error) {
	context, err := p.OpenFileContext(domain, path)
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

func (p *PageClient) ReadStringRepoFile(domain *PageDomain, path string) (string, error) {
	data, err := p.ReadRepoFile(domain, path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (p *PageClient) ReadRepoFile(domain *PageDomain, path string) ([]byte, error) {
	context, err := p.OpenFileContext(domain, path)
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

func (p *PageClient) OpenFileContext(domain *PageDomain, path string) (*http.Response, error) {
	var (
		giteaURL string
		err      error
	)
	// gitea sdk doesn't support "media" type for lfs/non-lfs
	giteaURL, err = url.JoinPath(p.Server+"/api/v1/repos/", domain.Owner, domain.Repo, "media", path)
	if err != nil {
		return nil, err
	}
	giteaURL += "?ref=" + url.QueryEscape(domain.Branch)
	req, err := http.NewRequest(http.MethodGet, giteaURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "token "+p.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	switch resp.StatusCode {
	case http.StatusForbidden:
		return nil, errors.Wrap(ErrorNotFound, "domain file not forbidden")
	case http.StatusNotFound:
		return nil, errors.Wrap(ErrorNotFound, "domain file not found")
	case http.StatusOK:
	default:
		return nil, errors.New(fmt.Sprintf("unexpected status code '%d'", resp.StatusCode))
	}
	return resp, nil
}
