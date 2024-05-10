package pages

import "github.com/pkg/errors"

var (
	// ErrorNotMatches 确认这不是 Gitea Pages 相关的域名
	ErrorNotMatches = errors.New("not matching")
	ErrorNotFound   = errors.New("not found")
	ErrorInternal   = errors.New("internal error")
)
