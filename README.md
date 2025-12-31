# Miphira Object Storage Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/miphira/go-client-sdk.svg)](https://pkg.go.dev/github.com/miphira/go-client-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/miphira/go-client-sdk)](https://goreportcard.com/report/github.com/miphira/go-client-sdk)

A Go SDK for generating presigned URLs and public URLs to interact with the Miphira Object Storage API.

## Features

- ✅ **Public URLs** - Simple, permanent links for public assets (beta mode)
- ✅ **Presigned URLs** - Secure, time-limited access with HMAC-SHA256 signatures
- ✅ **Upload/Download/Delete** - Complete file lifecycle management
- ✅ **Easy Integration** - Simple API, works out of the box
- ✅ **Production Ready** - Comprehensive error handling, logging, and best practices

## Documentation

📚 **Comprehensive Guides:**

- 🚀 **[Getting Started](docs/getting-started.md)** - Quick setup and your first upload/download
- 📤 **[Upload Guide](docs/upload-guide.md)** - Detailed examples: metadata, folders, multiple files, progress tracking
- 🔐 **[Public vs Presigned URLs](docs/public-vs-presigned.md)** - When to use each approach, security considerations
- ⚡ **[Best Practices](docs/best-practices.md)** - Security, performance, error handling, production deployment

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

### Simple Upload and Get Public URL

```go
package main

import (
    "fmt"
    "log"
    "os"
    
    sdk "github.com/miphira/go-client-sdk"
)

func main() {
    // Create client
    client := sdk.NewClient(
        os.Getenv("STORAGE_BASE_URL"),    // e.g., "https://storage.miphiraapis.com"
        os.Getenv("STORAGE_PROJECT_ID"),  // e.g., "550e8400-e29b-41d4-a716-446655440000"
        os.Getenv("STORAGE_BUCKET"),      // e.g., "images"
        os.Getenv("STORAGE_ACCESS_KEY"),  // e.g., "MOS_xxxxxxxxxxxxxxxxxxxx"
        os.Getenv("STORAGE_SECRET_KEY"),  // e.g., "xxxxxxxxxxxxxxxxxxxxxxxx"
    )

    // Upload file - SDK automatically parses response with UUID filename
    resp, err := client.Upload("photo.jpg", nil)
    if err != nil {
        log.Fatal(err)
    }

    // Server response contains the actual URL with UUID filename
    fmt.Printf("Uploaded!\n")
    fmt.Printf("File ID: %s\n", resp.ID)
    fmt.Printf("Original Name: %s\n", resp.OriginalName)
    fmt.Printf("Public URL: %s\n", resp.URL)
    // Example URL: https://storage.miphiraapis.com/api/v1/public/projects/{projectId}/buckets/images/8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg
    
    // ✅ CORRECT: Use resp.URL directly (contains server-generated UUID filename)
    // ❌ WRONG: Don't use client.GetPublicObjectURL("photo.jpg") - will return 404!
}
```

**Important:** The server generates a UUID-based filename (e.g., `8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg`) for every uploaded file. Always use `resp.URL` from the upload response, not your original filename!

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

### Upload (Recommended)

Uploads a file and returns the server response with URL containing UUID filename.

```go
func (c *Client) Upload(filePath string, opts *UploadOptions) (*FileResponse, error)
```

**Example:**
```go
resp, err := client.Upload("photo.jpg", &sdk.UploadOptions{
    Metadata: map[string]interface{}{"category": "profile"},
})
if err != nil {
    log.Fatal(err)
}
fmt.Println("Public URL:", resp.URL) // Contains server-generated UUID filename
```

**Required Permission:** `write`

### UploadBytes (Recommended)

Uploads file content from memory.

```go
func (c *Client) UploadBytes(filename string, data []byte, opts *UploadOptions) (*FileResponse, error)
```

**Example:**
```go
data := []byte("Hello, World!")
resp, err := client.UploadBytes("hello.txt", data, nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Public URL:", resp.URL)
```

**Required Permission:** `write`

### Download (Recommended)

Downloads a file to local filesystem.

```go
func (c *Client) Download(filename string, localPath string, expiresIn time.Duration) error
```

**Example:**
```go
// Use UUID filename from upload response
err := client.Download("8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg", "local_photo.jpg", time.Hour)
if err != nil {
    log.Fatal(err)
}
```

**Required Permission:** `read`

### Delete (Recommended)

Deletes a file from storage.

```go
func (c *Client) Delete(filename string, expiresIn time.Duration) error
```

**Example:**
```go
err := client.Delete("8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg", time.Hour)
if err != nil {
    log.Fatal(err)
}
```

**Required Permission:** `delete`

### GetObjectURL

Generates a presigned URL for downloading/viewing an object.

```go
func (c *Client) GetObjectURL(filename string, expiresIn time.Duration) string
```

**Required Permission:** `read`

### UploadObjectURL

Generates a presigned URL for uploading an object. Use `Upload()` method instead for easier implementation.

```go
func (c *Client) UploadObjectURL(expiresIn time.Duration) string
```

**Required Permission:** `write`

### DeleteObjectURL

Generates a presigned URL for deleting an object. Use `Delete()` method instead for easier implementation.

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
    "encoding/json"
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
    uploadURL := client.UploadObjectURL(time.Hour)

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
    req, _ := http.NewRequest("POST", uploadURL, body)
    req.Header.Set("Content-Type", writer.FormDataContentType())

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusCreated {
        // Parse response to get the actual file URL with server-generated UUID
        var fileResp sdk.FileResponse
        if err := json.NewDecoder(resp.Body).Decode(&fileResp); err != nil {
            panic(err)
        }

        fmt.Println("File uploaded successfully!")
        fmt.Printf("File ID: %s\n", fileResp.ID)
        fmt.Printf("Original Name: %s\n", fileResp.OriginalName)
        fmt.Printf("Size: %s\n", fileResp.SizeFormatted)
        fmt.Printf("Public URL: %s\n", fileResp.URL)
        // The URL contains the server-generated UUID filename, not your original filename!
        // Example: https://storage.miphiraapis.com/api/v1/public/projects/{projectId}/buckets/images/8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg
    } else {
        fmt.Printf("Upload failed with status: %s\n", resp.Status)
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

## Important: Server-Generated Filenames

### Understanding File URLs

When you upload a file to Miphira Object Storage, **the server generates a new UUID-based filename** for security and uniqueness. This means:

❌ **Wrong - Don't do this:**
```go
// Upload a file named "photo.jpg"
uploadURL := client.UploadObjectURL(time.Hour)
// ... upload the file ...

// ERROR: This will NOT work! The server changed the filename!
wrongURL := client.GetPublicObjectURL("photo.jpg")
```

✅ **Correct - Do this:**
```go
// Upload a file
uploadURL := client.UploadObjectURL(time.Hour)
// ... upload the file ...

// Parse the response to get the actual filename
var fileResp sdk.FileResponse
json.NewDecoder(resp.Body).Decode(&fileResp)

// Use the URL from the response (contains server-generated UUID filename)
correctURL := fileResp.URL
// Or extract the filename from fileResp.URL and use SDK methods
```

### Example Flow

```go
// Step 1: Upload file "my-photo.jpg"
uploadURL := client.UploadObjectURL(time.Hour)
// ... multipart upload ...

// Step 2: Server responds with:
// {
//   "id": "8aabd7f7-1dbf-4ea4-8918-db66069746e7",
//   "original_name": "my-photo.jpg",
//   "url": "https://storage.miphiraapis.com/api/v1/public/projects/550e8400.../buckets/images/8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg"
// }

// Step 3: Store fileResp.URL in your database or use it directly
// The filename is now: 8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg (NOT my-photo.jpg!)

// Step 4: To access the file later, use the filename from the response
publicURL := client.GetPublicObjectURL("8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg")
```

### Why Does This Happen?

The server generates UUID filenames to:
- ✅ **Prevent naming conflicts** - Multiple users can upload files with the same name
- ✅ **Security** - Original filenames may contain sensitive information
- ✅ **Consistency** - All files follow the same naming pattern
- ✅ **Uniqueness** - Guaranteed unique identifiers across the system

### What You Should Store

After uploading, you should store in your database:
- `fileResp.ID` - File ID for API operations
- `fileResp.URL` - Complete URL (recommended - can be used directly)
- `fileResp.OriginalName` - The original filename (for display purposes)

## Best Practices

1. **Short expiry times** - Use the shortest practical expiry (e.g., 5-15 minutes for uploads)
2. **Environment variables** - Never hardcode credentials in source code
3. **Secure storage** - Store secret keys encrypted at rest
4. **Rotate keys** - Regularly rotate API keys for security
5. **Parse upload responses** - Always parse the `FileResponse` to get the actual file URL with UUID filename

## Documentation Index

### Quick Links

- 📖 **[Getting Started Guide](docs/getting-started.md)** - Set up and first examples
  - Installation and prerequisites
  - Getting credentials (project, bucket, API keys)
  - Your first upload
  - Your first download
  - Common issues and troubleshooting

- 📤 **[Upload Guide](docs/upload-guide.md)** - Complete upload examples
  - Basic upload
  - Upload with metadata
  - Upload with folder organization
  - Multiple file upload
  - Upload from memory
  - Progress tracking
  - Error handling
  - Understanding server response

- 🔐 **[Public vs Presigned URLs](docs/public-vs-presigned.md)** - Choose the right approach
  - Quick comparison table
  - Public URLs explained
  - Presigned URLs explained
  - Use case scenarios
  - Security considerations
  - Performance comparison

- ⚡ **[Best Practices](docs/best-practices.md)** - Production-ready patterns
  - Security (credentials, access control, validation)
  - Performance (client reuse, concurrency, streaming)
  - Error handling (retry logic, graceful degradation)
  - Resource management (cleanup, context usage)
  - Production deployment
  - Monitoring and logging
  - Cost optimization

## License

MIT License - see [LICENSE](LICENSE) for details.
