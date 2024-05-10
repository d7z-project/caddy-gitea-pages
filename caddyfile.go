package pages

import (
	"fmt"
	"git.d7z.net/d7z-project/caddy-gitea-pages/pages"
	"github.com/alecthomas/units"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
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
			case "cache":
				remainingArgs := d.RemainingArgs()
				if len(remainingArgs) != 3 {
					return d.Errf("expected 3 argument for 'cache'; got %v", remainingArgs)
				}
				var err error
				m.Config.CacheTimeout, err = time.ParseDuration(remainingArgs[0])
				if err != nil {
					return d.Errf("invalid duration: %v", err)
				}
				m.Config.CacheRefresh, err = time.ParseDuration(remainingArgs[1])
				if err != nil {
					return d.Errf("invalid duration: %v", err)
				}
				size, err := units.ParseBase2Bytes(remainingArgs[2])
				if err != nil {
					return d.Errf("invalid CacheSize: %v", err)
				}
				m.Config.CacheMaxSize = int(size)
			case "domain":
				d.Args(&m.Config.Domain)
			case "alias":
				remainingArgs := d.RemainingArgs()
				if len(remainingArgs) == 0 {
					return d.Errf("expected 2 argument for 'alias'; got %v", remainingArgs)
				}
				if len(remainingArgs) == 1 {
					m.Config.Alias = remainingArgs[0]
				}
				if len(remainingArgs) == 2 && remainingArgs[1] == "shared" {
					m.Config.Alias = remainingArgs[0]
					m.Config.SharedAlias = true
				}
			case "headers":
				if d.NextArg() {
					return d.ArgErr()
				}
				for nesting := d.Nesting(); d.NextBlock(nesting); {
					args := []string{d.Val()}
					args = append(args, d.RemainingArgs()...)
					if len(args) != 2 {
						return d.Errf("expected 2 arguments, got %d", len(args))
					}
					if m.Config.CustomHeaders == nil {
						m.Config.CustomHeaders = make(map[string]string)
					}
					m.Config.CustomHeaders[args[0]] = args[1]
				}
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
				remainingArgs := d.RemainingArgs()
				if len(remainingArgs) != 2 {
					return d.Errf("expected 2 arguments, got %d", len(remainingArgs))
				}
				code, err := strconv.Atoi(remainingArgs[1])
				if err != nil {
					return d.WrapErr(err)
				}
				m.Config.AutoRedirect = &pages.AutoRedirect{
					Enabled: true,
					Scheme:  remainingArgs[0],
					Code:    code,
				}
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
		if middleware.Config.ErrorPages == nil {
			middleware.Config.ErrorPages = make(map[string]string)
		}
		if middleware.Config.CustomHeaders == nil {
			middleware.Config.CustomHeaders = make(map[string]string)
		}
		if middleware.Config.AutoRedirect == nil {
			middleware.Config.AutoRedirect = &pages.AutoRedirect{
				Enabled: false,
			}
		}
		if middleware.Config.CacheRefresh <= 0 {
			middleware.Config.CacheRefresh = 1 * time.Minute
		}
		if middleware.Config.CacheTimeout <= 0 {
			middleware.Config.CacheTimeout = 3 * time.Minute
		}
		if middleware.Config.CacheMaxSize <= 0 {
			middleware.Config.CacheMaxSize = 3 * 1024 * 1024
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
	return m.Client.Validate()
}

func (m *Middleware) Cleanup() error {
	m.Logger.Info("cleaning up gitea middleware.")
	return m.Client.Close()
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
	_ caddy.CleanerUpper          = (*Middleware)(nil)
	_ caddy.Validator             = (*Middleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*Middleware)(nil)
	_ caddyfile.Unmarshaler       = (*Middleware)(nil)
)
