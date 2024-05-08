package pages

import (
	"fmt"
	"git.d7z.net/d7z-project/caddy-gitea-pages/pages"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"strings"
)

func init() {
	middle := &Middleware{
		Config: &pages.MiddlewareConfig{},
	}
	caddy.RegisterModule(middle)
	httpcaddyfile.RegisterHandlerDirective("gitea", parseCaddyfile(middle))
}

func (m *Middleware) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for n := d.Nesting(); d.NextBlock(n); {
			switch d.Val() {
			case "server":
				d.Args(&m.Config.Server)
			case "token":
				d.Args(&m.Config.Token)
			case "domain":
				d.Args(&m.Config.Domain)
			case "alias":
				d.Args(&m.Config.Alias)
			case "shared":
				m.Config.SharedAlias = true
			case "errors":
				if d.NextArg() {
					return d.ArgErr()
				}
				for nesting := d.Nesting(); d.NextBlock(nesting); {
					args := []string{d.Val()}
					args = append(args, d.RemainingArgs()...)
					if len(args) != 2 {
						return d.Errf("expected 2 arguments, got %d", len(args))
					}
					body, err := parseBody(args[1])
					if err != nil {
						return d.Errf("failed to parse %s: %v", args[0], err)
					}
					if m.Config.ErrorPages == nil {
						m.Config.ErrorPages = make(map[string]string)
					}
					m.Config.ErrorPages[strings.ToLower(args[0])] = body
				}
			case "redirect":
				m.Config.AutoRedirect = true
			case "proto":
				d.Args(&m.Config.ServerProto)
			default:
				return d.Errf("unrecognized subdirective '%s'", d.Val())
			}
		}
	}
	return nil
}

func parseBody(path string) (string, error) {
	fileData, err := os.ReadFile(path)
	if err == nil {
		return string(fileData), nil
	} else if strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://") {
		resp, err := http.Get(path)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", errors.New(resp.Status)
		}
		all, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return string(all), nil
	}
	return "", err
}

func parseCaddyfile(middleware *Middleware) func(httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	return func(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
		err := middleware.UnmarshalCaddyfile(h.Dispenser)
		if err != nil {
			return nil, err
		}
		if middleware.Config.ServerProto == "" {
			middleware.Config.ServerProto = "http"
		}
		return middleware, nil
	}
}

type Middleware struct {
	Config *pages.MiddlewareConfig `json:"config"`
	Logger *zap.Logger             `json:"-"`
	Client *pages.PageClient       `json:"-"`
}

func (m *Middleware) ServeHTTP(
	writer http.ResponseWriter,
	request *http.Request,
	handler caddyhttp.Handler,
) error {

	type stackTracer interface {
		StackTrace() errors.StackTrace
	}
	err := m.Client.Route(writer, request)
	if errors.Is(err, pages.ErrorNotMatches) {
		return handler.ServeHTTP(writer, request)
	} else {
		if err != nil {
			if err, ok := err.(stackTracer); ok {
				for _, f := range err.StackTrace() {
					fmt.Printf("%+s:%d\n", f, f)
				}
			}
		}
		return err
	}
}

func (m *Middleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "http.handlers.gitea",
		New: func() caddy.Module {
			return new(Middleware)
		},
	}
}

func (m *Middleware) Validate() error {
	err := m.Client.Validate()
	return err
}

func (m *Middleware) Provision(ctx caddy.Context) error {
	var err error
	m.Logger = ctx.Logger() // g.Logger is a *zap.Logger
	m.Client, err = pages.NewPageClient(
		m.Config,
		m.Logger,
	)
	if err != nil {
		return err
	}
	return nil
}

var (
	_ caddy.Provisioner           = (*Middleware)(nil)
	_ caddy.Validator             = (*Middleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*Middleware)(nil)
	_ caddyfile.Unmarshaler       = (*Middleware)(nil)
)
