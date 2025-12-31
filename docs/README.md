# Miphira Object Storage SDK Documentation

Welcome to the comprehensive documentation for the Miphira Object Storage Go SDK!

## 📚 Documentation Structure

### 1. [Getting Started](getting-started.md)
**Perfect for:** First-time users, quick setup

Learn the basics and get your first upload/download working in minutes.

**Topics:**
- Installation and prerequisites
- Getting your credentials (Project ID, Bucket, API Keys)
- Your first upload example
- Your first download example
- Common issues and troubleshooting

**Estimated time:** 10-15 minutes

---

### 2. [Upload Guide](upload-guide.md)
**Perfect for:** Developers implementing file upload features

Comprehensive examples for every upload scenario you'll encounter.

**Topics:**
- Basic upload
- Upload with custom metadata
- Upload with folder organization (nested folders)
- Multiple file upload (concurrent)
- Upload from memory (without saving to disk)
- Progress tracking for large files
- Robust error handling with retries
- Understanding server responses and UUID filenames

**Estimated time:** 20-30 minutes

---

### 3. [Public vs Presigned URLs](public-vs-presigned.md)
**Perfect for:** Architects, security-conscious developers

Understand when to use public URLs vs presigned URLs for optimal performance and security.

**Topics:**
- Quick comparison table
- Public URLs explained (when, why, how)
- Presigned URLs explained (when, why, how)
- Real-world use case scenarios (e-commerce, SaaS, social media)
- Security considerations
- Performance comparison
- Best practices for choosing

**Estimated time:** 15-20 minutes

---

### 4. [Best Practices](best-practices.md)
**Perfect for:** Production deployments, experienced developers

Production-ready patterns for security, performance, and reliability.

**Topics:**
- **Security:** Credential management, access control, input validation
- **Performance:** Client reuse, concurrent uploads, connection pooling, streaming
- **Error Handling:** Retry logic with exponential backoff, graceful degradation
- **Resource Management:** Cleanup, context usage, memory management
- **Production Deployment:** Configuration, health checks, graceful shutdown
- **Monitoring & Logging:** Structured logging, metrics collection
- **Cost Optimization:** Compression, batch operations, cleanup strategies

**Estimated time:** 30-40 minutes

---

## 🚀 Quick Navigation

### I want to...

**...get started quickly**
→ [Getting Started](getting-started.md)

**...upload a file with metadata**
→ [Upload Guide - Upload with Metadata](upload-guide.md#upload-with-metadata)

**...organize files in folders**
→ [Upload Guide - Folder Organization](upload-guide.md#upload-with-folder-organization)

**...upload multiple files at once**
→ [Upload Guide - Multiple Files](upload-guide.md#multiple-file-upload)

**...track upload progress**
→ [Upload Guide - Progress Tracking](upload-guide.md#progress-tracking)

**...understand public vs presigned URLs**
→ [Public vs Presigned URLs](public-vs-presigned.md)

**...secure my production deployment**
→ [Best Practices - Security](best-practices.md#security)

**...improve upload performance**
→ [Best Practices - Performance](best-practices.md#performance)

**...implement retry logic**
→ [Best Practices - Error Handling](best-practices.md#error-handling)

**...deploy to production**
→ [Best Practices - Production Deployment](best-practices.md#production-deployment)

---

## 📖 Learning Path

### Beginner Path (Total: ~30 minutes)

1. **[Getting Started](getting-started.md)** - Set up and first examples (15 min)
2. **[Upload Guide - Basic Upload](upload-guide.md#basic-upload)** - Simple upload (5 min)
3. **[Public vs Presigned - Quick Comparison](public-vs-presigned.md#quick-comparison)** - Understand URL types (10 min)

### Intermediate Path (Total: ~60 minutes)

1. Complete Beginner Path (30 min)
2. **[Upload Guide - Metadata & Folders](upload-guide.md)** - Advanced uploads (15 min)
3. **[Public vs Presigned - Use Cases](public-vs-presigned.md#use-case-scenarios)** - Real scenarios (15 min)

### Advanced Path (Total: ~90 minutes)

1. Complete Intermediate Path (60 min)
2. **[Best Practices - Security](best-practices.md#security)** - Secure your app (15 min)
3. **[Best Practices - Performance](best-practices.md#performance)** - Optimize performance (15 min)

### Production Path (Total: ~2 hours)

1. Complete Advanced Path (90 min)
2. **[Best Practices - Full Guide](best-practices.md)** - Production-ready (30 min)

---

## 💡 Key Concepts

### Server-Generated Filenames

**Important:** When you upload a file, the server generates a new UUID-based filename.

```
Your upload: "photo.jpg"
Server stores: "8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg"
```

**Always use the filename from the upload response!**

```go
// Upload file
var fileResp sdk.FileResponse
json.NewDecoder(resp.Body).Decode(&fileResp)

// Use URL from response (contains server-generated UUID filename)
correctURL := fileResp.URL
```

📖 Learn more: [Upload Guide - Understanding Server Response](upload-guide.md#understanding-server-response)

---

### Public URLs vs Presigned URLs

**Public URLs:** Simple, permanent, no authentication
```
Best for: Static assets, public images, CDN content
```

**Presigned URLs:** Secure, time-limited, HMAC-signed
```
Best for: Private files, temporary access, user documents
```

📖 Learn more: [Public vs Presigned URLs Guide](public-vs-presigned.md)

---

### URL Structure

**Public URL Format:**
```
{baseURL}/api/v1/public/projects/{projectId}/buckets/{bucket}/{filename}
```

**Presigned URL Format:**
```
{baseURL}/api/v1/projects/{projectId}/buckets/{bucket}/objects/{filename}?X-Mos-AccessKey={key}&X-Mos-Expires={time}&X-Mos-Signature={sig}
```

---

## 🔧 Common Tasks

### Basic Upload
```go
uploadURL := client.UploadObjectURL(time.Hour)
// ... multipart form upload ...
var fileResp sdk.FileResponse
json.NewDecoder(resp.Body).Decode(&fileResp)
fmt.Println(fileResp.URL) // Use this URL!
```

### Generate Public URL
```go
publicURL := client.GetPublicObjectURL("8aabd7f7-...-db66069746e7.jpg")
// No authentication needed, never expires
```

### Generate Presigned URL
```go
downloadURL := client.GetObjectURL("8aabd7f7-...-db66069746e7.jpg", 30*time.Minute)
// Secure, expires after 30 minutes
```

---

## 🆘 Troubleshooting

### "File not found" after upload
**Problem:** Using original filename instead of server-generated UUID

**Solution:** Parse `FileResponse` and use `fileResp.URL`

📖 [Getting Started - Common Issues](getting-started.md#common-issues)

### "Invalid signature" error
**Problem:** Wrong secret key or expired URL

**Solution:** Verify credentials, regenerate URL

📖 [Getting Started - Common Issues](getting-started.md#common-issues)

### "Unauthorized" error
**Problem:** Missing or incorrect access key

**Solution:** Check `STORAGE_ACCESS_KEY` environment variable

📖 [Getting Started - Common Issues](getting-started.md#common-issues)

---

## 📞 Support

- 📧 **Email:** support@miphira.com
- 📖 **Documentation:** https://docs.miphira.com
- 🐛 **Issues:** https://github.com/miphira/go-client-sdk/issues
- 💬 **Community:** https://community.miphira.com

---

## 🔗 External Resources

- [Go SDK GitHub Repository](https://github.com/miphira/go-client-sdk)
- [Miphira Object Storage API Documentation](https://docs.miphira.com/storage)
- [Go Reference](https://pkg.go.dev/github.com/miphira/go-client-sdk)

---

## ✨ Examples Repository

Find complete, runnable examples in the [examples directory](../examples/) (coming soon):
- Basic upload/download
- User file management system
- Image gallery with CDN
- Secure document sharing
- Batch operations

---

Happy coding! 🚀
