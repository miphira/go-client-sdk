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
curl -X POST "https://storage.miphiraapis.com/api/v1/auth/signin" \
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
curl -X POST "https://storage.miphiraapis.com/api/v1/projects" \
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
curl -X POST "https://storage.miphiraapis.com/api/v1/projects/550e8400-e29b-41d4-a716-446655440000/buckets" \
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
curl -X POST "https://storage.miphiraapis.com/api/v1/projects/550e8400-e29b-41d4-a716-446655440000/keys" \
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
| `baseURL` | Fixed API endpoint | `https://storage.miphiraapis.com` |
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
    // Create client with all configuration
    client := sdk.NewClient(
        os.Getenv("STORAGE_BASE_URL"),    // e.g., "https://storage.miphiraapis.com"
        os.Getenv("STORAGE_PROJECT_ID"),  // e.g., "550e8400-e29b-41d4-a716-446655440000"
        os.Getenv("STORAGE_BUCKET"),      // e.g., "images"
        os.Getenv("STORAGE_ACCESS_KEY"),  // e.g., "MOS_xxxxxxxxxxxxxxxxxxxx"
        os.Getenv("STORAGE_SECRET_KEY"),  // e.g., "xxxxxxxxxxxxxxxxxxxxxxxx"
    )

    // Option 1: Public URLs (Beta - no auth required)
    publicURL := client.GetPublicObjectURL("photo.jpg")
    fmt.Println("Public URL:", publicURL)

    // Option 2: Presigned URLs (Secure - with expiration)
    downloadURL := client.GetObjectURL("photo.jpg", time.Hour)
    uploadURL := client.UploadObjectURL(time.Hour)
    deleteURL := client.DeleteObjectURL("photo.jpg", time.Hour)

    fmt.Println("Download:", downloadURL)
    fmt.Println("Upload:", uploadURL)
    fmt.Println("Delete:", deleteURL)
}
```

**Environment variables (.env):**
```bash
STORAGE_BASE_URL=https://storage.miphiraapis.com
STORAGE_PROJECT_ID=550e8400-e29b-41d4-a716-446655440000
STORAGE_BUCKET=images
STORAGE_ACCESS_KEY=MOS_xxxxxxxxxxxxxxxxxxxx
STORAGE_SECRET_KEY=xxxxxxxxxxxxxxxxxxxxxxxx
```

## Public URLs vs Presigned URLs

### When to Use Public URLs

**Public URLs** (`GetPublicObjectURL`) are best for:
- ✅ **Beta/Development** - All buckets are public in beta mode
- ✅ **CDN Integration** - Public assets served through CDN
- ✅ **Static Assets** - Images, logos, public documents
- ✅ **Simple Access** - No authentication required
- ✅ **Permanent Links** - URLs never expire

**Example:**
```go
// Generate public URL (works immediately, no expiration)
url := client.GetPublicObjectURL("logo.png")
// https://storage.miphiraapis.com/api/v1/public/projects/{id}/buckets/images/logo.png

// Use in HTML, share publicly, embed in emails
<img src="{url}" />
```

### When to Use Presigned URLs

**Presigned URLs** (`GetObjectURL`, `UploadObjectURL`, `DeleteObjectURL`) are best for:
- ✅ **Production** - Secure access control
- ✅ **Private Files** - User data, documents, private media
- ✅ **Time-Limited Access** - Temporary download/upload links
- ✅ **Permission Control** - read/write/delete permissions
- ✅ **Secure Operations** - Upload, delete operations

**Example:**
```go
// Generate presigned URL (secure, expires after 1 hour)
url := client.GetObjectURL("private-document.pdf", time.Hour)
// https://storage.miphiraapis.com/api/v1/projects/{id}/buckets/docs/objects/private-document.pdf?X-Mos-AccessKey=...&X-Mos-Expires=...&X-Mos-Signature=...

// Share with specific users, expires automatically
```

### Comparison Table

| Feature | Public URL | Presigned URL |
|---------|------------|---------------|
| **Authentication** | None | HMAC-SHA256 signature |
| **Expiration** | Never | Configurable (e.g., 1 hour) |
| **Security** | Public access | Access-controlled |
| **Use Case** | Static assets | Private files |
| **Beta Mode** | ✅ Available | ✅ Available |
| **Production** | Public buckets only | All buckets |
| **URL Length** | Short | Long (with signature) |

## API Reference

### NewClient

Creates a new Object Storage client with all required configuration.

```go
func NewClient(baseURL, projectID, bucketName, accessKey, secretKey string) *Client
```

| Parameter | Description |
|-----------|-------------|
| `baseURL` | Storage server URL |
| `projectID` | Project UUID (from Step 2) |
| `bucketName` | Bucket name (from Step 3) |
| `accessKey` | API access key (from Step 4) |
| `secretKey` | API secret key (from Step 4) |

### GetObjectURL

Generates a presigned URL for downloading/viewing an object.

```go
func (c *Client) GetObjectURL(filename string, expiresIn time.Duration) string
```

**Required Permission:** `read`

### UploadObjectURL

Generates a presigned URL for uploading an object.

```go
func (c *Client) UploadObjectURL(expiresIn time.Duration) string
```

**Required Permission:** `write`

### DeleteObjectURL

Generates a presigned URL for deleting an object.

```go
func (c *Client) DeleteObjectURL(filename string, expiresIn time.Duration) string
```

**Required Permission:** `delete`

### GetPublicObjectURL

Generates a public URL for accessing an object without authentication. No signature or expiration required.

```go
func (c *Client) GetPublicObjectURL(filename string) string
```

**Note:** This only works in beta mode where all buckets are public. In production, use presigned URLs with `GetObjectURL()` for secure access.

**Example:**
```go
// Generate public URL (no authentication required)
publicURL := client.GetPublicObjectURL("photo.jpg")
// Returns: https://storage.miphiraapis.com/api/v1/public/projects/550e8400-e29b-41d4-a716-446655440000/buckets/images/photo.jpg

// Anyone can access this URL directly in a browser or with curl
// curl https://storage.miphiraapis.com/api/v1/public/projects/{projectId}/buckets/{bucket}/photo.jpg
```

### GeneratePresignedURL

Low-level method to generate a presigned URL for any HTTP method and path.

```go
func (c *Client) GeneratePresignedURL(method, path string, expiresIn time.Duration) string
```

## Complete Examples

### Access a File via Public URL

```go
package main

import (
    "fmt"
    "io"
    "net/http"
    "os"

    sdk "github.com/miphira/go-client-sdk"
)

func main() {
    client := sdk.NewClient(
        os.Getenv("STORAGE_BASE_URL"),
        os.Getenv("STORAGE_PROJECT_ID"),
        os.Getenv("STORAGE_BUCKET"),
        os.Getenv("STORAGE_ACCESS_KEY"),
        os.Getenv("STORAGE_SECRET_KEY"),
    )

    // Generate public URL (no authentication required)
    publicURL := client.GetPublicObjectURL("photo.jpg")
    fmt.Println("Public URL:", publicURL)

    // Download the file (no authentication needed!)
    resp, err := http.Get(publicURL)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        fmt.Printf("Failed to download: %s\n", resp.Status)
        return
    }

    // Save to local file
    file, _ := os.Create("downloaded_photo.jpg")
    defer file.Close()
    io.Copy(file, resp.Body)

    fmt.Println("File downloaded successfully!")
}
```

### Download a File (with Presigned URL)

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
        os.Getenv("STORAGE_PROJECT_ID"),
        os.Getenv("STORAGE_BUCKET"),
        os.Getenv("STORAGE_ACCESS_KEY"),
        os.Getenv("STORAGE_SECRET_KEY"),
    )

    // Generate presigned URL
    url := client.GetObjectURL("photo.jpg", time.Hour)

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
        os.Getenv("STORAGE_PROJECT_ID"),
        os.Getenv("STORAGE_BUCKET"),
        os.Getenv("STORAGE_ACCESS_KEY"),
        os.Getenv("STORAGE_SECRET_KEY"),
    )

    // Generate presigned upload URL
    url := client.UploadObjectURL(time.Hour)

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
        os.Getenv("STORAGE_PROJECT_ID"),
        os.Getenv("STORAGE_BUCKET"),
        os.Getenv("STORAGE_ACCESS_KEY"),
        os.Getenv("STORAGE_SECRET_KEY"),
    )

    // Generate presigned delete URL
    url := client.DeleteObjectURL("photo.jpg", time.Hour)

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
