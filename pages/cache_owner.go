package pages

import (
	"code.gitea.io/sdk/gitea"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"net/http"
	"strings"
	"sync"
	"time"
)

type OwnerCache struct {
	ttl time.Duration
	*cache.Cache
	mutexes sync.Map
}

func NewOwnerCache(ttl time.Duration, cacheTtl time.Duration) OwnerCache {
	return OwnerCache{
		ttl:     ttl,
		mutexes: sync.Map{},
		Cache:   cache.New(cacheTtl, cacheTtl*2),
	}
}

type OwnerConfig struct {
	FetchTime  int64           `json:"fetch_time,omitempty"`
	Repos      map[string]bool `json:"repos,omitempty"`
	LowerRepos map[string]bool `json:"lower_repos,omitempty"`
}

func NewOwnerConfig() *OwnerConfig {
	return &OwnerConfig{
		Repos:      make(map[string]bool),
		LowerRepos: make(map[string]bool),
	}
}

// 直接查询 Owner 信息
func getOwner(giteaConfig *GiteaConfig, owner string) (*OwnerConfig, error) {
	result := NewOwnerConfig()
	repos, resp, err := giteaConfig.Client.ListOrgRepos(owner, gitea.ListOrgReposOptions{
		ListOptions: gitea.ListOptions{
			PageSize: 999,
		},
	})
	if err != nil && resp.StatusCode == http.StatusNotFound {
		// 调用用户接口查询
		repos, resp, err = giteaConfig.Client.ListUserRepos(owner, gitea.ListReposOptions{
			ListOptions: gitea.ListOptions{
				PageSize: 999,
			},
		})
		if err != nil && resp.StatusCode == http.StatusNotFound {
			return nil, errors.Wrap(ErrorNotFound, err.Error())
		} else if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	for _, repo := range repos {
		result.Repos[repo.Name] = true
		result.LowerRepos[strings.ToLower(repo.Name)] = true
	}
	result.FetchTime = time.Now().UnixMilli()
	return result, nil
}

func (c *OwnerCache) GetOwnerConfig(giteaConfig *GiteaConfig, owner string) (*OwnerConfig, error) {
	raw, _ := c.Get(owner)
	// 每固定时间刷新一次
	nextTime := time.Now().UnixMilli() - c.ttl.Milliseconds()
	var result *OwnerConfig
	if raw != nil {
		result = raw.(*OwnerConfig)
		if nextTime > result.FetchTime {
			//移除旧数据
			c.Delete(owner)
			raw = nil
		}
	}
	if raw == nil {
		lock := c.Lock(owner)
		defer lock()
		if raw, find := c.Get(owner); find {
			return raw.(*OwnerConfig), nil
		}
		//不存在缓存
		var err error
		result, err = getOwner(giteaConfig, owner)
		if err != nil {
			return nil, errors.Wrap(err, "owner config not found")
		}
		c.Set(owner, result, cache.DefaultExpiration)
	}
	return result, nil
}

func (c *OwnerCache) Lock(any string) func() {
	value, _ := c.mutexes.LoadOrStore(any, &sync.Mutex{})
	mtx := value.(*sync.Mutex)
	mtx.Lock()

	return func() { mtx.Unlock() }
}

func (c *OwnerConfig) Exists(repo string) bool {
	return c.LowerRepos[strings.ToLower(repo)]
}
