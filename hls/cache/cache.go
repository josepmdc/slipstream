package cache

import "context"

type SegmentLoaderFunc func(ctx context.Context, url string) ([]byte, error)
