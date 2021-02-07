// implements a cache for hypergraph decomposition algorithms
package lib

import "sync"

type CompCache struct {
	Succ []uint64
	Fail []uint64
}

type Cache struct {
	cache    map[uint64]*CompCache
	cacheMux *sync.RWMutex
	once     sync.Once
}

// needs to be called to initialise the cache
func (c *Cache) Init() {
	c.once.Do(c.initFunction)
}

func (c *Cache) initFunction() {
	if c.cache == nil {
		var newMutex sync.RWMutex

		c.cacheMux = &newMutex

		c.cache = make(map[uint64]*CompCache)

	}
}

func (c *Cache) Len() int {
	c.cacheMux.Lock()
	defer c.cacheMux.Unlock()

	return len(c.cache)
}

func (c *Cache) AddPositive(sep Edges, comp Graph) {
	c.cacheMux.Lock()

	_, ok := c.cache[sep.Hash()]

	if !ok {
		var newCache CompCache
		c.cache[sep.Hash()] = &newCache
	}

	c.cache[sep.Hash()].Succ = append(c.cache[sep.Hash()].Succ, comp.Hash())
	c.cacheMux.Unlock()
}

func (c *Cache) AddNegative(sep Edges, comp Graph) {

	c.cacheMux.Lock()
	defer c.cacheMux.Unlock()

	_, ok := c.cache[sep.Hash()]

	if !ok {
		var newCache CompCache
		c.cache[sep.Hash()] = &newCache
	}
	// fmt.Println("Addding negative, current length of cache", len(c.cache))

	c.cache[sep.Hash()].Fail = append(c.cache[sep.Hash()].Fail, comp.Hash())

}

// func (c *Cache) CheckNegative(sep Edges, comp Graph) bool {

// 	compCachePrev, _ := c.cache[sep.Hash()]
// 	for i := range compCachePrev.Fail {
// 		if comp.Edges.Hash() == compCachePrev.Fail[i] {
// 			//  log.Println("Comp ", comp, "(hash ", comp.Edges.Hash(), ")  known as negative for sep ", sep)
// 			return true
// 		}

// 	}

// 	return false
// }

func (c *Cache) CheckNegative(sep Edges, comps []Graph) bool {

	c.cacheMux.RLock()
	defer c.cacheMux.RUnlock()

	//check chache for previous encounters

	compCachePrev, ok := c.cache[sep.Hash()]

	if !ok {
		return false
	} else {
		for j := range comps {
			for i := range compCachePrev.Fail {
				if comps[j].Hash() == compCachePrev.Fail[i] {
					//  log.Println("Comp ", comp, "(hash ", comp.Edges.Hash(), ")  known as negative for sep ", sep)

					return true
				}

			}

		}
	}

	return false
}

func (c *Cache) CheckPositive(sep Edges, comp Graph) bool {
	c.cacheMux.RLock()
	defer c.cacheMux.RUnlock()

	compCachePrev, _ := c.cache[sep.Hash()]
	for i := range compCachePrev.Fail {
		if comp.Hash() == compCachePrev.Succ[i] {
			//  log.Println("Comp ", comp, " known as negative for sep ", sep)
			return true
		}

	}

	return false
}
