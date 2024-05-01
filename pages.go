package pages

import "github.com/caddyserver/caddy/v2"

func init() {

}

type GiteaPage struct {
}

func (p *GiteaPage) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "http.handlers.gitea",
		New: func() caddy.Module {
			return new(GiteaPage)
		},
	}
}
