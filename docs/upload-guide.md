# Upload Guide

This guide covers everything you need to know about uploading files with the Miphira Object Storage SDK.

## Table of Contents

- [Basic Upload](#basic-upload)
- [Upload with Metadata](#upload-with-metadata)
- [Upload with Folder Organization](#upload-with-folder-organization)
- [Multiple File Upload](#multiple-file-upload)
- [Upload from Memory](#upload-from-memory)
- [Progress Tracking](#progress-tracking)
- [Error Handling](#error-handling)
- [Understanding Server Response](#understanding-server-response)

## Basic Upload

The simplest way to upload a file:

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

func uploadFile(client *sdk.Client, filePath string) (*sdk.FileResponse, error) {
    // Generate presigned upload URL
    uploadURL := client.UploadObjectURL(time.Hour)

    // Open file
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    // Create multipart form
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("file", filepath.Base(filePath))
    if err != nil {
        return nil, err
    }
    io.Copy(part, file)
    writer.Close()

    // Upload
    req, err := http.NewRequest("POST", uploadURL, body)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        return nil, fmt.Errorf("upload failed with status: %s", resp.Status)
    }

    // Parse response
    var fileResp sdk.FileResponse
    if err := json.NewDecoder(resp.Body).Decode(&fileResp); err != nil {
        return nil, err
    }

    return &fileResp, nil
}

func main() {
    client := sdk.NewClient(
        os.Getenv("STORAGE_BASE_URL"),
        os.Getenv("STORAGE_PROJECT_ID"),
        os.Getenv("STORAGE_BUCKET"),
        os.Getenv("STORAGE_ACCESS_KEY"),
        os.Getenv("STORAGE_SECRET_KEY"),
    )

    fileResp, err := uploadFile(client, "photo.jpg")
    if err != nil {
        panic(err)
    }

    fmt.Printf("✅ Uploaded: %s\n", fileResp.URL)
}
```

## Upload with Metadata

Add custom metadata to your files:

```go
func uploadWithMetadata(client *sdk.Client, filePath string, metadata map[string]interface{}) (*sdk.FileResponse, error) {
    uploadURL := client.UploadObjectURL(time.Hour)

    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    // Create multipart form
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    // Add file
    part, err := writer.CreateFormFile("file", filepath.Base(filePath))
    if err != nil {
        return nil, err
    }
    io.Copy(part, file)

    // Add metadata as JSON
    if metadata != nil {
        metadataJSON, err := json.Marshal(metadata)
        if err != nil {
            return nil, err
        }
        writer.WriteField("metadata", string(metadataJSON))
    }

    writer.Close()

    // Upload
    req, err := http.NewRequest("POST", uploadURL, body)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        return nil, fmt.Errorf("upload failed: %s", resp.Status)
    }

    var fileResp sdk.FileResponse
    json.NewDecoder(resp.Body).Decode(&fileResp)
    return &fileResp, nil
}

// Usage
metadata := map[string]interface{}{
    "userId":      "usr_123",
    "uploadedBy":  "mobile-app",
    "category":    "profile-picture",
    "public":      true,
    "tags":        []string{"avatar", "profile"},
}

fileResp, err := uploadWithMetadata(client, "avatar.jpg", metadata)
```

## Upload with Folder Organization

Organize files in folders within your bucket:

```go
func uploadToFolder(client *sdk.Client, filePath string, folderPath string) (*sdk.FileResponse, error) {
    uploadURL := client.UploadObjectURL(time.Hour)

    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    // Create multipart form
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    // Add file
    part, err := writer.CreateFormFile("file", filepath.Base(filePath))
    if err != nil {
        return nil, err
    }
    io.Copy(part, file)

    // Add folder path
    if folderPath != "" {
        writer.WriteField("folder_path", folderPath)
    }

    writer.Close()

    // Upload
    req, err := http.NewRequest("POST", uploadURL, body)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        return nil, fmt.Errorf("upload failed: %s", resp.Status)
    }

    var fileResp sdk.FileResponse
    json.NewDecoder(resp.Body).Decode(&fileResp)
    return &fileResp, nil
}

// Usage examples
uploadToFolder(client, "avatar.jpg", "avatars")
// URL: .../buckets/images/avatars/8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg

uploadToFolder(client, "contract.pdf", "documents/contracts")
// URL: .../buckets/docs/documents/contracts/8aabd7f7-1dbf-4ea4-8918-db66069746e7.pdf

uploadToFolder(client, "report.pdf", "reports/2024/Q4")
// URL: .../buckets/docs/reports/2024/Q4/8aabd7f7-1dbf-4ea4-8918-db66069746e7.pdf
```

## Multiple File Upload

Upload multiple files efficiently:

```go
func uploadMultipleFiles(client *sdk.Client, filePaths []string) ([]*sdk.FileResponse, error) {
    results := make([]*sdk.FileResponse, 0, len(filePaths))
    errors := make([]error, 0)

    // Upload files concurrently
    type result struct {
        resp *sdk.FileResponse
        err  error
    }
    
    resultChan := make(chan result, len(filePaths))

    for _, filePath := range filePaths {
        go func(path string) {
            resp, err := uploadFile(client, path)
            resultChan <- result{resp: resp, err: err}
        }(filePath)
    }

    // Collect results
    for i := 0; i < len(filePaths); i++ {
        r := <-resultChan
        if r.err != nil {
            errors = append(errors, r.err)
        } else {
            results = append(results, r.resp)
        }
    }

    if len(errors) > 0 {
        return results, fmt.Errorf("some uploads failed: %v", errors)
    }

    return results, nil
}

// Usage
files := []string{
    "photo1.jpg",
    "photo2.jpg",
    "photo3.jpg",
    "document.pdf",
}

results, err := uploadMultipleFiles(client, files)
if err != nil {
    fmt.Printf("Some uploads failed: %v\n", err)
}

for _, fileResp := range results {
    fmt.Printf("✅ Uploaded: %s (%s)\n", fileResp.OriginalName, fileResp.URL)
}
```

## Upload from Memory

Upload data directly from memory without saving to disk:

```go
func uploadFromBytes(client *sdk.Client, data []byte, filename string) (*sdk.FileResponse, error) {
    uploadURL := client.UploadObjectURL(time.Hour)

    // Create multipart form
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    // Add file from bytes
    part, err := writer.CreateFormFile("file", filename)
    if err != nil {
        return nil, err
    }
    part.Write(data)
    writer.Close()

    // Upload
    req, err := http.NewRequest("POST", uploadURL, body)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        return nil, fmt.Errorf("upload failed: %s", resp.Status)
    }

    var fileResp sdk.FileResponse
    json.NewDecoder(resp.Body).Decode(&fileResp)
    return &fileResp, nil
}

// Usage examples

// Example 1: Upload generated image
imageData := generateImage() // returns []byte
fileResp, err := uploadFromBytes(client, imageData, "generated-image.png")

// Example 2: Upload JSON data
jsonData, _ := json.Marshal(map[string]interface{}{
    "name": "John Doe",
    "age":  30,
})
fileResp, err := uploadFromBytes(client, jsonData, "data.json")

// Example 3: Upload CSV data
csvData := []byte("name,age,email\nJohn,30,john@example.com\n")
fileResp, err := uploadFromBytes(client, csvData, "users.csv")
```

## Progress Tracking

Track upload progress for large files:

```go
type ProgressReader struct {
    reader   io.Reader
    total    int64
    read     int64
    callback func(read, total int64, percent float64)
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
    n, err := pr.reader.Read(p)
    pr.read += int64(n)
    if pr.callback != nil {
        percent := float64(pr.read) / float64(pr.total) * 100
        pr.callback(pr.read, pr.total, percent)
    }
    return n, err
}

func uploadWithProgress(client *sdk.Client, filePath string) (*sdk.FileResponse, error) {
    uploadURL := client.UploadObjectURL(time.Hour)

    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    // Get file size
    fileInfo, err := file.Stat()
    if err != nil {
        return nil, err
    }
    fileSize := fileInfo.Size()

    // Create multipart form
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("file", filepath.Base(filePath))
    if err != nil {
        return nil, err
    }

    // Wrap with progress reader
    progressReader := &ProgressReader{
        reader: file,
        total:  fileSize,
        callback: func(read, total int64, percent float64) {
            fmt.Printf("\rUploading: %.2f%% (%d/%d bytes)", percent, read, total)
        },
    }

    io.Copy(part, progressReader)
    writer.Close()
    fmt.Println() // New line after progress

    // Upload
    req, err := http.NewRequest("POST", uploadURL, body)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        return nil, fmt.Errorf("upload failed: %s", resp.Status)
    }

    var fileResp sdk.FileResponse
    json.NewDecoder(resp.Body).Decode(&fileResp)
    return &fileResp, nil
}

// Usage
fileResp, err := uploadWithProgress(client, "large-video.mp4")
// Output: Uploading: 100.00% (104857600/104857600 bytes)
```

## Error Handling

Robust error handling for uploads:

```go
type UploadError struct {
    StatusCode int
    Message    string
    FilePath   string
}

func (e *UploadError) Error() string {
    return fmt.Sprintf("upload failed for %s: %s (status: %d)", e.FilePath, e.Message, e.StatusCode)
}

func uploadWithRetry(client *sdk.Client, filePath string, maxRetries int) (*sdk.FileResponse, error) {
    var lastErr error

    for attempt := 0; attempt <= maxRetries; attempt++ {
        if attempt > 0 {
            fmt.Printf("Retry attempt %d/%d...\n", attempt, maxRetries)
            time.Sleep(time.Second * time.Duration(attempt)) // Exponential backoff
        }

        uploadURL := client.UploadObjectURL(time.Hour)

        file, err := os.Open(filePath)
        if err != nil {
            return nil, err // File not found, no retry
        }
        defer file.Close()

        body := &bytes.Buffer{}
        writer := multipart.NewWriter(body)
        part, err := writer.CreateFormFile("file", filepath.Base(filePath))
        if err != nil {
            return nil, err
        }
        io.Copy(part, file)
        writer.Close()

        req, err := http.NewRequest("POST", uploadURL, body)
        if err != nil {
            return nil, err
        }
        req.Header.Set("Content-Type", writer.FormDataContentType())

        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            lastErr = err
            continue // Retry on network error
        }
        defer resp.Body.Close()

        if resp.StatusCode == http.StatusCreated {
            var fileResp sdk.FileResponse
            if err := json.NewDecoder(resp.Body).Decode(&fileResp); err != nil {
                return nil, err
            }
            return &fileResp, nil
        }

        // Handle specific error codes
        switch resp.StatusCode {
        case http.StatusBadRequest:
            // Client error, don't retry
            return nil, &UploadError{
                StatusCode: resp.StatusCode,
                Message:    "invalid request",
                FilePath:   filePath,
            }
        case http.StatusUnauthorized:
            // Auth error, don't retry
            return nil, &UploadError{
                StatusCode: resp.StatusCode,
                Message:    "unauthorized - check your API keys",
                FilePath:   filePath,
            }
        case http.StatusTooManyRequests:
            // Rate limited, retry with longer delay
            time.Sleep(time.Second * 5)
            lastErr = &UploadError{
                StatusCode: resp.StatusCode,
                Message:    "rate limited",
                FilePath:   filePath,
            }
            continue
        default:
            // Server error, retry
            lastErr = &UploadError{
                StatusCode: resp.StatusCode,
                Message:    resp.Status,
                FilePath:   filePath,
            }
            continue
        }
    }

    return nil, fmt.Errorf("upload failed after %d retries: %v", maxRetries, lastErr)
}

// Usage
fileResp, err := uploadWithRetry(client, "important-file.pdf", 3)
if err != nil {
    if uploadErr, ok := err.(*UploadError); ok {
        fmt.Printf("Upload error: %s (status: %d)\n", uploadErr.Message, uploadErr.StatusCode)
    } else {
        fmt.Printf("Upload error: %v\n", err)
    }
    return
}
```

## Understanding Server Response

The server always returns a `FileResponse` with important information:

```go
type FileResponse struct {
    ID            string                 // Server-generated file UUID
    Name          string                 // Display name (same as OriginalName)
    OriginalName  string                 // Your original filename
    Size          int64                  // File size in bytes
    SizeFormatted string                 // Human-readable size (e.g., "2.5 MB")
    MimeType      string                 // Content type (e.g., "image/jpeg")
    BucketID      string                 // Bucket UUID
    URL           string                 // Complete public URL (IMPORTANT!)
    Metadata      map[string]interface{} // Your custom metadata
    CreatedAt     string                 // ISO 8601 timestamp
    UpdatedAt     string                 // ISO 8601 timestamp
}
```

### Key Points

1. **Server-Generated UUID**: The filename in `URL` is a UUID, NOT your original filename
2. **Store the URL**: Save `fileResp.URL` to your database for direct access
3. **Original Filename**: Use `fileResp.OriginalName` for display purposes
4. **File ID**: Use `fileResp.ID` for API operations (delete, update)

### Example Response

```json
{
  "id": "8aabd7f7-1dbf-4ea4-8918-db66069746e7",
  "name": "photo.jpg",
  "original_name": "photo.jpg",
  "size": 2621440,
  "size_formatted": "2.5 MB",
  "mime_type": "image/jpeg",
  "bucket_id": "bkt_3m4n5o6p",
  "url": "https://storage.miphiraapis.com/api/v1/public/projects/550e8400.../buckets/images/8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg",
  "metadata": {
    "userId": "usr_123",
    "category": "profile-picture"
  },
  "created_at": "2025-12-31T10:30:00Z",
  "updated_at": "2025-12-31T10:30:00Z"
}
```

## Complete Example: Full-Featured Upload

Putting it all together:

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

type UploadOptions struct {
    FolderPath string
    Metadata   map[string]interface{}
    MaxRetries int
}

func fullFeaturedUpload(client *sdk.Client, filePath string, opts *UploadOptions) (*sdk.FileResponse, error) {
    if opts == nil {
        opts = &UploadOptions{MaxRetries: 3}
    }

    var lastErr error

    for attempt := 0; attempt <= opts.MaxRetries; attempt++ {
        if attempt > 0 {
            fmt.Printf("Retry %d/%d...\n", attempt, opts.MaxRetries)
            time.Sleep(time.Second * time.Duration(attempt))
        }

        uploadURL := client.UploadObjectURL(time.Hour)

        file, err := os.Open(filePath)
        if err != nil {
            return nil, err
        }
        defer file.Close()

        // Get file size for progress
        fileInfo, _ := file.Stat()
        fileSize := fileInfo.Size()

        // Create multipart form
        body := &bytes.Buffer{}
        writer := multipart.NewWriter(body)

        // Add file with progress
        part, _ := writer.CreateFormFile("file", filepath.Base(filePath))
        
        progressReader := &ProgressReader{
            reader: file,
            total:  fileSize,
            callback: func(read, total int64, percent float64) {
                fmt.Printf("\r📤 Uploading: %.1f%%", percent)
            },
        }
        io.Copy(part, progressReader)
        fmt.Println() // New line

        // Add optional folder path
        if opts.FolderPath != "" {
            writer.WriteField("folder_path", opts.FolderPath)
        }

        // Add optional metadata
        if opts.Metadata != nil {
            metadataJSON, _ := json.Marshal(opts.Metadata)
            writer.WriteField("metadata", string(metadataJSON))
        }

        writer.Close()

        // Upload
        req, _ := http.NewRequest("POST", uploadURL, body)
        req.Header.Set("Content-Type", writer.FormDataContentType())

        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            lastErr = err
            continue
        }
        defer resp.Body.Close()

        if resp.StatusCode == http.StatusCreated {
            var fileResp sdk.FileResponse
            json.NewDecoder(resp.Body).Decode(&fileResp)
            return &fileResp, nil
        }

        lastErr = fmt.Errorf("upload failed: %s", resp.Status)
    }

    return nil, lastErr
}

func main() {
    client := sdk.NewClient(
        os.Getenv("STORAGE_BASE_URL"),
        os.Getenv("STORAGE_PROJECT_ID"),
        os.Getenv("STORAGE_BUCKET"),
        os.Getenv("STORAGE_ACCESS_KEY"),
        os.Getenv("STORAGE_SECRET_KEY"),
    )

    fileResp, err := fullFeaturedUpload(client, "photo.jpg", &UploadOptions{
        FolderPath: "avatars",
        Metadata: map[string]interface{}{
            "userId":   "usr_123",
            "category": "profile",
        },
        MaxRetries: 3,
    })

    if err != nil {
        panic(err)
    }

    fmt.Printf("\n✅ Upload successful!\n")
    fmt.Printf("🔗 URL: %s\n", fileResp.URL)
    fmt.Printf("📏 Size: %s\n", fileResp.SizeFormatted)
}
```

## Next Steps

- 📖 [Public vs Presigned URLs](public-vs-presigned.md) - Learn when to use each
- ⚡ [Best Practices](best-practices.md) - Optimize your uploads
- 🏠 [Getting Started](getting-started.md) - Return to basics
