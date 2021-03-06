package lib

// cache.go implements a cache for hypergraph decomposition algorithms, loosely based on Samer and Gottlob 2009

import "sync"

// compCache stores the hashes of subgraphs for which a separator is known to have failed or succeeded
type compCache struct {
	Succ []uint64
	Fail []uint64
}

// Cache implements a caching mechanism for generic hypergraph decomposition algorithms
type Cache struct {
	cache    map[uint64]*compCache
	cacheMux *sync.RWMutex
	once     sync.Once
}

// CopyRef allows for safe copying of a cache by reference, not value
func (c *Cache) CopyRef(other *Cache) {
	if c.cache == nil { // to be sure only an initialised cache is copied
		c.Init()
	}
	c.cacheMux.RLock()
	defer c.cacheMux.RUnlock()

	other.cache = c.cache
	other.cacheMux = c.cacheMux
	other.once.Do(func() {}) // if cache is copied, it's assumed to already be initialised, so once is pre-fired here
}

// Reset will throw out all saved cache entries
func (c *Cache) Reset() {
	if c.cacheMux == nil {
		return // don't do anything if cache wasn't initialised yet
	}
	c.cacheMux.Lock()
	defer c.cacheMux.Unlock()

	c.cache = make(map[uint64]*compCache)

}

// Init needs to be called to initialise the cache
func (c *Cache) Init() {
	c.once.Do(c.initFunction)
}

func (c *Cache) initFunction() {
	if c.cache == nil {
		var newMutex sync.RWMutex
		c.cacheMux = &newMutex
		c.cache = make(map[uint64]*compCache)
	}
}

// Len returns the number of bindings in the cache
func (c *Cache) Len() int {
	c.cacheMux.RLock()
	defer c.cacheMux.RUnlock()

	return len(c.cache)
}

// AddPositive adds a separator sep and subgraph comp as a known successor case
// TODO: not really used and tested
func (c *Cache) AddPositive(sep Edges, comp Graph) {
	c.cacheMux.Lock()

	_, ok := c.cache[sep.Hash()]

	if !ok {
		var newCache compCache
		c.cache[sep.Hash()] = &newCache
	}

	c.cache[sep.Hash()].Succ = append(c.cache[sep.Hash()].Succ, comp.Hash())
	c.cacheMux.Unlock()
}

// AddNegative adds a separator sep and subgraph comp as a known failure case
func (c *Cache) AddNegative(sep Edges, comp Graph) {
	c.cacheMux.Lock()
	defer c.cacheMux.Unlock()

	_, ok := c.cache[sep.Hash()]
	if !ok {
		var newCache compCache
		c.cache[sep.Hash()] = &newCache
	}

	c.cache[sep.Hash()].Fail = append(c.cache[sep.Hash()].Fail, comp.Hash())
}

// CheckNegative checks for a separator sep and a subgraph whether it is a known failure case
func (c *Cache) CheckNegative(sep Edges, comps []Graph) bool {
	c.cacheMux.RLock()
	defer c.cacheMux.RUnlock()

	//check cache for previous encounters
	compCachePrev, ok := c.cache[sep.Hash()]

	if !ok { // sep not encountered before
		return false
	}

	for j := range comps {
		for i := range compCachePrev.Fail {
			if comps[j].Hash() == compCachePrev.Fail[i] {
				return true
			}
		}
	}

	return false
}

// CheckPositive checks for a separator sep and a subgraph whether it is a known successor case
// TODO: not really used and tested
func (c *Cache) CheckPositive(sep Edges, comps []Graph) bool {
	c.cacheMux.RLock()
	defer c.cacheMux.RUnlock()

	compCachePrev, ok := c.cache[sep.Hash()]

	if !ok { // sep not encountered before
		return false
	}

	for j := range comps {
		for i := range compCachePrev.Succ {
			if comps[j].Hash() == compCachePrev.Succ[i] {
				return true
			}
		}
	}

	return false
}
