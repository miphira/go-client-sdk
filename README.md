# Miphira Object Storage Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/miphira/go-client-sdk.svg)](https://pkg.go.dev/github.com/miphira/go-client-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/miphira/go-client-sdk)](https://goreportcard.com/report/github.com/miphira/go-client-sdk)

A Go SDK for generating presigned URLs to interact with the Miphira Object Storage API.

## Installation

```bash
go get github.com/miphira/go-client-sdk
```

## Quick Start

```go
package main

import (
    "fmt"
    "time"
    
    sdk "github.com/miphira/go-client-sdk"
)

func main() {
    // Create client with your credentials
    client := sdk.NewClient(
        "https://storage.miphira.com",      // Base URL
        "MOS_YourAccessKey12345678",        // Access Key
        "your-secret-key-here",             // Secret Key
    )

    // Generate a presigned URL for downloading a file (valid for 1 hour)
    url := client.GetObjectURL(
        "550e8400-e29b-41d4-a716-446655440000",  // Project ID
        "images",                                 // Bucket name
        "photo.jpg",                              // Filename
        time.Hour,                                // Expiry duration
    )

    fmt.Println("Download URL:", url)
}
```

## Getting Your API Credentials

### 1. Create an API Key via REST API

```bash
curl -X POST "https://storage.miphira.com/api/v1/projects/{projectId}/keys" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-app-key",
    "permissions": ["read", "write", "delete"]
  }'
```

**Response:**
```json
{
  "id": "key-uuid",
  "name": "my-app-key",
  "access_key": "MOS_<your_access_key_here>",
  "secret_key": "<your_secret_key_shown_only_once>",
  "permissions": ["read", "write", "delete"],
  "created_at": "2025-12-28T10:00:00Z",
  "warning": "Save the secret_key now. It cannot be retrieved later."
}
```

> **IMPORTANT:** Save the `secret_key` immediately! It is only shown once and cannot be retrieved later.

### 2. Store Credentials Securely

```go
// Use environment variables (recommended)
client := sdk.NewClient(
    os.Getenv("STORAGE_BASE_URL"),
    os.Getenv("STORAGE_ACCESS_KEY"),
    os.Getenv("STORAGE_SECRET_KEY"),
)
```

## API Reference

### NewClient

Creates a new Object Storage client.

```go
func NewClient(baseURL, accessKey, secretKey string) *Client
```

### GetObjectURL

Generates a presigned URL for downloading/viewing an object.

```go
func (c *Client) GetObjectURL(projectID, bucketName, filename string, expiresIn time.Duration) string
```

**Required Permission:** `read`

### UploadObjectURL

Generates a presigned URL for uploading an object.

```go
func (c *Client) UploadObjectURL(projectID, bucketName string, expiresIn time.Duration) string
```

**Required Permission:** `write`

### DeleteObjectURL

Generates a presigned URL for deleting an object.

```go
func (c *Client) DeleteObjectURL(projectID, bucketName, filename string, expiresIn time.Duration) string
```

**Required Permission:** `delete`

### GeneratePresignedURL

Low-level method to generate a presigned URL for any HTTP method and path.

```go
func (c *Client) GeneratePresignedURL(method, path string, expiresIn time.Duration) string
```

## Complete Examples

### Download a File

```go
package main

import (
    "fmt"
    "io"
    "net/http"
    "os"
    "time"

    sdk "github.com/miphira/go-client-sdk"
)

func main() {
    client := sdk.NewClient(
        os.Getenv("STORAGE_BASE_URL"),
        os.Getenv("STORAGE_ACCESS_KEY"),
        os.Getenv("STORAGE_SECRET_KEY"),
    )

    // Generate presigned URL
    url := client.GetObjectURL(
        "550e8400-e29b-41d4-a716-446655440000",
        "images",
        "photo.jpg",
        time.Hour,
    )

    // Download the file
    resp, err := http.Get(url)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    // Save to local file
    file, _ := os.Create("downloaded_photo.jpg")
    defer file.Close()
    io.Copy(file, resp.Body)

    fmt.Println("File downloaded successfully!")
}
```

### Upload a File

```go
package main

import (
    "bytes"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "os"
    "path/filepath"
    "time"

    sdk "github.com/miphira/go-client-sdk"
)

func main() {
    client := sdk.NewClient(
        os.Getenv("STORAGE_BASE_URL"),
        os.Getenv("STORAGE_ACCESS_KEY"),
        os.Getenv("STORAGE_SECRET_KEY"),
    )

    // Generate presigned upload URL
    url := client.UploadObjectURL(
        "550e8400-e29b-41d4-a716-446655440000",
        "images",
        time.Hour,
    )

    // Prepare file for upload
    filePath := "local_photo.jpg"
    file, err := os.Open(filePath)
    if err != nil {
        panic(err)
    }
    defer file.Close()

    // Create multipart form
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, _ := writer.CreateFormFile("file", filepath.Base(filePath))
    io.Copy(part, file)
    writer.Close()

    // Upload the file
    req, _ := http.NewRequest("POST", url, body)
    req.Header.Set("Content-Type", writer.FormDataContentType())

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusCreated {
        fmt.Println("File uploaded successfully!")
    }
}
```

### Delete a File

```go
package main

import (
    "fmt"
    "net/http"
    "os"
    "time"

    sdk "github.com/miphira/go-client-sdk"
)

func main() {
    client := sdk.NewClient(
        os.Getenv("STORAGE_BASE_URL"),
        os.Getenv("STORAGE_ACCESS_KEY"),
        os.Getenv("STORAGE_SECRET_KEY"),
    )

    // Generate presigned delete URL
    url := client.DeleteObjectURL(
        "550e8400-e29b-41d4-a716-446655440000",
        "images",
        "photo.jpg",
        time.Hour,
    )

    // Send DELETE request
    req, _ := http.NewRequest("DELETE", url, nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusNoContent {
        fmt.Println("File deleted successfully!")
    }
}
```

## Permissions

| Permission | Operations |
|------------|------------|
| `read` | Download/view objects (GET) |
| `write` | Upload objects (POST) |
| `delete` | Delete objects (DELETE) |

## Error Handling

| Status Code | Error | Description |
|-------------|-------|-------------|
| 401 | `invalid_access_key` | Access key not found |
| 401 | `expired_signature` | URL has expired |
| 401 | `invalid_signature` | Signature verification failed |
| 403 | `permission_denied` | Missing required permission |

## Best Practices

1. **Short expiry times** - Use the shortest practical expiry (e.g., 5-15 minutes for uploads)
2. **Environment variables** - Never hardcode credentials in source code
3. **Secure storage** - Store secret keys encrypted at rest
4. **Rotate keys** - Regularly rotate API keys for security

## License

MIT License - see [LICENSE](LICENSE) for details.
