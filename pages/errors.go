package pages

import (
	"embed"
	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"text/template"
)

var (
	// ErrorNotMatches 确认这不是 Gitea Pages 相关的域名
	ErrorNotMatches = errors.New("not matching")
	ErrorNotFound   = errors.New("not found")
)

type ErrorMetadata struct {
	StatusCode int
	Request    *http.Request
	Error      string
}

var (
	//go:embed 40x.gohtml 50x.gohtml
	embedPages embed.FS
)

type ErrorPages struct {
	errorPages map[string]*template.Template
}

func newTemplate(key, text string) (*template.Template, error) {
	return template.New(key).Funcs(sprig.TxtFuncMap()).Parse(text)
}
func NewErrorPages(pagesTmpl map[string]string) (*ErrorPages, error) {
	pages := make(map[string]*template.Template)
	for key, value := range pagesTmpl {
		tmpl, err := newTemplate(key, value)
		if err != nil {
			return nil, err
		}
		pages[key] = tmpl
	}

	if pages["40x"] == nil {
		data, err := embedPages.ReadFile("40x.gohtml")
		if err != nil {
			return nil, err
		}
		pages["40x"], err = newTemplate("40x", string(data))
		if err != nil {
			return nil, err
		}
	}
	if pages["50x"] == nil {
		data, err := embedPages.ReadFile("50x.gohtml")
		if err != nil {
			return nil, err
		}
		pages["50x"], err = newTemplate("50x", string(data))
		if err != nil {
			return nil, err
		}
	}
	return &ErrorPages{
		errorPages: pages,
	}, nil
}

func (p *ErrorPages) flushError(err error, request *http.Request, writer http.ResponseWriter) error {
	var code = http.StatusInternalServerError
	if errors.Is(err, ErrorNotMatches) {
		// 跳过不匹配
		return err
	} else if errors.Is(err, ErrorNotFound) {
		code = http.StatusNotFound
	} else {
		code = http.StatusInternalServerError
	}
	var metadata = &ErrorMetadata{
		StatusCode: code,
		Request:    request,
		Error:      err.Error(),
	}
	codeStr := strconv.Itoa(code)
	writer.Header().Add("Content-Type", "text/html;charset=utf-8")
	writer.WriteHeader(code)
	if result := p.errorPages[codeStr]; result != nil {
		return result.Execute(writer, metadata)
	}
	switch {
	case code >= 400 && code < 500:
		return p.errorPages["40x"].Execute(writer, metadata)
	case code >= 500:
		return p.errorPages["50x"].Execute(writer, metadata)
	default:
		return p.errorPages["50x"].Execute(writer, metadata)
	}
}
