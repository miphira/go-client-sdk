package sdk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"
)

// Test constants - these are NOT real credentials, only used for unit testing
const (
	testBaseURL   = "https://storage.example.com"
	testAccessKey = "MOS_UNIT_TEST_FAKE_KEY" // #nosec - fake test value
	testSecretKey = "unit_test_fake_secret"  // #nosec - fake test value
)

func TestNewClient(t *testing.T) {
	client := NewClient(testBaseURL, testAccessKey, testSecretKey)

	if client.BaseURL != testBaseURL {
		t.Errorf("expected BaseURL %s, got %s", testBaseURL, client.BaseURL)
	}
	if client.AccessKey != testAccessKey {
		t.Errorf("expected AccessKey %s, got %s", testAccessKey, client.AccessKey)
	}
	if client.SecretKey != testSecretKey {
		t.Errorf("expected SecretKey %s, got %s", testSecretKey, client.SecretKey)
	}
}

func TestGenerateSignature(t *testing.T) {
	client := NewClient(testBaseURL, testAccessKey, testSecretKey)

	method := "GET"
	path := "/api/v1/projects/test-project/buckets/test-bucket/objects/test.jpg"
	expires := int64(1735344000)

	signature := client.GenerateSignature(method, path, expires)

	// Verify signature is not empty
	if signature == "" {
		t.Error("signature should not be empty")
	}

	// Verify signature is base64 URL encoded
	_, err := base64.URLEncoding.DecodeString(signature)
	if err != nil {
		t.Errorf("signature should be valid base64 URL encoding: %v", err)
	}

	// Verify signature is deterministic
	signature2 := client.GenerateSignature(method, path, expires)
	if signature != signature2 {
		t.Error("signature should be deterministic for same inputs")
	}

	// Verify signature changes with different inputs
	signature3 := client.GenerateSignature("POST", path, expires)
	if signature == signature3 {
		t.Error("signature should differ for different HTTP methods")
	}
}

func TestGenerateSignature_MatchesAlgorithm(t *testing.T) {
	client := NewClient(testBaseURL, testAccessKey, testSecretKey)

	method := "GET"
	path := "/api/v1/projects/uuid/buckets/images/objects/photo.jpg"
	expires := int64(1735344000)

	// Manually compute expected signature
	stringToSign := fmt.Sprintf("%s\n%s\n%d", method, path, expires)
	h := hmac.New(sha256.New, []byte(testSecretKey))
	h.Write([]byte(stringToSign))
	expected := base64.URLEncoding.EncodeToString(h.Sum(nil))

	actual := client.GenerateSignature(method, path, expires)

	if actual != expected {
		t.Errorf("signature mismatch: expected %s, got %s", expected, actual)
	}
}

func TestGeneratePresignedURL(t *testing.T) {
	client := NewClient(testBaseURL, testAccessKey, testSecretKey)

	path := "/api/v1/projects/test-project/buckets/test-bucket/objects/test.jpg"
	expiresIn := time.Hour

	presignedURL := client.GeneratePresignedURL("GET", path, expiresIn)

	// Verify URL starts with base URL
	if !strings.HasPrefix(presignedURL, testBaseURL) {
		t.Errorf("URL should start with base URL: %s", presignedURL)
	}

	// Verify URL contains path
	if !strings.Contains(presignedURL, path) {
		t.Errorf("URL should contain path: %s", presignedURL)
	}

	// Verify URL contains required query parameters
	if !strings.Contains(presignedURL, "X-Mos-AccessKey="+testAccessKey) {
		t.Errorf("URL should contain X-Mos-AccessKey: %s", presignedURL)
	}
	if !strings.Contains(presignedURL, "X-Mos-Expires=") {
		t.Errorf("URL should contain X-Mos-Expires: %s", presignedURL)
	}
	if !strings.Contains(presignedURL, "X-Mos-Signature=") {
		t.Errorf("URL should contain X-Mos-Signature: %s", presignedURL)
	}
}

func TestGeneratePresignedURL_ValidURL(t *testing.T) {
	client := NewClient(testBaseURL, testAccessKey, testSecretKey)

	path := "/api/v1/projects/test-project/buckets/test-bucket/objects/test.jpg"
	presignedURL := client.GeneratePresignedURL("GET", path, time.Hour)

	// Verify it's a valid URL
	parsed, err := url.Parse(presignedURL)
	if err != nil {
		t.Errorf("generated URL should be valid: %v", err)
	}

	// Verify query parameters can be parsed
	query := parsed.Query()
	if query.Get("X-Mos-AccessKey") != testAccessKey {
		t.Errorf("X-Mos-AccessKey mismatch: got %s", query.Get("X-Mos-AccessKey"))
	}
	if query.Get("X-Mos-Expires") == "" {
		t.Error("X-Mos-Expires should not be empty")
	}
	if query.Get("X-Mos-Signature") == "" {
		t.Error("X-Mos-Signature should not be empty")
	}
}

func TestGetObjectURL(t *testing.T) {
	client := NewClient(testBaseURL, testAccessKey, testSecretKey)

	projectID := "550e8400-e29b-41d4-a716-446655440000"
	bucketName := "images"
	filename := "photo.jpg"

	resultURL := client.GetObjectURL(projectID, bucketName, filename, time.Hour)

	expectedPath := fmt.Sprintf("/api/v1/projects/%s/buckets/%s/objects/%s", projectID, bucketName, filename)
	if !strings.Contains(resultURL, expectedPath) {
		t.Errorf("URL should contain correct path: %s", resultURL)
	}
}

func TestUploadObjectURL(t *testing.T) {
	client := NewClient(testBaseURL, testAccessKey, testSecretKey)

	projectID := "550e8400-e29b-41d4-a716-446655440000"
	bucketName := "images"

	resultURL := client.UploadObjectURL(projectID, bucketName, time.Hour)

	expectedPath := fmt.Sprintf("/api/v1/projects/%s/buckets/%s/objects", projectID, bucketName)
	if !strings.Contains(resultURL, expectedPath) {
		t.Errorf("URL should contain correct path: %s", resultURL)
	}

	// Verify it doesn't contain filename (upload URL doesn't have one)
	if strings.HasSuffix(strings.Split(resultURL, "?")[0], ".jpg") {
		t.Error("upload URL should not have filename in path")
	}
}

func TestDeleteObjectURL(t *testing.T) {
	client := NewClient(testBaseURL, testAccessKey, testSecretKey)

	projectID := "550e8400-e29b-41d4-a716-446655440000"
	bucketName := "images"
	filename := "photo.jpg"

	resultURL := client.DeleteObjectURL(projectID, bucketName, filename, time.Hour)

	expectedPath := fmt.Sprintf("/api/v1/projects/%s/buckets/%s/objects/%s", projectID, bucketName, filename)
	if !strings.Contains(resultURL, expectedPath) {
		t.Errorf("URL should contain correct path: %s", resultURL)
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.ExpiresIn != time.Hour {
		t.Errorf("default ExpiresIn should be 1 hour, got %v", opts.ExpiresIn)
	}
}

func TestExpiryTime(t *testing.T) {
	client := NewClient(testBaseURL, testAccessKey, testSecretKey)

	path := "/api/v1/projects/test/buckets/test/objects/test.jpg"

	// Generate URL with 1 hour expiry
	beforeGen := time.Now().Unix()
	presignedURL := client.GeneratePresignedURL("GET", path, time.Hour)
	afterGen := time.Now().Unix()

	// Parse the expires value from URL
	parsed, _ := url.Parse(presignedURL)
	expiresStr := parsed.Query().Get("X-Mos-Expires")

	var expires int64
	fmt.Sscanf(expiresStr, "%d", &expires)

	// Verify expires is approximately 1 hour in the future
	expectedMin := beforeGen + 3600
	expectedMax := afterGen + 3600

	if expires < expectedMin || expires > expectedMax {
		t.Errorf("expires should be ~1 hour from now: got %d, expected between %d and %d", expires, expectedMin, expectedMax)
	}
}
