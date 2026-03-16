package cdn

import (
	"fmt"
	"net/http"
	"time"
)

// Client communicates with the zencial-cdn service.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a new CDN client.
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// TriggerTranscode sends a POST request to the CDN service to start HLS transcoding.
func (c *Client) TriggerTranscode(videoID string) error {
	url := fmt.Sprintf("%s/api/v1/transcode/%s", c.baseURL, videoID)
	resp, err := c.httpClient.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("calling CDN transcode: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("CDN transcode returned status %d", resp.StatusCode)
	}
	return nil
}
