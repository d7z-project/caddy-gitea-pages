package pages

import (
	"code.gitea.io/sdk/gitea"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

type AutoRedirect struct {
	Enabled bool
	Scheme  string
	Code    int
}

type PageClient struct {
	BaseDomain       string
	GiteaConfig      *GiteaConfig
	DomainAlias      *CustomDomains
	ErrorPages       *ErrorPages
	AutoRedirect     *AutoRedirect
	OwnerCache       *OwnerCache
	DomainCache      *DomainCache
	logger           *zap.Logger
	FileMaxCacheSize int
}

func (p *PageClient) Close() error {
	if p.OwnerCache != nil {
		_ = p.OwnerCache.Close()
	}
	if p.DomainCache != nil {
		_ = p.DomainCache.Close()
	}
	return nil
}

func NewPageClient(
	config *MiddlewareConfig,
	logger *zap.Logger,
) (*PageClient, error) {
	options := make([]gitea.ClientOption, 0)
	if config.Token != "" {
		options = append(options, gitea.SetToken(config.Token))
	}
	options = append(options, gitea.SetGiteaVersion(""))
	client, err := gitea.NewClient(config.Server, options...)
	if err != nil {
		return nil, err
	}
	alias, err := NewCustomDomains(config.Alias, config.SharedAlias)
	if err != nil {
		return nil, err
	}
	pages, err := NewErrorPages(config.ErrorPages)
	if err != nil {
		return nil, err
	}
	ownerCache := NewOwnerCache(config.CacheTimeout)
	giteaConfig := &GiteaConfig{
		Server: config.Server,
		Token:  config.Token,
		Client: client,
		Logger: logger,
	}
	domainCache := NewDomainCache(config.CacheTimeout)
	logger.Info("gitea cache ttl " + strconv.FormatInt(config.CacheTimeout.Milliseconds(), 10) + " ms .")
	return &PageClient{
		GiteaConfig:      giteaConfig,
		BaseDomain:       "." + strings.Trim(config.Domain, "."),
		DomainAlias:      alias,
		ErrorPages:       pages,
		logger:           logger,
		AutoRedirect:     config.AutoRedirect,
		DomainCache:      &domainCache,
		OwnerCache:       &ownerCache,
		FileMaxCacheSize: config.CacheMaxSize,
	}, nil
}

func (p *PageClient) Validate() error {
	ver, _, err := p.GiteaConfig.Client.ServerVersion()
	p.logger.Info("Gitea Version ", zap.String("version", ver))
	if err != nil {
		p.logger.Warn("Failed to get Gitea version", zap.Error(err))
	}
	return nil
}
