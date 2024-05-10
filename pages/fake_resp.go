package pages

import (
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
)

type FakeResponse struct {
	*http.Response
}

func (r *FakeResponse) Length(len int) {
	r.ContentLength = int64(len)
	r.Header.Set("Content-Length", strconv.Itoa(len))
}

func (r *FakeResponse) CacheModeIgnore() {
	r.CacheMode("SKIP")
}
func (r *FakeResponse) CacheModeMiss() {
	r.CacheMode("MISS")
}
func (r *FakeResponse) CacheModeHit() {
	r.CacheMode("HIT")
}
func (r *FakeResponse) CacheMode(mode string) {
	r.SetHeader("Pages-Server-Cache", mode)
}
func (r *FakeResponse) ContentTypeExt(path string) {
	r.ContentType(mime.TypeByExtension(filepath.Ext(path)))
}
func (r *FakeResponse) ContentType(types string) {
	r.Header.Set("Content-Type", types)
}
func (r *FakeResponse) ETag(tag string) {
	r.Header.Set("ETag", fmt.Sprintf("\"%s\"", tag))
}

func NewFakeResponse() *FakeResponse {
	return &FakeResponse{
		&http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
		},
	}
}

func (r *FakeResponse) SetHeader(key string, value string) {
	r.Header.Set(key, value)
}
