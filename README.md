# Miphira Object Storage Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/miphira/go-client-sdk.svg)](https://pkg.go.dev/github.com/miphira/go-client-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/miphira/go-client-sdk)](https://goreportcard.com/report/github.com/miphira/go-client-sdk)

A Go SDK for generating presigned URLs to interact with the Miphira Object Storage API.

## Installation

```bash
go get github.com/miphira/go-client-sdk
```

## Prerequisites - Get Your Credentials

Before using the SDK, you need to create a project, bucket, and API key via the REST API.

### Step 1: Sign In (Get JWT Token)

```bash
curl -X POST "https://storage.miphira.com/api/v1/auth/signin" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "your-email@example.com",
    "password": "your-password"
  }'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": { "id": "user-uuid", "email": "your-email@example.com" }
}
```

### Step 2: Create a Project (Get Project ID)

```bash
curl -X POST "https://storage.miphira.com/api/v1/projects" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-app",
    "description": "My application storage"
  }'
```

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",  <-- THIS IS YOUR PROJECT ID
  "name": "my-app",
  "description": "My application storage"
}
```

### Step 3: Create a Bucket (Choose Bucket Name)

```bash
curl -X POST "https://storage.miphira.com/api/v1/projects/550e8400-e29b-41d4-a716-446655440000/buckets" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "images"
  }'
```

Response:
```json
{
  "id": "bucket-uuid",
  "name": "images",  <-- THIS IS YOUR BUCKET NAME
  "project_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Step 4: Create an API Key (Get Access Key & Secret Key)

```bash
curl -X POST "https://storage.miphira.com/api/v1/projects/550e8400-e29b-41d4-a716-446655440000/keys" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-app-key",
    "permissions": ["read", "write", "delete"]
  }'
```

Response:
```json
{
  "id": "key-uuid",
  "name": "my-app-key",
  "access_key": "MOS_xxxxxxxxxxxxxxxxxxxx",  <-- YOUR ACCESS KEY
  "secret_key": "xxxxxxxxxxxxxxxxxxxxxxxx",  <-- YOUR SECRET KEY (SAVE IT!)
  "permissions": ["read", "write", "delete"],
  "warning": "Save the secret_key now. It cannot be retrieved later."
}
```

> **IMPORTANT:** Save the `secret_key` immediately! It is only shown once.

### Summary - What You Need

| Parameter | Where to Get | Example |
|-----------|--------------|---------|
| `baseURL` | Your server URL | `https://storage.miphira.com` |
| `projectID` | Step 2 response `id` | `550e8400-e29b-41d4-a716-446655440000` |
| `bucketName` | Step 3 request `name` | `images` |
| `accessKey` | Step 4 response `access_key` | `MOS_xxxxxxxxxxxxxxxxxxxx` |
| `secretKey` | Step 4 response `secret_key` | `xxxxxxxxxxxxxxxxxxxxxxxx` |

## Quick Start

```go
package main

import (
    "fmt"
    "os"
    "time"
    
    sdk "github.com/miphira/go-client-sdk"
)

func main() {
    // Load from environment variables (recommended)
    client := sdk.NewClient(
        os.Getenv("STORAGE_BASE_URL"),    // e.g., "https://storage.miphira.com"
        os.Getenv("STORAGE_ACCESS_KEY"),  // e.g., "MOS_xxxxxxxxxxxxxxxxxxxx"
        os.Getenv("STORAGE_SECRET_KEY"),  // e.g., "xxxxxxxxxxxxxxxxxxxxxxxx"
    )

    // These values come from Steps 2 & 3 above
    projectID := os.Getenv("STORAGE_PROJECT_ID")  // e.g., "550e8400-e29b-41d4-a716-446655440000"
    bucketName := os.Getenv("STORAGE_BUCKET")     // e.g., "images"

    // Generate a presigned URL for downloading a file (valid for 1 hour)
    url := client.GetObjectURL(projectID, bucketName, "photo.jpg", time.Hour)

    fmt.Println("Download URL:", url)
}
```

**Environment variables (.env):**
```bash
STORAGE_BASE_URL=https://storage.miphira.com
STORAGE_PROJECT_ID=550e8400-e29b-41d4-a716-446655440000
STORAGE_BUCKET=images
STORAGE_ACCESS_KEY=MOS_xxxxxxxxxxxxxxxxxxxx
STORAGE_SECRET_KEY=xxxxxxxxxxxxxxxxxxxxxxxx
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
