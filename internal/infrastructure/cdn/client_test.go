package cdn

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

// reverifyURL re-implements the CDN's signature check to cross-validate that
// the API and CDN agree on the canonical form. If this test ever fails after
// editing client.go, the CDN's auth.Verify must be updated in lockstep.
func reverifyURL(t *testing.T, secret []byte, op, signedURL string) {
	t.Helper()
	u, err := url.Parse(signedURL)
	if err != nil {
		t.Fatalf("parse signed URL: %v", err)
	}
	q := u.Query()

	gotOp := q.Get("op")
	if gotOp != op {
		t.Fatalf("op = %q, want %q", gotOp, op)
	}
	expStr := q.Get("exp")
	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		t.Fatalf("invalid exp %q: %v", expStr, err)
	}
	if time.Now().Unix() > exp {
		t.Fatalf("token already expired")
	}
	keyID := q.Get("keyId")

	mac := hmac.New(sha256.New, secret)
	_, _ = fmt.Fprintf(mac, "%s|%s|%d|%s", op, u.Path, exp, keyID)
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	got, err := base64.RawURLEncoding.DecodeString(q.Get("sig"))
	if err != nil {
		t.Fatalf("sig is not valid base64: %v", err)
	}
	wantBytes, _ := base64.RawURLEncoding.DecodeString(expected)
	if !hmac.Equal(got, wantBytes) {
		t.Fatalf("signature mismatch — API and CDN canonicalization disagree")
	}
}

func TestSignVideoUploadURL_RoundTrips(t *testing.T) {
	t.Parallel()

	const secret = "super-secret-shared-key"
	c := New("https://cdn.example.test", "", secret, "v1")

	signed, expiresAt, err := c.SignVideoUploadURL("00000000-0000-0000-0000-000000000001", "original.mp4", 5*time.Minute)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if !strings.HasPrefix(signed, "https://cdn.example.test/api/v1/uploads/videos/") {
		t.Fatalf("unexpected URL host or path: %s", signed)
	}
	if time.Until(expiresAt) <= 0 || time.Until(expiresAt) > 5*time.Minute+time.Second {
		t.Fatalf("unexpected expiry: %v", expiresAt)
	}
	reverifyURL(t, []byte(secret), OpVideoUpload, signed)
}

func TestSignVideoUploadURL_NoKeyConfigured(t *testing.T) {
	t.Parallel()

	c := New("https://cdn.example.test", "", "", "v1")

	_, _, err := c.SignVideoUploadURL("vid", "original.mp4", time.Minute)
	if err == nil {
		t.Fatal("expected error when signing key is unset")
	}
}

func TestThumbnailURL(t *testing.T) {
	t.Parallel()

	c := New("https://cdn.example.test/", "", "secret", "v1") // trailing slash should be stripped
	got := c.ThumbnailURL("abc-123")
	want := "https://cdn.example.test/api/v1/thumbnails/abc-123"
	if got != want {
		t.Fatalf("ThumbnailURL = %q, want %q", got, want)
	}
}

func TestUploadThumbnail_StreamsToInternalURL(t *testing.T) {
	t.Parallel()

	const secret = "secret"
	body := []byte("\xff\xd8\xff\xe0fake-jpeg")
	var (
		gotMethod      string
		gotPath        string
		gotContentType string
		gotBody        []byte
		gotQuery       url.Values
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		gotQuery = r.URL.Query()
		buf, _ := io.ReadAll(r.Body)
		gotBody = buf
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"object_key":"videos/v1/thumbnail.jpg"}`))
	}))
	defer srv.Close()

	// publicBaseURL points at a wrong host on purpose — UploadThumbnail must
	// hit the *internal* URL so we know the routing is correct.
	c := New("https://wrong.example", srv.URL, secret, "v1")

	key, err := c.UploadThumbnail(context.Background(), "vid-1", ".jpg", "image/jpeg", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("UploadThumbnail: %v", err)
	}
	if key != "videos/vid-1/thumbnail.jpg" {
		t.Fatalf("key = %q", key)
	}
	if gotMethod != http.MethodPut {
		t.Fatalf("method = %q", gotMethod)
	}
	if gotPath != "/api/v1/uploads/thumbnails/vid-1/thumbnail.jpg" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotContentType != "image/jpeg" {
		t.Fatalf("content-type = %q", gotContentType)
	}
	if !bytes.Equal(gotBody, body) {
		t.Fatalf("body mismatch")
	}
	if gotQuery.Get("op") != OpThumbnailUpload {
		t.Fatalf("op = %q", gotQuery.Get("op"))
	}
	if gotQuery.Get("sig") == "" {
		t.Fatal("missing sig")
	}

	// And the URL the CDN sees must verify against the same secret.
	reverifyURL(t, []byte(secret), OpThumbnailUpload, srv.URL+gotPath+"?"+gotQuery.Encode())
}

func TestUploadThumbnail_ErrorOnNonSuccessStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := New("", srv.URL, "secret", "v1")
	_, err := c.UploadThumbnail(context.Background(), "vid", ".jpg", "image/jpeg", strings.NewReader("x"))
	if err == nil {
		t.Fatal("expected error on 401 from CDN")
	}
}

func TestUploadThumbnail_RequiresLeadingDotExt(t *testing.T) {
	t.Parallel()

	c := New("", "https://internal", "secret", "v1")
	_, err := c.UploadThumbnail(context.Background(), "vid", "jpg", "image/jpeg", strings.NewReader("x"))
	if err == nil {
		t.Fatal("expected error for extension without leading dot")
	}
}
