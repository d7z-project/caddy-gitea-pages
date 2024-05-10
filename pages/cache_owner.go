package pages

import (
	"code.gitea.io/sdk/gitea"
	"context"
	"encoding/json"
	"github.com/allegro/bigcache/v3"
	"github.com/pkg/errors"
	"net/http"
	"strings"
	"sync"
	"time"
)

type OwnerCache struct {
	ttl time.Duration
	*bigcache.BigCache
	mutexes sync.Map
}

func NewOwnerCache(ttl time.Duration, cacheTtl time.Duration) OwnerCache {
	cache, _ := bigcache.New(context.Background(), bigcache.DefaultConfig(cacheTtl))
	return OwnerCache{
		ttl:      ttl,
		mutexes:  sync.Map{},
		BigCache: cache,
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
	raw, err := c.Get(owner)
	result := NewOwnerConfig()
	// 每固定时间刷新一次
	nextTime := time.Now().UnixMilli() - c.ttl.Milliseconds()
	if err == nil {
		err = json.Unmarshal(raw, result)
		if err != nil {
			return nil, err
		}
		if nextTime > result.FetchTime {
			err = bigcache.ErrEntryNotFound
		}
	}
	if errors.Is(err, bigcache.ErrEntryNotFound) {
		lock := c.Lock(owner)
		defer lock()
		//不存在缓存
		config, err := getOwner(giteaConfig, owner)
		if err != nil {
			return nil, errors.Wrap(err, "owner config not found")
		}
		marshal, err := json.Marshal(config)
		if err != nil {
			return nil, err
		}
		return config, c.Set(owner, marshal)
	} else if err != nil {
		return nil, err
	} else {
		return result, nil
	}
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
