package pages

import (
	"embed"
	"net/http"
	"strconv"
)

var (
	//go:embed 40x.html 50x.html
	embedPages embed.FS
)

type ErrorPages struct {
	errorPages map[string]string
}

func NewErrorPages(pages map[string]string) (*ErrorPages, error) {
	if pages == nil {
		pages = make(map[string]string)
	}
	if pages["40x"] == "" {
		data, err := embedPages.ReadFile("40x.html")
		if err != nil {
			return nil, err
		}
		pages["40x"] = string(data)
	}
	if pages["50x"] == "" {
		data, err := embedPages.ReadFile("50x.html")
		if err != nil {
			return nil, err
		}
		pages["50x"] = string(data)
	}
	return &ErrorPages{
		errorPages: pages,
	}, nil
}

func (p *ErrorPages) flushErrorPages(code int, writer http.ResponseWriter) error {
	codeStr := strconv.Itoa(code)
	if result := p.errorPages[codeStr]; result != "" {
		return flushPages(code, result, writer)
	}
	switch {
	case code >= 400 && code < 500:
		return flushPages(404, p.errorPages["40x"], writer)
	case code >= 500:
		return flushPages(502, p.errorPages["50x"], writer)
	default:
		return flushPages(502, p.errorPages["50x"], writer)
	}
}

func flushPages(code int, page string, writer http.ResponseWriter) error {
	writer.Header().Add("Content-Type", "text/html;charset=utf-8")
	writer.WriteHeader(code)
	_, err := writer.Write([]byte(page))
	return err
}
