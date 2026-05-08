package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/josepmdc/slipstream/config"
	"github.com/josepmdc/slipstream/lib/acestream"
	"github.com/josepmdc/slipstream/lib/must"
)

func main() {
	cfg := must.Do(config.Load())

	acestreamClient := acestream.NewClient(cfg)
	response := must.Do(
		acestreamClient.FetchStreamInfo(context.Background(), "895d08633bb22c2573281655cf3ca44de476cc73"),
	)

	fmt.Printf("->> ->> ->> %s\n", must.Do(json.MarshalIndent(response, "	", "")))
}
