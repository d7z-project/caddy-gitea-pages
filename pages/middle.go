package pages

type MiddlewareConfig struct {
	Server       string            `json:"server"`
	Token        string            `json:"token"`
	Domain       string            `json:"domain"`
	Alias        string            `json:"alias"`
	ErrorPages   map[string]string `json:"errors"`
	AutoRedirect bool              `json:"redirect"`
	ServerProto  string            `json:"proto"`
	SharedAlias  bool              `json:"shared_alias"`
}
