package pages

import (
	"code.gitea.io/sdk/gitea"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"strings"
)

var (
	// ErrorNotMatches 确认这不是 Gitea Pages 相关的域名
	ErrorNotMatches = errors.New("not matching")
	ErrorNotFound   = errors.New("not found")
)

type PageClient struct {
	Server       string
	Token        string
	BaseDomain   string
	client       *gitea.Client
	DomainAlias  *CustomDomains
	pagesConfig  *PageConfigGroup
	ErrorPages   *ErrorPages
	AutoRedirect bool
	ServerProto  string
	logger       *zap.Logger
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
	pages, err := NewErrorPages(config.Error40xPage, config.Error50xPage)
	if err != nil {
		return nil, err
	}
	return &PageClient{
		Server:       config.Server,
		Token:        config.Token,
		BaseDomain:   "." + strings.Trim(config.Domain, "."),
		client:       client,
		DomainAlias:  alias,
		pagesConfig:  NewDomainConfig(),
		ErrorPages:   pages,
		logger:       logger,
		AutoRedirect: config.AutoRedirect,
		ServerProto:  config.ServerProto,
	}, nil
}

func (p *PageClient) Validate() error {
	ver, _, err := p.client.ServerVersion()
	p.logger.Info("Gitea Version ", zap.String("version", ver))
	if err != nil {
		return err
	}
	return nil
}
