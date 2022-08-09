package sdk

import "sync"

/**
  @author: baisongyuan
  @since: 2022/7/28
**/

// 配置缓存

var configurationCache confCache

type configurationModels struct {
	MetaId  uint32
	Name    string
	Content string
	md5     string
}
type confCache struct {
	cache map[uint32]configurationModels
	lock  sync.RWMutex
}

func (c *confCache) Get(metaId uint32) configurationModels {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.cache[metaId]
}

func (c *confCache) AndOrUpdate(conf configurationModels)  {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cache[conf.MetaId] = conf
}

func (c *confCache)CheckIsSame(metaId uint32, md5 string) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if _, ok := c.cache[metaId]; !ok {
		return ok
	}
	return c.cache[metaId].md5 == md5
}

func init()  {
	configurationCache = confCache{
		cache: make(map[uint32]configurationModels),
	}
}
