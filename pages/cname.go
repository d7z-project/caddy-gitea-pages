package pages

import (
	"encoding/json"
	"fmt"
	cmap "github.com/orcaman/concurrent-map/v2"
	"os"
	"strings"
	"sync"
)

type CustomDomains struct {
	/// 映射关系
	Alias *cmap.ConcurrentMap[string, PageDomain] `json:"DomainAlias"`
	/// 反向链接
	Reverse *cmap.ConcurrentMap[string, string] `json:"reverse"`
	/// 写锁
	Mutex sync.Mutex `json:"-"`
	/// 文件落盘
	Local string `json:"-"`
}

func (d *CustomDomains) Get(host string) (PageDomain, bool) {
	return d.Alias.Get(strings.ToLower(host))
}

func (d *CustomDomains) add(domain *PageDomain, alias string) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	key := strings.ToLower(domain.key())
	alias = strings.ToLower(alias)
	old, b := d.Reverse.Get(key)
	if b {
		// 移除旧的映射关系
		d.Alias.Remove(old)
		d.Reverse.Remove(key)
	}
	d.Alias.Set(alias, *domain)
	d.Reverse.Set(key, alias)
	if d.Local != "" {
		marshal, err := json.Marshal(d)
		err = os.WriteFile(d.Local, marshal, 0644)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
	}
}

func NewCustomDomains(local string) (*CustomDomains, error) {
	stat, err := os.Stat(local)
	alias := cmap.New[PageDomain]()
	reverse := cmap.New[string]()
	result := CustomDomains{
		Alias:   &alias,
		Reverse: &reverse,
		Mutex:   sync.Mutex{},
		Local:   local,
	}
	if local != "" && os.IsExist(err) && !stat.IsDir() {
		bytes, err := os.ReadFile(local)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, &result)
		if err != nil {
			return nil, err
		}
	}
	return &result, nil
}
