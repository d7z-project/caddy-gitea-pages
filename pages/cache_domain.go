package pages

import (
	"bufio"
	"bytes"
	"code.gitea.io/sdk/gitea"
	"crypto/sha1"
	"fmt"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

type DomainCache struct {
	ttl time.Duration
	*cache.Cache
	mutexes sync.Map
}

func (c *DomainCache) Close() error {
	if c.Cache != nil {
		c.Cache.Flush()
	}
	return nil
}

type DomainConfig struct {
	FetchTime int64 //上次刷新时间

	PageDomain PageDomain
	Exists     bool         // 当前项目是否为 Pages
	FileCache  *cache.Cache // 文件缓存

	CNAME    []string        // 重定向地址
	SHA      string          // 缓存 SHA
	DATE     time.Time       // 文件提交时间
	BasePath string          // 根目录
	Topics   map[string]bool // 存储库标记

	Index    string //默认页面
	NotFound string //不存在页面
}

func (receiver *DomainConfig) Close() error {
	receiver.FileCache.Flush()
	return nil
}
func (receiver *DomainConfig) IsRoutePage() bool {
	return receiver.Topics["routes-history"] || receiver.Topics["routes-hash"]
}

func NewDomainCache(ttl time.Duration, refreshTtl time.Duration) DomainCache {
	c := cache.New(refreshTtl, 2*refreshTtl)
	c.OnEvicted(func(_ string, i interface{}) {
		config := i.(*DomainConfig)
		if config != nil {
			err := config.Close()
			if err != nil {
				return
			}
		}
	})
	return DomainCache{
		ttl:     ttl,
		Cache:   c,
		mutexes: sync.Map{},
	}
}

func fetch(client *GiteaConfig, domain *PageDomain, result *DomainConfig) error {
	branches, resp, err := client.Client.ListRepoBranches(domain.Owner, domain.Repo,
		gitea.ListRepoBranchesOptions{
			ListOptions: gitea.ListOptions{
				PageSize: 999,
			},
		})
	// 缓存 404 内容
	if resp != nil && resp.StatusCode >= 400 && resp.StatusCode < 500 {
		result.Exists = false
		return nil
	}
	if err != nil {
		return err
	}
	topics, resp, err := client.Client.ListRepoTopics(domain.Owner, domain.Repo, gitea.ListRepoTopicsOptions{
		ListOptions: gitea.ListOptions{
			PageSize: 999,
		},
	})
	if err != nil {
		return err
	}
	branchIndex := slices.IndexFunc(branches, func(x *gitea.Branch) bool { return x.Name == domain.Branch })
	if branchIndex == -1 {
		return errors.Wrap(ErrorNotFound, "branch not found")
	}
	currentSHA := branches[branchIndex].Commit.ID
	commitTime := branches[branchIndex].Commit.Timestamp
	result.Topics = make(map[string]bool)
	for _, topic := range topics {
		result.Topics[strings.ToLower(topic)] = true
	}
	if result.SHA == currentSHA {
		// 历史缓存一致，跳过
		result.FetchTime = time.Now().UnixMilli()
		return nil
	}
	// 清理历史缓存
	if result.SHA != currentSHA {
		if result.FileCache != nil {
			result.FileCache.Flush()
		}
		result.SHA = currentSHA
		result.DATE = commitTime
	}
	//查询是否为仓库
	result.Exists, err = client.FileExists(domain, result.BasePath+"/index.html")
	if err != nil {
		return err
	}
	if !result.Exists {
		return nil
	}
	result.Index = "index.html"
	//############# 处理 404
	if result.IsRoutePage() {
		result.NotFound = "/index.html"
	} else {
		notFound, err := client.FileExists(domain, result.BasePath+"/404.html")
		if err != nil {
			return err
		}
		if notFound {
			result.NotFound = "/404.html"
		}
	}
	// ############ 拉取 CNAME
	cname, err := client.ReadStringRepoFile(domain, "/CNAME")
	if err != nil && !errors.Is(err, ErrorNotFound) {
		// ignore not fond error
		return err
	} else if cname != "" {
		// 清理重定向
		result.CNAME = make([]string, 0)
		scanner := bufio.NewScanner(strings.NewReader(cname))
		for scanner.Scan() {
			alias := scanner.Text()
			alias = strings.TrimSpace(alias)
			alias = strings.TrimPrefix(strings.TrimPrefix(alias, "https://"), "http://")
			alias = strings.Split(alias, "/")[0]
			if len(strings.TrimSpace(alias)) > 0 {
				result.CNAME = append(result.CNAME, alias)
			}
		}
	}
	result.FetchTime = time.Now().UnixMilli()
	return nil
}

func (receiver *DomainConfig) tag(path string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(
		fmt.Sprintf("%s|%s|%s", receiver.SHA, receiver.PageDomain.Key(), path))))
}

func (receiver *DomainConfig) withNotFoundPage(
	client *GiteaConfig,
	response *FakeResponse,
) error {
	if receiver.NotFound == "" {
		// 没有默认页面
		return ErrorNotFound
	}
	notFound, _ := receiver.FileCache.Get(receiver.NotFound)
	if notFound == nil {
		// 不存在 notfound
		domain := &receiver.PageDomain
		fileContext, err := client.OpenFileContext(domain, receiver.BasePath+receiver.NotFound)
		if errors.Is(err, ErrorNotFound) {
			//缓存 not found 不存在
			receiver.FileCache.Set(receiver.NotFound, make([]byte, 0), cache.DefaultExpiration)
			return err
		} else if err != nil {
			return err
		}
		length, _ := strconv.Atoi(fileContext.Header.Get("Content-Length"))
		if length > client.CacheMaxSize {
			client.Logger.Debug("default page too large.")
			response.Body = fileContext.Body
			response.CacheModeIgnore()
		} else {
			// 保存缓存
			client.Logger.Debug("create default error page.")
			defer fileContext.Body.Close()
			defBuf, _ := io.ReadAll(fileContext.Body)
			receiver.FileCache.Set(receiver.NotFound, defBuf, cache.DefaultExpiration)
			response.Body = NewByteBuf(defBuf)
			response.CacheModeMiss()
		}
		response.ContentTypeExt(receiver.NotFound)
		response.Length(length)
	} else {
		notFound := notFound.([]byte)
		if len(notFound) == 0 {
			// 不存在 NotFound
			return ErrorNotFound
		}
		response.Length(len(notFound))
		response.Body = NewByteBuf(notFound)
		response.CacheModeHit()
	}
	client.Logger.Debug("use cache error page.")
	response.ContentTypeExt(receiver.NotFound)
	if receiver.IsRoutePage() {
		response.StatusCode = http.StatusOK
	} else {
		response.StatusCode = http.StatusNotFound
	}
	return nil
}

func (receiver *DomainConfig) getCachedData(
	client *GiteaConfig,
	path string,
) (*FakeResponse, error) {
	result := NewFakeResponse()
	if strings.HasSuffix(path, "/") {
		path = path + "index.html"
	}
	for k, v := range client.CustomHeaders {
		result.SetHeader(k, v)
	}
	result.ETag(receiver.tag(path))
	result.ContentTypeExt(path)
	cacheBuf, _ := receiver.FileCache.Get(path)
	// 使用缓存内容
	if cacheBuf != nil {
		cacheBuf := cacheBuf.([]byte)
		if len(cacheBuf) == 0 {
			//使用 NotFound 内容
			client.Logger.Debug("location not found ,", zap.Any("path", path))
			return result, receiver.withNotFoundPage(client, result)
		} else {
			// 使用缓存
			client.Logger.Debug("location use cache ,", zap.Any("path", path))
			result.Body = ByteBuf{
				bytes.NewBuffer(cacheBuf),
			}
			result.Length(len(cacheBuf))
			result.CacheModeHit()
			return result, nil
		}
	} else {
		// 添加缓存
		client.Logger.Debug("location add cache ,", zap.Any("path", path))
		domain := *(&receiver.PageDomain)
		domain.Branch = receiver.SHA
		fileContext, err := client.OpenFileContext(&domain, receiver.BasePath+path)
		if err != nil && !errors.Is(err, ErrorNotFound) {
			return nil, err
		} else if errors.Is(err, ErrorNotFound) {
			client.Logger.Debug("location not found and src not found,", zap.Any("path", path))
			// 不存在且源不存在
			receiver.FileCache.Set(path, make([]byte, 0), cache.DefaultExpiration)
			return result, receiver.withNotFoundPage(client, result)
		} else {
			// 源存在，执行缓存
			client.Logger.Debug("location found and set cache,", zap.Any("path", path))
			length, _ := strconv.Atoi(fileContext.Header.Get("Content-Length"))
			if length > client.CacheMaxSize {
				client.Logger.Debug("location too large , skip cache.", zap.Any("path", path))
				// 超过大小，回源
				result.Body = fileContext.Body
				result.Length(length)
				result.CacheModeIgnore()
				return result, nil
			} else {
				client.Logger.Debug("location saved,", zap.Any("path", path))
				// 未超过大小，缓存
				body, _ := io.ReadAll(fileContext.Body)
				receiver.FileCache.Set(path, body, cache.DefaultExpiration)
				result.Body = NewByteBuf(body)
				result.Length(len(body))
				result.CacheModeMiss()
				return result, nil
			}
		}
	}
}

func (receiver *DomainConfig) Copy(
	client *GiteaConfig,
	path string,
	writer http.ResponseWriter,
	_ *http.Request,
) (bool, error) {
	fakeResp, err := receiver.getCachedData(client, path)
	if err != nil {
		return false, err
	}
	for k, v := range fakeResp.Header {
		for _, s := range v {
			writer.Header().Add(k, s)
		}
	}
	writer.Header().Add("Pages-Server-Hash", receiver.SHA)
	writer.Header().Add("Last-Modified", receiver.DATE.UTC().Format(http.TimeFormat))
	writer.WriteHeader(fakeResp.StatusCode)
	defer fakeResp.Body.Close()
	_, err = io.Copy(writer, fakeResp.Body)
	return true, err
}

// FetchRepo 拉取 Repo 信息
func (c *DomainCache) FetchRepo(client *GiteaConfig, domain *PageDomain) (*DomainConfig, bool, error) {
	nextTime := time.Now().UnixMilli() - c.ttl.Milliseconds()
	lock := c.Lock(domain)
	defer lock()
	cacheKey := domain.Key()
	result, exists := c.Get(cacheKey)
	if !exists {
		result = &DomainConfig{
			PageDomain: *domain,
			FileCache:  cache.New(c.ttl, c.ttl*2),
		}
		if err := fetch(client, domain, result.(*DomainConfig)); err != nil {
			return nil, false, err
		}
		err := c.Add(cacheKey, result, cache.DefaultExpiration)
		if err != nil {
			return nil, false, err
		}
		return result.(*DomainConfig), false, nil
	} else {
		config := result.(*DomainConfig)
		if nextTime > config.FetchTime {
			// 刷新旧的缓存
			if err := fetch(client, domain, config); err != nil {
				return nil, false, err
			}
		}
		return config, true, nil
	}

}

func (c *DomainCache) Lock(any *PageDomain) func() {
	return c.LockAny(any.Key())
}

func (c *DomainCache) LockAny(any string) func() {
	value, _ := c.mutexes.LoadOrStore(any, &sync.Mutex{})
	mtx := value.(*sync.Mutex)
	mtx.Lock()

	return func() { mtx.Unlock() }
}
