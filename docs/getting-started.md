# Getting Started with Miphira Object Storage SDK

This guide will help you get started with the Miphira Object Storage Go SDK in just a few minutes.

## Table of Contents

- [Installation](#installation)
- [Prerequisites](#prerequisites)
- [Quick Setup](#quick-setup)
- [Your First Upload](#your-first-upload)
- [Your First Download](#your-first-download)
- [Next Steps](#next-steps)

## Installation

Install the SDK using Go modules:

```bash
go get github.com/miphira/go-client-sdk
```

## Prerequisites

Before you can use the SDK, you need to obtain credentials from the Miphira Object Storage API. You'll need:

1. **Project ID** - Created via `/api/v1/projects`
2. **Bucket Name** - Created via `/api/v1/projects/{projectId}/buckets`
3. **Access Key & Secret Key** - Created via `/api/v1/projects/{projectId}/keys`

### Step-by-Step: Getting Your Credentials

#### 1. Sign In

First, authenticate to get a JWT token:

```bash
curl -X POST "https://storage.miphiraapis.com/api/v1/auth/signin" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "your-email@example.com",
    "password": "your-password"
  }'
```

Save the `token` from the response.

#### 2. Create a Project

```bash
curl -X POST "https://storage.miphiraapis.com/api/v1/projects" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-app",
    "description": "My application storage"
  }'
```

Save the `id` field - this is your **Project ID**.

#### 3. Create a Bucket

```bash
curl -X POST "https://storage.miphiraapis.com/api/v1/projects/PROJECT_ID/buckets" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "images"
  }'
```

The `name` field is your **Bucket Name**.

#### 4. Create API Keys

```bash
curl -X POST "https://storage.miphiraapis.com/api/v1/projects/PROJECT_ID/keys" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-app-key",
    "permissions": ["read", "write", "delete"]
  }'
```

Save both `access_key` and `secret_key`. **Important:** The secret key is only shown once!

## Quick Setup

### Environment Variables

Create a `.env` file in your project:

```bash
STORAGE_BASE_URL=https://storage.miphiraapis.com
STORAGE_PROJECT_ID=550e8400-e29b-41d4-a716-446655440000
STORAGE_BUCKET=images
STORAGE_ACCESS_KEY=MOS_xxxxxxxxxxxxxxxxxxxx
STORAGE_SECRET_KEY=xxxxxxxxxxxxxxxxxxxxxxxx
```

### Initialize the Client

```go
package main

import (
    "log"
    "os"

    sdk "github.com/miphira/go-client-sdk"
)

func main() {
    // Load environment variables (use godotenv or your preferred method)
    
    // Initialize client
    client := sdk.NewClient(
        os.Getenv("STORAGE_BASE_URL"),
        os.Getenv("STORAGE_PROJECT_ID"),
        os.Getenv("STORAGE_BUCKET"),
        os.Getenv("STORAGE_ACCESS_KEY"),
        os.Getenv("STORAGE_SECRET_KEY"),
    )

    log.Println("Client initialized successfully!")
}
```

## Your First Upload

Here's a complete example to upload a file and get its public URL:

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "mime/multipart"
    "net/http"
    "os"
    "path/filepath"
    "time"

    sdk "github.com/miphira/go-client-sdk"
)

func main() {
    // Initialize client
    client := sdk.NewClient(
        os.Getenv("STORAGE_BASE_URL"),
        os.Getenv("STORAGE_PROJECT_ID"),
        os.Getenv("STORAGE_BUCKET"),
        os.Getenv("STORAGE_ACCESS_KEY"),
        os.Getenv("STORAGE_SECRET_KEY"),
    )

    // Generate presigned upload URL (valid for 1 hour)
    uploadURL := client.UploadObjectURL(time.Hour)
    log.Printf("Upload URL: %s", uploadURL)

    // Prepare the file
    filePath := "example.jpg"
    file, err := os.Open(filePath)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    // Create multipart form data
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("file", filepath.Base(filePath))
    if err != nil {
        log.Fatal(err)
    }
    io.Copy(part, file)
    writer.Close()

    // Upload the file
    req, err := http.NewRequest("POST", uploadURL, body)
    if err != nil {
        log.Fatal(err)
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    // Check response
    if resp.StatusCode != http.StatusCreated {
        log.Fatalf("Upload failed with status: %s", resp.Status)
    }

    // Parse response to get file information
    var fileResp sdk.FileResponse
    if err := json.NewDecoder(resp.Body).Decode(&fileResp); err != nil {
        log.Fatal(err)
    }

    // Success! Print file information
    fmt.Println("\n✅ File uploaded successfully!")
    fmt.Printf("📁 File ID: %s\n", fileResp.ID)
    fmt.Printf("📝 Original Name: %s\n", fileResp.OriginalName)
    fmt.Printf("📏 Size: %s\n", fileResp.SizeFormatted)
    fmt.Printf("🔗 Public URL: %s\n", fileResp.URL)
    fmt.Println("\nYou can now access your file at the URL above!")
}
```

**Output:**
```
✅ File uploaded successfully!
📁 File ID: 8aabd7f7-1dbf-4ea4-8918-db66069746e7
📝 Original Name: example.jpg
📏 Size: 2.5 MB
🔗 Public URL: https://storage.miphiraapis.com/api/v1/public/projects/550e8400.../buckets/images/8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg

You can now access your file at the URL above!
```

## Your First Download

Download a file using either a public URL (beta mode) or a presigned URL:

### Option 1: Public URL (Recommended for Beta)

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

    // Use the filename from the upload response (server-generated UUID)
    filename := "8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg"
    
    // Generate public URL
    publicURL := client.GetPublicObjectURL(filename)
    fmt.Printf("Downloading from: %s\n", publicURL)

    // Download the file
    resp, err := http.Get(publicURL)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        fmt.Printf("Download failed: %s\n", resp.Status)
        return
    }

    // Save to local file
    outFile, err := os.Create("downloaded_file.jpg")
    if err != nil {
        panic(err)
    }
    defer outFile.Close()

    _, err = io.Copy(outFile, resp.Body)
    if err != nil {
        panic(err)
    }

    fmt.Println("✅ File downloaded successfully!")
}
```

### Option 2: Presigned URL (Secure)

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

    // Generate presigned URL (valid for 15 minutes)
    filename := "8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg"
    downloadURL := client.GetObjectURL(filename, 15*time.Minute)

    // Download using presigned URL
    resp, err := http.Get(downloadURL)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    // ... save file as above ...

    fmt.Println("✅ File downloaded successfully!")
}
```

## Next Steps

Now that you have the basics working, explore more advanced features:

- 📖 [Upload Guide](upload-guide.md) - Detailed upload examples with metadata, folder organization
- 🔐 [Public vs Presigned URLs](public-vs-presigned.md) - When to use each approach
- ⚡ [Best Practices](best-practices.md) - Security, performance, and error handling tips
- 📚 [Main README](../README.md) - Complete API reference

## Common Issues

### Issue: "File not found" after upload

**Problem:** You're using the original filename instead of the server-generated UUID filename.

**Solution:** Always parse the `FileResponse` after upload and use `fileResp.URL` or extract the filename from the response.

```go
// ❌ Wrong
publicURL := client.GetPublicObjectURL("myfile.jpg")

// ✅ Correct
var fileResp sdk.FileResponse
json.NewDecoder(resp.Body).Decode(&fileResp)
publicURL := fileResp.URL
```

### Issue: "Invalid signature" error

**Problem:** Your secret key might be incorrect or the URL has expired.

**Solution:**
1. Verify your `STORAGE_SECRET_KEY` is correct
2. For presigned URLs, check the expiration time
3. Regenerate the URL if it has expired

### Issue: "Unauthorized" error

**Problem:** Missing or incorrect access key.

**Solution:** Verify your `STORAGE_ACCESS_KEY` is correct and the API key has the required permissions.

## Need Help?

- 📧 Email: support@miphira.com
- 📖 Documentation: https://docs.miphira.com
- 🐛 Issues: https://github.com/miphira/go-client-sdk/issues
