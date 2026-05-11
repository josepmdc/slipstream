package cache

import "context"

type SegmentLoaderFunc func(ctx context.Context, url string) ([]byte, error)
type ManifestLoaderFunc func(ctx context.Context, aceID string) ([]byte, error)
