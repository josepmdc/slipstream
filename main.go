package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/josepmdc/slipstream/api"
	"github.com/josepmdc/slipstream/api/middleware"
	"github.com/josepmdc/slipstream/config"
	"github.com/josepmdc/slipstream/hls"
	"github.com/josepmdc/slipstream/hls/cache"
	"github.com/josepmdc/slipstream/lib/acestream"
	"github.com/josepmdc/slipstream/lib/must"
)

func main() {
	cfg := must.Do(config.Load())

	acestreamClient := acestream.NewClient(cfg)

	segmentCache := cache.NewInMemorySegmentCache(cfg)
	manifestCache := cache.NewInMemoryManifestCache(cfg)

	hls := hls.NewProxy(cfg, acestreamClient, segmentCache, manifestCache)

	router := api.NewRouter()

	router.Use(middleware.Logging)

	router.HandleFunc("GET /check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(must.Do(json.Marshal(`{"status": "ok"}`)))
	})
	router.HandleFunc("GET /ace/manifest.m3u8", hls.ServeManifest)
	router.HandleFunc("GET /ace/c/{infohash}/{segment}", hls.ServeSegment)

	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	slog.Info("starting server", "address", address)

	log.Fatalf("failed to start HTTP server: %s", http.ListenAndServe(address, router))
}
