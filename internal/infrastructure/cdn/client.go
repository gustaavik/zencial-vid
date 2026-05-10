// Package cdn is the VOD API's client for the zencial-cdn service.
//
// The CDN sits between the frontend and the underlying object store. The API
// mints short-lived HMAC-signed URLs that authorize the browser (or this API
// itself, when proxying small uploads) to PUT bytes into S3 via the CDN. The
// CDN never trusts callers — every write request is verified against the
// signature we generate here.
package cdn

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Operation labels are part of the canonical signature input. The CDN uses
// them to match a signed URL against the route it landed on, so a token that
// authorizes a thumbnail upload cannot be replayed against the video upload
// route (and vice-versa).
const (
	OpVideoUpload     = "video-upload"
	OpThumbnailUpload = "thumbnail-upload"
)

// Default expiry for upload URLs handed to the browser. Long enough for slow
// connections to finish a multi-GB video, short enough to bound replay risk.
const DefaultUploadExpiry = 30 * time.Minute

// Client communicates with the zencial-cdn service.
type Client struct {
	publicBaseURL   string // host the browser sees (returned in API responses)
	internalBaseURL string // host this process uses for outbound HTTP to the CDN
	signingKey      []byte
	keyID           string
	httpClient      *http.Client
}

// New creates a new CDN client. publicBaseURL is the browser-facing URL of
// the CDN (e.g. https://cdn.example.com). internalBaseURL is the URL this
// API uses to talk to the CDN over a private network (e.g. http://cdn:8090);
// it falls back to publicBaseURL when empty. signingKey and keyID must match
// the CDN's CDN_UPLOAD_SIGNING_KEY/CDN_UPLOAD_SIGNING_KEY_ID configuration.
func New(publicBaseURL, internalBaseURL, signingKey, keyID string) *Client {
	if internalBaseURL == "" {
		internalBaseURL = publicBaseURL
	}
	return &Client{
		publicBaseURL:   strings.TrimRight(publicBaseURL, "/"),
		internalBaseURL: strings.TrimRight(internalBaseURL, "/"),
		signingKey:      []byte(signingKey),
		keyID:           keyID,
		httpClient:      &http.Client{Timeout: 30 * time.Minute},
	}
}

// PublicBaseURL returns the browser-facing CDN base URL.
func (c *Client) PublicBaseURL() string { return c.publicBaseURL }

// HasSigningKey reports whether a non-empty signing secret was configured.
// Callers can use this to fail fast at startup instead of generating bogus
// URLs at runtime.
func (c *Client) HasSigningKey() bool { return len(c.signingKey) > 0 }

// TriggerTranscode sends a POST request to the CDN service to start HLS transcoding.
func (c *Client) TriggerTranscode(videoID string) error {
	u := c.internalBaseURL + "/api/v1/transcode/" + videoID
	resp, err := c.httpClient.Post(u, "application/json", nil)
	if err != nil {
		return fmt.Errorf("calling CDN transcode: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("CDN transcode returned status %d", resp.StatusCode)
	}
	return nil
}

// SignVideoUploadURL returns a fully-formed, signed PUT URL on the public
// CDN host the browser uses to upload a video binary directly. videoID and
// filename together determine the resulting S3 key (videos/{id}/{filename}).
func (c *Client) SignVideoUploadURL(videoID, filename string, expiry time.Duration) (string, time.Time, error) {
	if !c.HasSigningKey() {
		return "", time.Time{}, errors.New("cdn: no signing key configured")
	}
	expiresAt := time.Now().UTC().Add(expiry)
	path := "/api/v1/uploads/videos/" + videoID + "/" + filename
	return c.buildSignedURL(c.publicBaseURL, OpVideoUpload, path, expiresAt), expiresAt, nil
}

// UploadThumbnail proxies a thumbnail body from this API to the CDN over the
// internal network. videoID identifies the video; ext is the file extension
// including the leading dot (e.g. ".jpg"). The body is streamed — never
// buffered in memory — and the resulting S3 object key is returned.
func (c *Client) UploadThumbnail(ctx context.Context, videoID, ext, contentType string, body io.Reader) (string, error) {
	if !c.HasSigningKey() {
		return "", errors.New("cdn: no signing key configured")
	}
	if ext == "" || !strings.HasPrefix(ext, ".") {
		return "", fmt.Errorf("cdn: extension must start with '.' (got %q)", ext)
	}
	filename := "thumbnail" + ext
	path := "/api/v1/uploads/thumbnails/" + videoID + "/" + filename

	expiresAt := time.Now().UTC().Add(5 * time.Minute)
	signedURL := c.buildSignedURL(c.internalBaseURL, OpThumbnailUpload, path, expiresAt)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, signedURL, body)
	if err != nil {
		return "", fmt.Errorf("building thumbnail PUT: %w", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("uploading thumbnail to CDN: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 300 {
		// Best-effort body capture for diagnostics. Bound the read so a
		// misconfigured CDN can't pull our memory off a cliff.
		buf, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("CDN thumbnail upload returned status %d: %s", resp.StatusCode, string(buf))
	}

	// The CDN returns {"object_key": "videos/<id>/thumbnail.<ext>", ...}.
	// We construct the same key locally rather than parsing JSON — the CDN
	// uses a deterministic mapping and parsing is unnecessary work.
	return "videos/" + videoID + "/" + filename, nil
}

// ThumbnailURL returns the public-facing URL the browser should fetch to
// display the thumbnail for videoID. Reads are unauthenticated — the CDN
// streams whatever object lives at videos/{videoID}/thumbnail.*.
func (c *Client) ThumbnailURL(videoID string) string {
	return c.publicBaseURL + "/api/v1/thumbnails/" + videoID
}

// buildSignedURL is the canonical URL builder. It is the single place the
// HMAC is constructed; SignUploadURL and UploadThumbnail both go through it
// so the CDN sees identically-formatted URLs from both call sites.
func (c *Client) buildSignedURL(host, op, path string, expiresAt time.Time) string {
	exp := expiresAt.Unix()
	sig := signCanonical(c.signingKey, op, path, c.keyID, exp)

	q := url.Values{}
	q.Set("op", op)
	q.Set("exp", fmt.Sprintf("%d", exp))
	q.Set("keyId", c.keyID)
	q.Set("sig", sig)
	return host + path + "?" + q.Encode()
}

// signCanonical is byte-for-byte compatible with the CDN's auth.Sign. Both
// implementations must produce the same bytes for the same inputs; the CDN
// repo has its own copy under internal/auth and tests cross-validate.
func signCanonical(secret []byte, op, path, keyID string, exp int64) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = fmt.Fprintf(mac, "%s|%s|%d|%s", op, path, exp, keyID)
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
