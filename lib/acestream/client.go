package acestream

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/google/uuid"

	"github.com/josepmdc/slipstream/config"
	"github.com/josepmdc/slipstream/lib/must"
)

type Client struct {
	streamInfoURL    string
	acestreamBaseURL string
	httpClient       *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		streamInfoURL: must.Do(
			url.JoinPath(cfg.AceStreamBaseURL, cfg.AceStreamEndpoint),
		),
		acestreamBaseURL: cfg.AceStreamBaseURL,
		httpClient:       &http.Client{Timeout: cfg.AceStreamTimeout},
	}
}

func (client *Client) FetchManifest(ctx context.Context, aceID string) ([]byte, error) {
	streamInfo, err := client.FetchStreamInfo(ctx, aceID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stream info: %w", err)
	}

	res, err := client.httpClient.Get(streamInfo.PlaybackURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch playback URL: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest's body: %w", err)
	}

	return body, nil
}

type StreamInfo struct {
	PlaybackURL       string `json:"playback_url"`
	StatURL           string `json:"stat_url"`
	CommandURL        string `json:"command_url"`
	Infohash          string `json:"infohash"`
	PlaybackSessionID string `json:"playback_session_id"`
	IsLive            int    `json:"is_live"`
	IsEncrypted       int    `json:"is_encrypted"`
	ClientSessionID   int    `json:"client_session_id"`
}

func (client *Client) FetchStreamInfo(ctx context.Context, aceID string) (*StreamInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, client.streamInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = url.Values{
		"id":     {aceID},
		"format": {"json"},
		"pid":    {uuid.NewString()},
	}.Encode()

	slog.Debug("making request to acestream", "url", req.URL.String())

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make http request to acestream: %w", err)
	}
	defer res.Body.Close()

	slog.Debug("stream response", "statusCode", res.StatusCode, "headers", res.Header)

	var envelope struct {
		Response StreamInfo `json:"response"`
		Error    string     `json:"error"`
	}

	if err := json.NewDecoder(res.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("failed to decode acestream response: %w", err)
	}

	if envelope.Error != "" {
		return nil, fmt.Errorf("acestream replied with an error: %s", envelope.Error)
	}

	slog.Debug("successfully recieved acestream response", "response", envelope.Response)

	return &envelope.Response, nil
}

func (client *Client) FetchSegment(ctx context.Context, segment string) ([]byte, error) {
	segmentURL, err := url.JoinPath(client.acestreamBaseURL, segment)
	if err != nil {
		return nil, fmt.Errorf("failed to form full acestream segment URL: %w", err)
	}

	resp, err := client.httpClient.Get(segmentURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch body, err := io.ReadAll(resp.Body); {
	case err != nil:
		return nil, fmt.Errorf("failed to read segment body: %w", err)
	case resp.StatusCode != http.StatusOK:
		slog.Error("acestream replied with a non-OK status code", "response", string(body))
		return nil, fmt.Errorf("acestream replied with a non-OK status code: %d", resp.StatusCode)
	default:
		return body, nil
	}
}
