package hls_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/josepmdc/slipstream/config"
	"github.com/josepmdc/slipstream/hls"
	"github.com/josepmdc/slipstream/hls/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAceStreamClient struct {
	manifest []byte
	segment  []byte
	err      error
}

func (m *mockAceStreamClient) FetchManifest(_ context.Context, _ string) ([]byte, error) {
	return m.manifest, m.err
}

func (m *mockAceStreamClient) FetchSegment(_ context.Context, _ string) ([]byte, error) {
	return m.segment, m.err
}

func newProxy(client hls.AceStreamClient) *hls.Proxy {
	cfg := &config.Config{
		AceStreamBaseURL:        "http://localhost:6878",
		PublicBaseURL:           "http://localhost:8080",
		SegmentCacheMaxSize:     100,
		SegmentCacheExpiration:  5 * time.Minute,
		ManifestCacheMaxSize:    100,
		ManifestCacheExpiration: 5 * time.Minute,
	}

	return hls.NewProxy(cfg, client, cache.NewInMemorySegmentCache(cfg), cache.NewInMemoryManifestCache(cfg))
}

const validAceID = "abc123def456abc123def456abc123def456abc1"

func TestRewriteManifest(t *testing.T) {
	t.Run("given a manifest with acestream URLs, rewrites them to the public proxy URL", func(t *testing.T) {
		proxy := newProxy(nil)
		raw := []byte("#EXTM3U\nhttp://localhost:6878/seg1.ts\nhttp://localhost:6878/seg2.ts\n")

		got := proxy.RewriteManifest(raw)

		assert.NotContains(t, string(got), "http://localhost:6878")
		assert.Contains(t, string(got), "http://localhost:8080/seg1.ts")
		assert.Contains(t, string(got), "http://localhost:8080/seg2.ts")
	})

	t.Run("given a manifest with no acestream URLs, returns it unchanged", func(t *testing.T) {
		proxy := newProxy(nil)
		raw := []byte("#EXTM3U\nhttp://acestream:8080/seg1.ts\nhttp://acestream:8080/seg2.ts\n")

		assert.Equal(t, raw, proxy.RewriteManifest(raw))
	})
}

func TestServeManifest(t *testing.T) {
	t.Run("given a valid ace ID, returns the rewritten manifest with correct content-type", func(t *testing.T) {
		raw := []byte("#EXTM3U\nhttp://localhost:6878/seg.ts\n")
		proxy := newProxy(&mockAceStreamClient{manifest: raw})

		r := httptest.NewRequest(http.MethodGet, "/hls/manifest.m3u8?id="+validAceID, nil)
		w := httptest.NewRecorder()
		proxy.ServeManifest(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/vnd.apple.mpegurl", w.Header().Get("Content-Type"))
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "#EXTM3U\nhttp://localhost:8080/seg.ts\n", w.Body.String())
	})

	t.Run("on cache hit, returns data without calling upstream", func(t *testing.T) {
		raw := []byte("original data that gets loaded into cache")
		upstream := &mockAceStreamClient{manifest: raw}
		proxy := newProxy(upstream)

		// first call, manifest is loaded into cache
		r := httptest.NewRequest(http.MethodGet, "/hls/manifest.m3u8?id="+validAceID, nil)
		proxy.ServeManifest(httptest.NewRecorder(), r)

		// we set another value just to make sure it's coming from cache and not from upstream
		upstream.manifest = []byte("another string to make sure this is not returned")

		w := httptest.NewRecorder()
		proxy.ServeManifest(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, []byte("original data that gets loaded into cache"), w.Body.Bytes())
	})

	t.Run("given no ace ID, returns 400", func(t *testing.T) {
		proxy := newProxy(&mockAceStreamClient{})

		r := httptest.NewRequest(http.MethodGet, "/hls/manifest.m3u8", nil)
		w := httptest.NewRecorder()
		proxy.ServeManifest(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("given an invalid ace ID, returns 400", func(t *testing.T) {
		proxy := newProxy(&mockAceStreamClient{})

		r := httptest.NewRequest(http.MethodGet, "/hls/manifest.m3u8?id=notvalid", nil)
		w := httptest.NewRecorder()
		proxy.ServeManifest(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("given the upstream fails, returns 500 with the error message", func(t *testing.T) {
		proxy := newProxy(&mockAceStreamClient{err: errors.New("timeout")})

		r := httptest.NewRequest(http.MethodGet, "/hls/manifest.m3u8?id="+validAceID, nil)
		w := httptest.NewRecorder()
		proxy.ServeManifest(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "timeout")
	})
}

func TestServeSegment(t *testing.T) {
	t.Run("on cache miss, fetches from upstream and returns data", func(t *testing.T) {
		data := []byte("fake-mpeg-ts-data")
		proxy := newProxy(&mockAceStreamClient{segment: data})

		r := httptest.NewRequest(http.MethodGet, "/seg1.ts", nil)
		w := httptest.NewRecorder()
		proxy.ServeSegment(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "video/mp2t", w.Header().Get("Content-Type"))
		assert.Equal(t, "max-age=300", w.Header().Get("Cache-Control"))
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, data, w.Body.Bytes())
	})

	t.Run("on cache hit, returns data without calling upstream", func(t *testing.T) {
		upstream := &mockAceStreamClient{
			segment: []byte("original data that gets loaded into cache"),
		}
		proxy := newProxy(upstream)

		// first call, segment is loaded into cache
		r := httptest.NewRequest(http.MethodGet, "/seg1.ts", nil)
		proxy.ServeSegment(httptest.NewRecorder(), r)

		// we set another value just to make sure it's coming from cache and not from upstream
		upstream.segment = []byte("another string to make sure this is not returned")

		w := httptest.NewRecorder()
		proxy.ServeSegment(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, []byte("original data that gets loaded into cache"), w.Body.Bytes())
	})

	t.Run("on upstream error, returns 502", func(t *testing.T) {
		proxy := newProxy(&mockAceStreamClient{err: errors.New("upstream down")})

		r := httptest.NewRequest(http.MethodGet, "/seg1.ts", nil)
		w := httptest.NewRecorder()
		proxy.ServeSegment(w, r)

		assert.Equal(t, http.StatusBadGateway, w.Code)
	})
}
