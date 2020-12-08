// Package IConsistent provides a IConsistent hashing function.
//
// IConsistent hashing is often used to distribute requests to a changing set of servers.  For example,
// say you have some cache servers cacheA, cacheB, and cacheC.  You want to decide which cache server
// to use to look up information on a user.
//
// You could use a typical hash table and hash the user id
// to one of cacheA, cacheB, or cacheC.  But with a typical hash table, if you add or remove a server,
// almost all keys will get remapped to different results, which basically could bring your service
// to a grinding halt while the caches get rebuilt.
//
// With a IConsistent hash, adding or removing a server drastically reduces the number of keys that
// get remapped.
//
// Read more about IConsistent hashing on wikipedia:  http://en.wikipedia.org/wiki/Consistent_hashing
//

package types

import (
	"fmt"
	"hash/crc32"
	"hash/fnv"
	"sort"
	"strconv"
	"sync"
)

// IConsistent holds the information about the members of the IConsistent hash circle.
type IConsistent struct {
	circle           map[uint32]Node
	members          map[string]Node
	sortedHashes     uints
	NumberOfReplicas int
	count            int64
	scratch          [64]byte
	UseFnv           bool
	sync.RWMutex
}

type Node interface {
	String() string
}

// NewIConsistent creates a new IConsistent object with a default setting of 20 replicas for each entry.
//
// To change the number of replicas, set NumberOfReplicas before adding entries.
func NewIConsistent() *IConsistent {
	c := new(IConsistent)
	c.NumberOfReplicas = 1237
	c.circle = make(map[uint32]Node)
	c.members = make(map[string]Node)
	return c
}

// eltKey generates a string key for an element with an index.
func (c *IConsistent) eltKey(elt Node, idx int) string {
	// return elt + "|" + strconv.Itoa(idx)
	return strconv.Itoa(idx) + elt.String()
}

// Add inserts a string element in the IConsistent hash.
func (c *IConsistent) Add(elt Node) {
	c.Lock()
	defer c.Unlock()
	c.add(elt)
	fmt.Println(len(c.circle))
}

// need c.Lock() before calling
func (c *IConsistent) add(elt Node) {
	for i := 0; i < c.NumberOfReplicas; i++ {
		c.circle[c.hashKey(c.eltKey(elt, i))] = elt
	}
	c.members[elt.String()] = elt
	c.updateSortedHashes()
	c.count++
}

// Remove removes an element from the hash.
func (c *IConsistent) Remove(elt Node) {
	c.Lock()
	defer c.Unlock()
	c.remove(elt)
}

func (c *IConsistent) RemoveByKey(key string) Node {
	c.Lock()
	defer c.Unlock()
	if node, ok := c.members[key]; ok {
		c.remove(node)
		return node
	}
	return nil
}

// need c.Lock() before calling
func (c *IConsistent) remove(elt Node) {
	for i := 0; i < c.NumberOfReplicas; i++ {
		delete(c.circle, c.hashKey(c.eltKey(elt, i)))
	}
	delete(c.members, elt.String())
	c.updateSortedHashes()
	c.count--
}

func (c *IConsistent) Members() []Node {
	c.RLock()
	defer c.RUnlock()
	var m []Node
	for _, k := range c.members {
		m = append(m, k)
	}
	return m
}

// Get returns an element close to where name hashes to in the circle.
func (c *IConsistent) Get(name string) (Node, error) {
	c.RLock()
	defer c.RUnlock()
	if len(c.circle) == 0 {
		return nil, ErrEmptyCircle
	}
	key := c.hashKey(name)
	i := c.search(key)
	return c.circle[c.sortedHashes[i]], nil
}

func (c *IConsistent) search(key uint32) (i int) {
	f := func(x int) bool {
		return c.sortedHashes[x] > key
	}
	i = sort.Search(len(c.sortedHashes), f)
	if i >= len(c.sortedHashes) {
		i = 0
	}
	return
}

func (c *IConsistent) hashKey(key string) uint32 {
	if c.UseFnv {
		return c.hashKeyFnv(key)
	}
	return c.hashKeyCRC32(key)
}

func (c *IConsistent) hashKeyCRC32(key string) uint32 {
	if len(key) < 64 {
		var scratch [64]byte
		copy(scratch[:], key)
		return crc32.ChecksumIEEE(scratch[:len(key)])
	}
	return crc32.ChecksumIEEE([]byte(key))
}

func (c *IConsistent) hashKeyFnv(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func (c *IConsistent) updateSortedHashes() {
	hashes := c.sortedHashes[:0]
	// reallocate if we're holding on to too much (1/4th)
	if cap(c.sortedHashes)/(c.NumberOfReplicas*4) > len(c.circle) {
		hashes = nil
	}
	for k := range c.circle {
		hashes = append(hashes, k)
	}
	sort.Sort(hashes)
	c.sortedHashes = hashes
}
