package lib

import (
	"fmt"
)

type SceneValue struct {
	Sep        Edges
	Perm       bool // one-time cached if false
	WoundingUp bool // created during wounding up
}

func (s SceneValue) String() string {
	var added string
	if s.WoundingUp {
		added = fmt.Sprint("WoundingUp")

	}
	return fmt.Sprint("Sep ", s.Sep, "Perm: ", s.Perm) + added
}

// type Scene struct {
// 	Sub uint64
// 	Val SceneValue
// }

type Scene struct {
	Sub Edges
	Val SceneValue
}

type HashMap struct {
	internalMap map[uint64][]Scene
}

func (h *HashMap) Init() {

	h.internalMap = make(map[uint64][]Scene)
}

func (h HashMap) Len() int {
	return len(h.internalMap)
}

func (h *HashMap) Add(Sub Edges, Val SceneValue) {

	hash := Sub.Hash()

	val, ok := h.internalMap[hash]

	if !ok {
		// add new entry

		h.internalMap[hash] = []Scene{Scene{Sub: Sub, Val: Val}}
	} else {
		// log.Panicln("oh neos")
		_, ok := h.Check(Sub) // don't allow for the same entry to be overwritten

		if !ok {

			val = append(val, Scene{Sub: Sub, Val: Val})
			h.internalMap[hash] = val
		}

	}
}
func remove(s []Scene, i int) []Scene {
	s[i] = s[len(s)-1]
	// We do not need to put s[i] at the end, as it will be discarded anyway
	return s[:len(s)-1]
}

func (h *HashMap) Check(Sub Edges) (SceneValue, bool) {
	// check if sub appears in the hashpam

	hash := Sub.Hash()

	val, ok := h.internalMap[hash]

	if !ok {
		return SceneValue{}, false
	}

	for i := range val {
		if equalEdges(val[i].Sub, Sub) {

			if !val[i].Val.Perm { // delete one-time cached scene from map
				if len(val) == 1 {
					delete(h.internalMap, hash)
				} else {
					val = remove(val, i)
					h.internalMap[hash] = val
				}

			}

			return val[i].Val, true
		}
	}

	return SceneValue{}, false
}
