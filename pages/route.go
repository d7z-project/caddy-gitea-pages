package pages

import (
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"runtime"
	"strings"
)

func (p *PageClient) Route(writer http.ResponseWriter, request *http.Request) error {
	defer func() error {
		//放在匿名函数里,err捕获到错误信息，并且输出
		err := recover()
		if err != nil {
			p.logger.Error("recovered from panic", zap.Any("err", err))
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			println(string(buf[:n]))
			return p.ErrorPages.flushError(errors.New(fmt.Sprintf("%v", err)), request, writer)
		}
		return nil
	}()
	err := p.RouteExists(writer, request)
	if err != nil {
		p.logger.Debug("route exists error", zap.String("host", request.Host),
			zap.String("path", request.RequestURI), zap.Error(err))
		return p.ErrorPages.flushError(err, request, writer)
	}
	return err
}

func (p *PageClient) RouteExists(writer http.ResponseWriter, request *http.Request) error {
	domain, filePath, err := p.parseDomain(request)
	if err != nil {
		return err
	}
	config, cache, err := p.DomainCache.FetchRepo(p.GiteaConfig, domain)
	if err != nil {
		return err
	}
	if !config.Exists {
		return ErrorNotFound
	}
	if !cache {
		p.logger.Info("Add CNAME", zap.Any("CNAME", config.CNAME))
		p.DomainAlias.add(domain, config.CNAME...)
	}
	// 跳过 30x 重定向
	if p.AutoRedirect.Enabled &&
		len(config.CNAME) > 0 &&
		strings.HasPrefix(request.Host, domain.Owner+p.BaseDomain) {
		http.Redirect(writer, request, p.AutoRedirect.Scheme+"://"+config.CNAME[0], p.AutoRedirect.Code)
		return nil
	}

	_, err = config.Copy(p.GiteaConfig, filePath, writer, request, p.FileMaxCacheSize)
	return err
}
