package cache

import (
	"context"

	"github.com/maypok86/otter/v2"

	"github.com/josepmdc/slipstream/config"
)

type InMemorySegmentCache struct {
	cache *otter.Cache[string, []byte]
}

func NewInMemorySegmentCache(cfg *config.Config) *InMemorySegmentCache {
	c := otter.Must(&otter.Options[string, []byte]{
		MaximumSize:      cfg.SegmentCacheMaxSize,
		ExpiryCalculator: otter.ExpiryAccessing[string, []byte](cfg.SegmentCacheExpiration),
	})

	return &InMemorySegmentCache{c}
}

func (c *InMemorySegmentCache) Get(ctx context.Context, url string, loader SegmentLoaderFunc) ([]byte, error) {
	return c.cache.Get(ctx, url, otter.LoaderFunc[string, []byte](loader))
}

type InMemoryManifestCache struct {
	cache *otter.Cache[string, []byte]
}

func NewInMemoryManifestCache(cfg *config.Config) *InMemoryManifestCache {
	c := otter.Must(&otter.Options[string, []byte]{
		MaximumSize:      cfg.ManifestCacheMaxSize,
		ExpiryCalculator: otter.ExpiryWriting[string, []byte](cfg.ManifestCacheExpiration),
	})

	return &InMemoryManifestCache{c}
}

func (c *InMemoryManifestCache) Get(ctx context.Context, aceID string, loader ManifestLoaderFunc) ([]byte, error) {
	return c.cache.Get(ctx, aceID, otter.LoaderFunc[string, []byte](loader))
}
