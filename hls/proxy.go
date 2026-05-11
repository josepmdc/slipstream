package hls

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/josepmdc/slipstream/config"
	"github.com/josepmdc/slipstream/hls/cache"
	"github.com/josepmdc/slipstream/lib/acestream"
)

type AceStreamClient interface {
	FetchManifest(ctx context.Context, aceID string) ([]byte, error)
	FetchSegment(ctx context.Context, segment string) ([]byte, error)
}

type SegmentCache interface {
	Get(ctx context.Context, url string, loader cache.SegmentLoaderFunc) ([]byte, error)
}

type ManifestCache interface {
	Get(ctx context.Context, aceID string, loader cache.ManifestLoaderFunc) ([]byte, error)
}

type Proxy struct {
	acestreamClient  AceStreamClient
	acestreamBaseURL string
	publicBaseURL    string
	segmentCache     SegmentCache
	manifestCache    ManifestCache
}

func NewProxy(cfg *config.Config, acestreamClient AceStreamClient, segmentCache SegmentCache, manifestCache ManifestCache) *Proxy {
	return &Proxy{
		acestreamClient:  acestreamClient,
		acestreamBaseURL: cfg.AceStreamBaseURL,
		publicBaseURL:    cfg.PublicBaseURL,
		segmentCache:     segmentCache,
		manifestCache:    manifestCache,
	}
}

func (proxy *Proxy) ServeManifest(w http.ResponseWriter, r *http.Request) {
	aceID := r.URL.Query().Get("id")

	if !acestream.IsValidAceID(aceID) {
		http.Error(w, "invalid acestream ID", http.StatusBadRequest)
		return
	}

	manifest, err := proxy.manifestCache.Get(r.Context(), aceID, proxy.acestreamClient.FetchManifest)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to fetch manifest: %s", err), http.StatusInternalServerError)
		return
	}

	rewritten := proxy.RewriteManifest(manifest)
	// TODO: we can prefetch segments to optimize UX

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Write(rewritten)
}

func (proxy *Proxy) RewriteManifest(raw []byte) []byte {
	return bytes.ReplaceAll(raw, []byte(proxy.acestreamBaseURL), []byte(proxy.publicBaseURL))
}

func (proxy *Proxy) ServeSegment(w http.ResponseWriter, r *http.Request) {
	data, err := proxy.segmentCache.Get(
		r.Context(),
		r.URL.Path,
		proxy.acestreamClient.FetchSegment,
	)
	if err != nil {
		http.Error(w, "upstream error", http.StatusBadGateway)
		return
	}

	w.Header().Set("Cache-Control", "max-age=300")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "video/mp2t")
	w.Write(data)
}
