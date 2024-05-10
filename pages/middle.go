package pages

import "time"

type MiddlewareConfig struct {
	Server       string            `json:"server"`
	Token        string            `json:"token"`
	Domain       string            `json:"domain"`
	Alias        string            `json:"alias"`
	CacheTimeout time.Duration     `json:"cache_timeout"`
	ErrorPages   map[string]string `json:"errors"`
	AutoRedirect *AutoRedirect     `json:"redirect"`
	SharedAlias  bool              `json:"shared_alias"`
	CacheMaxSize int               `json:"cache_max_size"`
}
