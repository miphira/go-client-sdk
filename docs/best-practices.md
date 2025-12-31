# Best Practices for Miphira Object Storage SDK

A comprehensive guide to using the SDK securely, efficiently, and reliably in production.

## Table of Contents

- [Security](#security)
- [Performance](#performance)
- [Error Handling](#error-handling)
- [Resource Management](#resource-management)
- [Production Deployment](#production-deployment)
- [Monitoring & Logging](#monitoring--logging)
- [Cost Optimization](#cost-optimization)

## Security

### 1. Credential Management

**❌ Never hardcode credentials:**
```go
// BAD - Credentials exposed in code
client := sdk.NewClient(
    "https://storage.miphiraapis.com",
    "550e8400-e29b-41d4-a716-446655440000",
    "images",
    "MOS_xxxxxxxxxxxxxxxxxxxx",      // ❌ Exposed!
    "xxxxxxxxxxxxxxxxxxxxxxxx",       // ❌ Exposed!
)
```

**✅ Use environment variables:**
```go
// GOOD - Credentials from environment
client := sdk.NewClient(
    os.Getenv("STORAGE_BASE_URL"),
    os.Getenv("STORAGE_PROJECT_ID"),
    os.Getenv("STORAGE_BUCKET"),
    os.Getenv("STORAGE_ACCESS_KEY"),
    os.Getenv("STORAGE_SECRET_KEY"),
)
```

**✅ Even better - Use secret management:**
```go
// BEST - Use proper secret management
func getStorageClient() (*sdk.Client, error) {
    config, err := secretmanager.GetSecrets("storage-config")
    if err != nil {
        return nil, err
    }

    return sdk.NewClient(
        config.BaseURL,
        config.ProjectID,
        config.Bucket,
        config.AccessKey,
        config.SecretKey,
    ), nil
}
```

### 2. Access Control

**Always verify permissions before generating URLs:**

```go
func generateDownloadURL(client *sdk.Client, fileID string, userID string) (string, error) {
    // ✅ Verify user has access
    file, err := database.GetFile(fileID)
    if err != nil {
        return "", err
    }

    if file.OwnerID != userID && !file.IsPublic {
        return "", errors.New("unauthorized access")
    }

    // Only generate URL after verification
    if file.IsPublic {
        return client.GetPublicObjectURL(file.Filename), nil
    }

    return client.GetObjectURL(file.Filename, 30*time.Minute), nil
}
```

### 3. Presigned URL Expiration

**Use appropriate expiration times:**

```go
// ✅ Short expiration for sensitive operations
uploadURL := client.UploadObjectURL(15 * time.Minute)        // Uploads
deleteURL := client.DeleteObjectURL(filename, 5 * time.Minute) // Deletes

// ✅ Medium expiration for downloads
downloadURL := client.GetObjectURL(filename, 1 * time.Hour)   // Private files

// ✅ Longer only when necessary
shareURL := client.GetObjectURL(filename, 24 * time.Hour)     // Shared links

// ❌ Don't use excessive expiration
badURL := client.GetObjectURL(filename, 30 * 24 * time.Hour)  // 30 days - too long!
```

### 4. Input Validation

**Sanitize and validate inputs:**

```go
func uploadUserFile(client *sdk.Client, filename string, data []byte) error {
    // ✅ Validate filename
    if !isValidFilename(filename) {
        return errors.New("invalid filename")
    }

    // ✅ Check file size
    maxSize := 100 * 1024 * 1024 // 100 MB
    if len(data) > maxSize {
        return errors.New("file too large")
    }

    // ✅ Validate content type
    contentType := http.DetectContentType(data)
    allowedTypes := []string{"image/jpeg", "image/png", "application/pdf"}
    if !contains(allowedTypes, contentType) {
        return errors.New("unsupported file type")
    }

    // Safe to proceed
    return uploadFile(client, filename, data)
}

func isValidFilename(filename string) bool {
    // Check for path traversal
    if strings.Contains(filename, "..") {
        return false
    }
    // Check for suspicious characters
    if strings.ContainsAny(filename, "<>:\"|?*") {
        return false
    }
    // Check length
    if len(filename) > 255 {
        return false
    }
    return true
}
```

### 5. HTTPS Enforcement

**Always use HTTPS:**

```go
// ✅ Use HTTPS in production
client := sdk.NewClient(
    "https://storage.miphiraapis.com",  // HTTPS
    projectID,
    bucket,
    accessKey,
    secretKey,
)

// ❌ Never use HTTP in production
badClient := sdk.NewClient(
    "http://storage.miphiraapis.com",   // Insecure!
    projectID,
    bucket,
    accessKey,
    secretKey,
)
```

### 6. Log Sensitive Operations

**Audit trail for security:**

```go
func generatePresignedURL(client *sdk.Client, fileID, userID, operation string) (string, error) {
    // ✅ Log who generated what URL
    log.Printf("Presigned URL generated: user=%s file=%s operation=%s", 
        userID, fileID, operation)

    switch operation {
    case "download":
        return client.GetObjectURL(fileID, 1*time.Hour), nil
    case "upload":
        return client.UploadObjectURL(15*time.Minute), nil
    case "delete":
        return client.DeleteObjectURL(fileID, 5*time.Minute), nil
    default:
        return "", errors.New("invalid operation")
    }
}
```

## Performance

### 1. Client Reuse

**Reuse client instances:**

```go
// ❌ BAD - Creating client for every request
func handleUpload(w http.ResponseWriter, r *http.Request) {
    client := sdk.NewClient(...)  // ❌ Creates new client every time
    url := client.UploadObjectURL(time.Hour)
    // ...
}

// ✅ GOOD - Singleton client
var storageClient *sdk.Client
var clientOnce sync.Once

func getStorageClient() *sdk.Client {
    clientOnce.Do(func() {
        storageClient = sdk.NewClient(
            os.Getenv("STORAGE_BASE_URL"),
            os.Getenv("STORAGE_PROJECT_ID"),
            os.Getenv("STORAGE_BUCKET"),
            os.Getenv("STORAGE_ACCESS_KEY"),
            os.Getenv("STORAGE_SECRET_KEY"),
        )
    })
    return storageClient
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
    client := getStorageClient()  // ✅ Reuses client
    url := client.UploadObjectURL(time.Hour)
    // ...
}
```

### 2. Concurrent Uploads

**Upload multiple files in parallel:**

```go
func uploadFilesParallel(client *sdk.Client, files []string, maxConcurrent int) error {
    semaphore := make(chan struct{}, maxConcurrent)
    errChan := make(chan error, len(files))
    var wg sync.WaitGroup

    for _, file := range files {
        wg.Add(1)
        go func(filePath string) {
            defer wg.Done()
            semaphore <- struct{}{}        // Acquire
            defer func() { <-semaphore }() // Release

            if err := uploadFile(client, filePath); err != nil {
                errChan <- err
            }
        }(file)
    }

    wg.Wait()
    close(errChan)

    // Collect errors
    var errors []error
    for err := range errChan {
        errors = append(errors, err)
    }

    if len(errors) > 0 {
        return fmt.Errorf("%d uploads failed", len(errors))
    }
    return nil
}

// Usage
uploadFilesParallel(client, files, 5) // Max 5 concurrent uploads
```

### 3. Connection Pooling

**Configure HTTP client for better performance:**

```go
// ✅ Custom HTTP client with connection pooling
var httpClient = &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
        DisableCompression:  false,
    },
}

func uploadWithCustomClient(client *sdk.Client, filePath string) error {
    uploadURL := client.UploadObjectURL(time.Hour)

    // ... prepare multipart form ...

    req, _ := http.NewRequest("POST", uploadURL, body)
    req.Header.Set("Content-Type", contentType)

    // Use custom HTTP client
    resp, err := httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}
```

### 4. Streaming Large Files

**Stream instead of loading into memory:**

```go
// ❌ BAD - Loads entire file into memory
func uploadLargeFileBad(client *sdk.Client, filePath string) error {
    data, err := os.ReadFile(filePath)  // ❌ Entire file in memory
    if err != nil {
        return err
    }
    
    body := bytes.NewBuffer(data)  // ❌ More memory
    // ... upload ...
}

// ✅ GOOD - Streams file
func uploadLargeFileGood(client *sdk.Client, filePath string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    // Create pipe for streaming
    pipeReader, pipeWriter := io.Pipe()
    writer := multipart.NewWriter(pipeWriter)

    // Stream file in goroutine
    go func() {
        defer pipeWriter.Close()
        part, _ := writer.CreateFormFile("file", filepath.Base(filePath))
        io.Copy(part, file)  // ✅ Streams data
        writer.Close()
    }()

    uploadURL := client.UploadObjectURL(time.Hour)
    req, _ := http.NewRequest("POST", uploadURL, pipeReader)
    req.Header.Set("Content-Type", writer.FormDataContentType())

    resp, err := httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}
```

### 5. Caching File URLs

**Cache public URLs to reduce SDK calls:**

```go
type URLCache struct {
    cache map[string]string
    mu    sync.RWMutex
}

func NewURLCache() *URLCache {
    return &URLCache{
        cache: make(map[string]string),
    }
}

func (c *URLCache) GetURL(client *sdk.Client, filename string) string {
    // Check cache first
    c.mu.RLock()
    if url, ok := c.cache[filename]; ok {
        c.mu.RUnlock()
        return url
    }
    c.mu.RUnlock()

    // Generate and cache
    c.mu.Lock()
    defer c.mu.Unlock()

    // Double-check after acquiring write lock
    if url, ok := c.cache[filename]; ok {
        return url
    }

    url := client.GetPublicObjectURL(filename)
    c.cache[filename] = url
    return url
}

// Usage
var urlCache = NewURLCache()

func getProductImageURL(filename string) string {
    return urlCache.GetURL(storageClient, filename)
}
```

## Error Handling

### 1. Retry Logic

**Implement exponential backoff:**

```go
func uploadWithRetry(client *sdk.Client, filePath string, maxRetries int) error {
    var lastErr error

    for attempt := 0; attempt <= maxRetries; attempt++ {
        if attempt > 0 {
            // Exponential backoff: 1s, 2s, 4s, 8s...
            backoff := time.Duration(1<<uint(attempt-1)) * time.Second
            log.Printf("Retry attempt %d/%d after %v", attempt, maxRetries, backoff)
            time.Sleep(backoff)
        }

        err := uploadFile(client, filePath)
        if err == nil {
            return nil // Success
        }

        // Don't retry on client errors
        if isClientError(err) {
            return err
        }

        lastErr = err
    }

    return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

func isClientError(err error) bool {
    // Don't retry 4xx errors (client fault)
    if httpErr, ok := err.(*HTTPError); ok {
        return httpErr.StatusCode >= 400 && httpErr.StatusCode < 500
    }
    return false
}
```

### 2. Graceful Degradation

**Handle failures gracefully:**

```go
func getFileURL(client *sdk.Client, fileID string) (string, error) {
    // Try to get public URL
    url := client.GetPublicObjectURL(fileID)

    // Verify URL is accessible (optional health check)
    resp, err := http.Head(url)
    if err != nil || resp.StatusCode != 200 {
        // Fallback: Generate presigned URL
        log.Printf("Public URL failed, using presigned: %v", err)
        return client.GetObjectURL(fileID, 1*time.Hour), nil
    }

    return url, nil
}
```

### 3. Comprehensive Error Types

**Create custom error types:**

```go
type StorageError struct {
    Operation  string    // "upload", "download", "delete"
    Filename   string
    StatusCode int
    Message    string
    Timestamp  time.Time
    Retryable  bool
}

func (e *StorageError) Error() string {
    return fmt.Sprintf("storage %s failed for %s: %s (status: %d, retryable: %v)",
        e.Operation, e.Filename, e.Message, e.StatusCode, e.Retryable)
}

func uploadFileWithError(client *sdk.Client, filename string) error {
    // ... upload logic ...

    if resp.StatusCode != 201 {
        return &StorageError{
            Operation:  "upload",
            Filename:   filename,
            StatusCode: resp.StatusCode,
            Message:    resp.Status,
            Timestamp:  time.Now(),
            Retryable:  resp.StatusCode >= 500, // 5xx are retryable
        }
    }

    return nil
}

// Usage
if err := uploadFileWithError(client, "file.jpg"); err != nil {
    if storageErr, ok := err.(*StorageError); ok {
        log.Printf("Storage error: %s", storageErr)
        if storageErr.Retryable {
            // Retry logic
        } else {
            // Alert user
        }
    }
}
```

## Resource Management

### 1. Always Close Readers

**Prevent resource leaks:**

```go
// ✅ GOOD - Proper cleanup
func downloadFile(url string) error {
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()  // ✅ Always close

    file, err := os.Create("output.jpg")
    if err != nil {
        return err
    }
    defer file.Close()  // ✅ Always close

    _, err = io.Copy(file, resp.Body)
    return err
}
```

### 2. Context Usage

**Use context for cancellation:**

```go
func uploadWithContext(ctx context.Context, client *sdk.Client, filePath string) error {
    uploadURL := client.UploadObjectURL(time.Hour)

    // ... prepare file ...

    req, err := http.NewRequest("POST", uploadURL, body)
    if err != nil {
        return err
    }

    // ✅ Add context to request
    req = req.WithContext(ctx)

    resp, err := httpClient.Do(req)
    if err != nil {
        // Check if canceled
        if ctx.Err() == context.Canceled {
            return errors.New("upload canceled by user")
        }
        return err
    }
    defer resp.Body.Close()

    return nil
}

// Usage with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

err := uploadWithContext(ctx, client, "large-file.mp4")
```

### 3. Cleanup Temporary Files

**Remove temp files after upload:**

```go
func uploadAndCleanup(client *sdk.Client, filePath string) error {
    // Create temp file
    tmpFile, err := os.CreateTemp("", "upload-*")
    if err != nil {
        return err
    }
    tmpPath := tmpFile.Name()
    tmpFile.Close()

    // ✅ Always cleanup
    defer os.Remove(tmpPath)

    // Process and upload file
    if err := processFile(filePath, tmpPath); err != nil {
        return err
    }

    return uploadFile(client, tmpPath)
}
```

## Production Deployment

### 1. Configuration Management

**Organize configuration:**

```go
type StorageConfig struct {
    BaseURL    string
    ProjectID  string
    Bucket     string
    AccessKey  string
    SecretKey  string
    MaxRetries int
    Timeout    time.Duration
}

func LoadConfig() (*StorageConfig, error) {
    return &StorageConfig{
        BaseURL:    requireEnv("STORAGE_BASE_URL"),
        ProjectID:  requireEnv("STORAGE_PROJECT_ID"),
        Bucket:     requireEnv("STORAGE_BUCKET"),
        AccessKey:  requireEnv("STORAGE_ACCESS_KEY"),
        SecretKey:  requireEnv("STORAGE_SECRET_KEY"),
        MaxRetries: getEnvInt("STORAGE_MAX_RETRIES", 3),
        Timeout:    getEnvDuration("STORAGE_TIMEOUT", 30*time.Second),
    }, nil
}

func requireEnv(key string) string {
    value := os.Getenv(key)
    if value == "" {
        log.Fatalf("Required environment variable %s not set", key)
    }
    return value
}
```

### 2. Health Checks

**Implement readiness/liveness probes:**

```go
func healthCheck(client *sdk.Client) error {
    // Generate a test URL
    testURL := client.GetPublicObjectURL("health-check.txt")

    // Try to access it
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    req, _ := http.NewRequestWithContext(ctx, "HEAD", testURL, nil)
    resp, err := httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("storage unhealthy: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 500 {
        return fmt.Errorf("storage unhealthy: status %d", resp.StatusCode)
    }

    return nil
}

// HTTP handler
func healthHandler(w http.ResponseWriter, r *http.Request) {
    if err := healthCheck(storageClient); err != nil {
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
        return
    }
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
```

### 3. Graceful Shutdown

**Cleanup on shutdown:**

```go
func main() {
    client := sdk.NewClient(...)

    // Setup graceful shutdown
    done := make(chan os.Signal, 1)
    signal.Notify(done, os.Interrupt, syscall.SIGTERM)

    // Start server
    server := &http.Server{Addr: ":8080"}
    go server.ListenAndServe()

    // Wait for signal
    <-done
    log.Println("Shutting down...")

    // Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Stop accepting new requests
    if err := server.Shutdown(ctx); err != nil {
        log.Printf("Server shutdown error: %v", err)
    }

    // Wait for in-flight uploads to complete
    // (if you're tracking them)

    log.Println("Shutdown complete")
}
```

## Monitoring & Logging

### 1. Structured Logging

**Log important operations:**

```go
type LogEntry struct {
    Timestamp time.Time
    Operation string
    FileID    string
    UserID    string
    Status    string
    Duration  time.Duration
    Error     string
}

func uploadWithLogging(client *sdk.Client, fileID, userID string) error {
    start := time.Now()
    entry := LogEntry{
        Timestamp: start,
        Operation: "upload",
        FileID:    fileID,
        UserID:    userID,
    }

    err := uploadFile(client, fileID)

    entry.Duration = time.Since(start)
    if err != nil {
        entry.Status = "error"
        entry.Error = err.Error()
    } else {
        entry.Status = "success"
    }

    logJSON, _ := json.Marshal(entry)
    log.Println(string(logJSON))

    return err
}
```

### 2. Metrics Collection

**Track important metrics:**

```go
type Metrics struct {
    UploadCount    prometheus.Counter
    UploadDuration prometheus.Histogram
    UploadErrors   prometheus.Counter
    UploadSize     prometheus.Histogram
}

func initMetrics() *Metrics {
    return &Metrics{
        UploadCount: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "storage_uploads_total",
            Help: "Total number of file uploads",
        }),
        UploadDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name:    "storage_upload_duration_seconds",
            Help:    "Upload duration in seconds",
            Buckets: prometheus.DefBuckets,
        }),
        UploadErrors: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "storage_upload_errors_total",
            Help: "Total number of upload errors",
        }),
        UploadSize: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name:    "storage_upload_size_bytes",
            Help:    "Upload size in bytes",
            Buckets: prometheus.ExponentialBuckets(1024, 2, 20),
        }),
    }
}

func uploadWithMetrics(metrics *Metrics, client *sdk.Client, fileID string, size int64) error {
    start := time.Now()

    err := uploadFile(client, fileID)

    metrics.UploadCount.Inc()
    metrics.UploadDuration.Observe(time.Since(start).Seconds())
    metrics.UploadSize.Observe(float64(size))

    if err != nil {
        metrics.UploadErrors.Inc()
    }

    return err
}
```

## Cost Optimization

### 1. Optimize File Sizes

**Compress before uploading:**

```go
func uploadCompressed(client *sdk.Client, filePath string) error {
    // Read original file
    data, err := os.ReadFile(filePath)
    if err != nil {
        return err
    }

    // Compress
    var buf bytes.Buffer
    gzWriter := gzip.NewWriter(&buf)
    gzWriter.Write(data)
    gzWriter.Close()

    compressed := buf.Bytes()
    savings := float64(len(data)-len(compressed)) / float64(len(data)) * 100

    log.Printf("Compressed %s: %.2f%% smaller (%d → %d bytes)",
        filePath, savings, len(data), len(compressed))

    // Upload compressed file
    return uploadFromBytes(client, compressed, filePath+".gz")
}
```

### 2. Batch Operations

**Reduce API calls:**

```go
// ❌ BAD - One request per file
func generateURLsIndividually(client *sdk.Client, files []string) []string {
    urls := make([]string, len(files))
    for i, file := range files {
        urls[i] = client.GetPublicObjectURL(file)  // ❌ N calls
    }
    return urls
}

// ✅ GOOD - Cache and reuse
func generateURLsBatch(client *sdk.Client, files []string) []string {
    // Public URLs are deterministic, can be cached
    urls := make([]string, len(files))
    for i, file := range files {
        urls[i] = client.GetPublicObjectURL(file)  // Same result every time
    }
    return urls
}
```

### 3. Delete Unused Files

**Cleanup old files:**

```go
func cleanupOldFiles(client *sdk.Client, days int) error {
    cutoff := time.Now().AddDate(0, 0, -days)

    // Get files older than cutoff
    oldFiles, err := database.GetFilesOlderThan(cutoff)
    if err != nil {
        return err
    }

    log.Printf("Found %d files to delete", len(oldFiles))

    for _, file := range oldFiles {
        deleteURL := client.DeleteObjectURL(file.Filename, 5*time.Minute)

        req, _ := http.NewRequest("DELETE", deleteURL, nil)
        resp, err := httpClient.Do(req)
        if err != nil {
            log.Printf("Failed to delete %s: %v", file.Filename, err)
            continue
        }
        resp.Body.Close()

        // Remove from database
        database.DeleteFile(file.ID)
    }

    return nil
}
```

## Summary Checklist

### Security ✅
- [ ] Never hardcode credentials
- [ ] Use environment variables or secret management
- [ ] Verify permissions before generating URLs
- [ ] Use appropriate presigned URL expiration times
- [ ] Validate all inputs (filenames, file sizes, content types)
- [ ] Always use HTTPS
- [ ] Log sensitive operations for audit trail

### Performance ⚡
- [ ] Reuse client instances (singleton pattern)
- [ ] Use concurrent uploads for multiple files
- [ ] Configure HTTP client with connection pooling
- [ ] Stream large files instead of loading into memory
- [ ] Cache public URLs to reduce SDK calls

### Reliability 🛡️
- [ ] Implement retry logic with exponential backoff
- [ ] Handle errors gracefully with fallbacks
- [ ] Use context for timeouts and cancellation
- [ ] Always close file handles and HTTP responses
- [ ] Cleanup temporary files

### Production 🚀
- [ ] Proper configuration management
- [ ] Implement health checks
- [ ] Graceful shutdown handling
- [ ] Structured logging
- [ ] Metrics collection
- [ ] Cost optimization (compression, cleanup)

## Next Steps

- 🏠 [Getting Started](getting-started.md) - Return to basics
- 📖 [Upload Guide](upload-guide.md) - Detailed upload examples
- 🔐 [Public vs Presigned](public-vs-presigned.md) - URL comparison
