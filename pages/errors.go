package pages

import (
	"github.com/pkg/errors"
	"io"
	"net/http"
	"os"
	"strings"
)

type ErrorPages struct {
	Error40xTemplate string
	Error50xTemplate string
}

func NewErrorPages(e40xPath, e50xPath string) (*ErrorPages, error) {
	var e40x = Default40xTemplate
	var e50x = Default50xTemplate
	var err error
	if e40xPath != "" {
		e40x, err = getTemplate(e40xPath)
		if err != nil {
			return nil, err
		}
	}
	if e50xPath != "" {
		e50x, err = getTemplate(e50xPath)
		if err != nil {
			return nil, err
		}
	}
	return &ErrorPages{
		Error40xTemplate: e40x,
		Error50xTemplate: e50x,
	}, nil
}

func getTemplate(path string) (string, error) {
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

func (p *ErrorPages) flush40xError(writer http.ResponseWriter) error {
	writer.Header().Add("Content-Type", "text/html;charset=utf-8")
	writer.WriteHeader(http.StatusNotFound)
	_, err := writer.Write([]byte(p.Error40xTemplate))
	return err
}

func (p *ErrorPages) flush50xError(writer http.ResponseWriter) error {
	writer.Header().Add("Content-Type", "text/html;charset=utf-8")
	writer.WriteHeader(http.StatusInternalServerError)
	_, err := writer.Write([]byte(p.Error50xTemplate))
	return err
}

const Default40xTemplate = `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
             <meta name="viewport" content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
                         <meta http-equiv="X-UA-Compatible" content="ie=edge">
             <title>404 Not Found</title>
</head>
<Body>
<div style="text-align: center;"><h1>404 Not Found</h1></div>
<hr><div style="text-align: center;">Caddy</div>
</Body>
</html>`
const Default50xTemplate = `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
             <meta name="viewport" content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
                         <meta http-equiv="X-UA-Compatible" content="ie=edge">
             <title>404 Not Found</title>
</head>
<Body>
<div style="text-align: center;"><h1>404 Not Found</h1></div>
<hr><div style="text-align: center;">Caddy</div>
</Body>
</html>`
