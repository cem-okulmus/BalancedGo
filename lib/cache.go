// implements a cache for hypergraph decomposition algorithms
package lib

import "sync"

type CompCache struct {
	Succ []uint64
	Fail []uint64
}

type Cache struct {
	Cache    map[uint64]*CompCache
	cacheMux sync.RWMutex
}

// needs to be called to initialise the cache
func (c *Cache) Init() {
	c.Cache = make(map[uint64]*CompCache)
}

func (c Cache) Len() int {
	return len(c.Cache)
}

func (c *Cache) AddPositive(sep Edges, comp Graph) {
	c.cacheMux.Lock()

	_, ok := c.Cache[sep.Hash()]

	if !ok {
		var newCache CompCache
		c.Cache[sep.Hash()] = &newCache
	}

	c.Cache[sep.Hash()].Succ = append(c.Cache[sep.Hash()].Succ, comp.Hash())
	c.cacheMux.Unlock()
}

func (c *Cache) AddNegative(sep Edges, comp Graph) {

	c.cacheMux.Lock()

	_, ok := c.Cache[sep.Hash()]

	if !ok {
		var newCache CompCache
		c.Cache[sep.Hash()] = &newCache
	}
	// fmt.Println("Addding negative, current length of cache", len(c.Cache))

	c.Cache[sep.Hash()].Fail = append(c.Cache[sep.Hash()].Fail, comp.Hash())
	c.cacheMux.Unlock()
}

// func (c *Cache) CheckNegative(sep Edges, comp Graph) bool {

// 	compCachePrev, _ := c.Cache[sep.Hash()]
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

	compCachePrev, ok := c.Cache[sep.Hash()]

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

	compCachePrev, _ := c.Cache[sep.Hash()]
	for i := range compCachePrev.Fail {
		if comp.Hash() == compCachePrev.Succ[i] {
			//  log.Println("Comp ", comp, " known as negative for sep ", sep)
			return true
		}

	}

	return false
}
