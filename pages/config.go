package pages

import (
	"context"
	"encoding/json"
	"github.com/allegro/bigcache/v3"
	"strings"
	"time"
)

type OwnerConfig struct {
	Repos      map[string]bool
	LowerRepos map[string]bool
}

func NewOwnerConfig() *OwnerConfig {
	return &OwnerConfig{
		Repos:      make(map[string]bool),
		LowerRepos: make(map[string]bool),
	}
}

type DomainConfig struct {
	RootPath     string
	NotFoundPath string
	Alias        string
}

type PageConfigGroup struct {
	locker      *DomainLocker
	ownerCache  *bigcache.BigCache
	domainCache *bigcache.BigCache
}

func (g *PageConfigGroup) GetOwnerCache(owner string) (*OwnerConfig, error) {
	result := &OwnerConfig{}
	raw, err := g.ownerCache.Get(strings.ToLower(owner))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(raw, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (g *PageConfigGroup) setOwnConfig(owner string, cache *OwnerConfig) error {
	marshal, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	return g.ownerCache.Set(owner, marshal)
}

func (g *PageConfigGroup) getDomainConfig(domain *PageDomain) (*DomainConfig, error) {
	result := &DomainConfig{}
	raw, err := g.domainCache.Get(strings.ToLower(domain.key()))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(raw, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (g *PageConfigGroup) setDomainConfig(domain *PageDomain, cache *DomainConfig) error {
	marshal, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	return g.domainCache.Set(domain.key(), marshal)
}

func NewDomainConfig() *PageConfigGroup {
	owner, _ := bigcache.New(context.Background(), bigcache.DefaultConfig(10*time.Minute))
	domain, _ := bigcache.New(context.Background(), bigcache.DefaultConfig(1*time.Minute))
	return &PageConfigGroup{
		locker:      NewDomainLocker(),
		ownerCache:  owner,
		domainCache: domain,
	}
}

func (c *OwnerConfig) Exists(repo string) bool {
	return c.LowerRepos[strings.ToLower(repo)]
}
