package pages

import (
	"encoding/json"
	"fmt"
	cmap "github.com/orcaman/concurrent-map/v2"
	"os"
	"strings"
	"sync"
)

var shared = cmap.New[PageDomain]()

type CustomDomains struct {
	/// 映射关系
	Alias *cmap.ConcurrentMap[string, PageDomain] `json:"DomainAlias"`
	/// 反向链接
	Reverse *cmap.ConcurrentMap[string, string] `json:"reverse"`
	/// 写锁
	Mutex sync.Mutex `json:"-"`
	/// 文件落盘
	Local string `json:"-"`
	// 是否全局共享
	Share bool
}

func (d *CustomDomains) Get(host string) (PageDomain, bool) {
	get, b := d.Alias.Get(strings.ToLower(host))
	if !b && d.Share {
		return shared.Get(strings.ToLower(host))
	}
	return get, b
}

func (d *CustomDomains) add(domain *PageDomain, alias string) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	key := strings.ToLower(domain.key())
	alias = strings.ToLower(alias)
	old, b := d.Reverse.Get(key)
	if b {
		// 移除旧的映射关系
		if d.Share {
			shared.Remove(old)
		}
		d.Alias.Remove(old)
		d.Reverse.Remove(key)
	}
	if d.Share {
		shared.Set(alias, *domain)
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

func NewCustomDomains(local string, share bool) (*CustomDomains, error) {
	if share {
		fmt.Printf("Global Alias Enabled.\n")
	}
	stat, err := os.Stat(local)
	alias := cmap.New[PageDomain]()
	reverse := cmap.New[string]()
	result := &CustomDomains{
		Alias:   &alias,
		Reverse: &reverse,
		Mutex:   sync.Mutex{},
		Local:   local,
		Share:   share,
	}
	fmt.Printf("Discover alias file :%s.\n", local)
	if local != "" && err == nil && !stat.IsDir() {
		bytes, err := os.ReadFile(local)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, result)
		fmt.Printf("Found %d Alias records.\n", result.Alias.Count())

		if err != nil {
			return nil, err
		}
		if share {
			for k, v := range result.Alias.Items() {
				shared.Set(k, v)
			}
		}
	}
	return result, nil
}
