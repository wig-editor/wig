package wig

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/pelletier/go-toml/v2"
)

type PositionEntry struct {
	Line      int   `toml:"line"`
	OpenCount int   `toml:"open_count"`
	Timestamp int64 `toml:"timestamp"`
}

type PositionCache struct {
	Files map[string]PositionEntry `toml:"files"`
}

func LoadPositionCache() *PositionCache {
	cache := &PositionCache{Files: make(map[string]PositionEntry)}
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".config", "wig", "position.toml")
	data, err := os.ReadFile(path)
	if err == nil {
		toml.Unmarshal(data, cache)
	}
	return cache
}

func (c *PositionCache) Save() {
	if len(c.Files) > 100 {
		type kv struct {
			Key string
			Val PositionEntry
		}
		var entries []kv
		for k, v := range c.Files {
			entries = append(entries, kv{k, v})
		}
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Val.OpenCount > entries[j].Val.OpenCount
		})
		if len(entries) > 100 {
			entries = entries[:100]
		}
		c.Files = make(map[string]PositionEntry)
		for _, e := range entries {
			c.Files[e.Key] = e.Val
		}
	}

	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "wig")
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, "position.toml")
	data, _ := toml.Marshal(c)
	os.WriteFile(path, data, 0644)
}
