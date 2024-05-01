package pages

type MiddlewareConfig struct {
	Server       string `json:"server"`
	Token        string `json:"token"`
	Domain       string `json:"domain"`
	Alias        string `json:"alias"`
	Error40xPage string `json:"error40x"`
	Error50xPage string `json:"error50x"`
	AutoRedirect bool   `json:"redirect"`
	ServerProto  string `json:"proto"`
	SharedAlias  bool   `json:"shared_alias"`
}
