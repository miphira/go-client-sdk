# Public URLs vs Presigned URLs

A comprehensive guide to understanding when and how to use public URLs versus presigned URLs in Miphira Object Storage.

## Table of Contents

- [Quick Comparison](#quick-comparison)
- [Public URLs](#public-urls)
- [Presigned URLs](#presigned-urls)
- [Use Case Scenarios](#use-case-scenarios)
- [Security Considerations](#security-considerations)
- [Performance Comparison](#performance-comparison)
- [Best Practices](#best-practices)

## Quick Comparison

| Feature | Public URL | Presigned URL |
|---------|------------|---------------|
| **Authentication** | ❌ None required | ✅ HMAC-SHA256 signature |
| **Expiration** | ♾️ Never expires | ⏱️ Configurable (e.g., 1 hour) |
| **URL Length** | 🔗 Short | 🔗🔗 Long (includes signature) |
| **Security** | 🌐 Anyone can access | 🔒 Access controlled |
| **Use Case** | Static public assets | Private/temporary access |
| **Beta Mode** | ✅ Available | ✅ Available |
| **Production** | Public buckets only | All buckets |
| **CDN Friendly** | ✅ Perfect | ⚠️ Cache challenges |
| **Share Easily** | ✅ Yes | ⚠️ URL expires |

## Public URLs

### What are Public URLs?

Public URLs provide direct, unauthenticated access to objects. They're simple, permanent links that anyone can use.

**URL Format:**
```
{baseURL}/api/v1/public/projects/{projectId}/buckets/{bucketName}/{filename}
```

**Example:**
```
https://storage.miphiraapis.com/api/v1/public/projects/550e8400-e29b-41d4-a716-446655440000/buckets/images/8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg
```

### How to Generate

```go
client := sdk.NewClient(
    "https://storage.miphiraapis.com",
    "550e8400-e29b-41d4-a716-446655440000",
    "images",
    "MOS_xxxxxxxxxxxxxxxxxxxx",
    "xxxxxxxxxxxxxxxxxxxxxxxx",
)

// Simple one-liner
publicURL := client.GetPublicObjectURL("8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg")

// No expiration, no signature, works immediately
fmt.Println(publicURL)
```

### When to Use Public URLs

#### ✅ Perfect For:

**1. Static Website Assets**
```go
// Logo, icons, CSS, JS files
logoURL := client.GetPublicObjectURL("logo.png")
iconURL := client.GetPublicObjectURL("icons/menu.svg")
styleURL := client.GetPublicObjectURL("styles/main.css")

// Use directly in HTML
<img src="{logoURL}" alt="Company Logo">
<link rel="stylesheet" href="{styleURL}">
```

**2. Public Product Images**
```go
// E-commerce product photos
productURL := client.GetPublicObjectURL("products/12345.jpg")

// Anyone can view without authentication
<img src="{productURL}" alt="Product">
```

**3. Blog/Content Images**
```go
// Blog post featured images
imageURL := client.GetPublicObjectURL("blog/2024/featured-image.jpg")

// Embed in markdown or HTML
![Article Image]({imageURL})
```

**4. CDN Integration**
```go
// Perfect for CDN caching
cdnURL := client.GetPublicObjectURL("assets/hero-image.jpg")

// CDN can cache indefinitely since URL never expires
```

**5. Email Marketing**
```go
// Images in marketing emails
bannerURL := client.GetPublicObjectURL("marketing/banner.jpg")

// Email clients can load without auth
<img src="{bannerURL}" width="600">
```

**6. Public Documents**
```go
// Terms of service, privacy policy
termsURL := client.GetPublicObjectURL("legal/terms.pdf")
privacyURL := client.GetPublicObjectURL("legal/privacy.pdf")

// Share publicly on website
<a href="{termsURL}">Terms of Service</a>
```

**7. Social Media Sharing**
```go
// Images shared on social platforms
ogImageURL := client.GetPublicObjectURL("og-images/homepage.jpg")

// Use in Open Graph tags
<meta property="og:image" content="{ogImageURL}">
```

#### ❌ Not Suitable For:

- Private user data
- Sensitive documents
- Content requiring access control
- Time-limited downloads
- Pay-per-view content
- User-uploaded private files

### Advantages

1. **Simple** - No authentication needed
2. **Fast** - No signature verification overhead
3. **CDN-Friendly** - Perfect for caching
4. **Permanent** - URLs never expire
5. **Short** - Easy to share and embed
6. **Direct Access** - Works in browsers, cURL, wget

### Disadvantages

1. **No Access Control** - Anyone with the URL can access
2. **No Expiration** - Can't revoke access easily
3. **Security Risk** - If URL leaks, anyone can access
4. **Beta Only** - In production, only works for public buckets

## Presigned URLs

### What are Presigned URLs?

Presigned URLs provide temporary, authenticated access to objects using cryptographic signatures.

**URL Format:**
```
{baseURL}/api/v1/projects/{projectId}/buckets/{bucketName}/objects/{filename}?X-Mos-AccessKey={key}&X-Mos-Expires={timestamp}&X-Mos-Signature={signature}
```

**Example:**
```
https://storage.miphiraapis.com/api/v1/projects/550e8400.../buckets/docs/objects/report.pdf?X-Mos-AccessKey=MOS_xxx&X-Mos-Expires=1735689600&X-Mos-Signature=xyz...
```

### How to Generate

```go
client := sdk.NewClient(
    "https://storage.miphiraapis.com",
    "550e8400-e29b-41d4-a716-446655440000",
    "documents",
    "MOS_xxxxxxxxxxxxxxxxxxxx",
    "xxxxxxxxxxxxxxxxxxxxxxxx",
)

// Download (GET)
downloadURL := client.GetObjectURL("report.pdf", 15*time.Minute)

// Upload (POST)
uploadURL := client.UploadObjectURL(1*time.Hour)

// Delete (DELETE)
deleteURL := client.DeleteObjectURL("old-file.pdf", 5*time.Minute)
```

### When to Use Presigned URLs

#### ✅ Perfect For:

**1. Private User Files**
```go
// User documents, photos, private data
userFileURL := client.GetObjectURL("users/123/document.pdf", 30*time.Minute)

// Only this user gets this temporary link
// Expires after 30 minutes
```

**2. Secure File Downloads**
```go
// Purchased content, premium files
purchasedURL := client.GetObjectURL("premium/ebook.pdf", 1*time.Hour)

// Generate only after payment verification
// Expires after 1 hour to prevent sharing
```

**3. Temporary Upload Links**
```go
// Let users upload directly to storage
uploadURL := client.UploadObjectURL(15*time.Minute)

// Send to frontend
// User can upload directly without backend proxy
// Expires after 15 minutes
```

**4. Time-Limited Sharing**
```go
// Share files temporarily
shareURL := client.GetObjectURL("shared/presentation.pdf", 24*time.Hour)

// Perfect for "Share for 24 hours" feature
// Auto-expires, no manual cleanup
```

**5. Medical/Legal Documents**
```go
// HIPAA, GDPR compliant file access
medicalURL := client.GetObjectURL("medical/patient-123.pdf", 10*time.Minute)

// Short expiration for compliance
// Audit trail via access logs
```

**6. Backup Downloads**
```go
// Database backups, system exports
backupURL := client.GetObjectURL("backups/db-2024-12-31.sql.gz", 2*time.Hour)

// Temporary access for administrators
// Expires automatically
```

**7. API File Delivery**
```go
// Generate URLs in API responses
func getDownloadLink(w http.ResponseWriter, r *http.Request) {
    // Verify user permission first
    if !userHasAccess(userID, fileID) {
        http.Error(w, "forbidden", 403)
        return
    }

    // Generate temporary URL
    url := client.GetObjectURL(filename, 30*time.Minute)
    
    json.NewEncoder(w).Encode(map[string]string{
        "download_url": url,
        "expires_in":   "30 minutes",
    })
}
```

#### ❌ Not Suitable For:

- Public website assets
- CDN-delivered content
- Permanent links
- Email signatures/logos
- Social media images

### Advantages

1. **Secure** - Cryptographic signature verification
2. **Time-Limited** - Automatic expiration
3. **Access Control** - Generate only for authorized users
4. **Revocable** - Old URLs expire automatically
5. **Permission-Based** - read/write/delete permissions
6. **Audit-Friendly** - Track who generated what

### Disadvantages

1. **Complex** - Requires signature generation
2. **Long URLs** - Not human-readable
3. **Expiration** - Need to regenerate periodically
4. **CDN Challenges** - Can't cache effectively
5. **Clock Sync** - Server time must be accurate

## Use Case Scenarios

### Scenario 1: E-Commerce Platform

**Product Images (Public)**
```go
// Product catalog images - use public URLs
productImage := client.GetPublicObjectURL("products/12345.jpg")
// ✅ Fast, cacheable, CDN-friendly
// Used in: Product listings, search results, recommendations
```

**Invoice PDFs (Presigned)**
```go
// User invoices - use presigned URLs
invoiceURL := client.GetObjectURL("invoices/user123/inv-456.pdf", 1*time.Hour)
// ✅ Secure, private, time-limited
// Generated after user logs in
```

### Scenario 2: Social Media Platform

**Profile Pictures (Public)**
```go
// Public profile avatars
avatarURL := client.GetPublicObjectURL("avatars/user123.jpg")
// ✅ Anyone can view, perfect for social sharing
```

**Private Messages/Attachments (Presigned)**
```go
// DM attachments
attachmentURL := client.GetObjectURL("messages/private/attach123.jpg", 24*time.Hour)
// ✅ Only sender and recipient can access
```

### Scenario 3: SaaS Application

**Marketing Assets (Public)**
```go
// Landing page images, logos
heroImage := client.GetPublicObjectURL("marketing/hero.jpg")
// ✅ Fast loading, CDN cached
```

**User Exports (Presigned)**
```go
// Data exports, CSV downloads
exportURL := client.GetObjectURL("exports/user123/data.csv", 2*time.Hour)
// ✅ Temporary, secure access
// Email to user, expires after 2 hours
```

### Scenario 4: Education Platform

**Course Thumbnails (Public)**
```go
// Course preview images
thumbURL := client.GetPublicObjectURL("courses/thumbnail-101.jpg")
// ✅ Public catalog browsing
```

**Course Videos (Presigned)**
```go
// Premium video content
videoURL := client.GetObjectURL("courses/video-101-lesson1.mp4", 3*time.Hour)
// ✅ Only enrolled students
// Expires after viewing session
```

## Security Considerations

### Public URLs

**Risks:**
- 🚨 URLs can be guessed (UUID makes it hard but not impossible)
- 🚨 Once shared, hard to revoke
- 🚨 Search engines may index
- 🚨 Anyone with link has permanent access

**Mitigation:**
```go
// 1. Use UUIDs (already done by server)
// 2. Don't expose public URLs for sensitive data
// 3. Use robots.txt to prevent indexing
// 4. Monitor access logs for abuse

// Good: Public logo
logoURL := client.GetPublicObjectURL("logo.png")

// Bad: Private user photo
userPhotoURL := client.GetPublicObjectURL("user-private-photo.jpg") // ❌ Don't do this!
```

### Presigned URLs

**Protections:**
- ✅ Signature verification prevents tampering
- ✅ Expires automatically
- ✅ Tied to specific permissions
- ✅ Requires valid access key

**Best Practices:**
```go
// 1. Short expiration times
uploadURL := client.UploadObjectURL(15*time.Minute) // ✅ Short

// 2. Verify permissions before generating
if !userOwnsFile(userID, fileID) {
    return errors.New("unauthorized")
}
downloadURL := client.GetObjectURL(filename, 30*time.Minute)

// 3. Use HTTPS
// SDK automatically uses HTTPS if baseURL starts with https://

// 4. Log URL generation
log.Printf("Generated presigned URL for user %s, file %s", userID, fileID)
```

## Performance Comparison

### Public URLs

**Speed:**
- ✅ No signature verification (fastest)
- ✅ CDN cacheable (even faster with cache)
- ✅ Direct file serving

**Bandwidth:**
- ✅ Short URLs save bandwidth
- ✅ Cache reduces origin requests

**Example:**
```go
// Public URL: ~100-150 chars
publicURL := client.GetPublicObjectURL("image.jpg")
// https://storage.miphiraapis.com/api/v1/public/projects/550e8400.../buckets/images/abc.jpg
```

### Presigned URLs

**Speed:**
- ⚠️ Signature verification adds ~1-5ms overhead
- ⚠️ Not CDN-friendly (query params prevent effective caching)
- ✅ Still fast direct file serving

**Bandwidth:**
- ⚠️ Longer URLs (~200-300 chars)

**Example:**
```go
// Presigned URL: ~250-350 chars
presignedURL := client.GetObjectURL("image.jpg", time.Hour)
// https://storage.miphiraapis.com/api/v1/projects/.../objects/abc.jpg?X-Mos-AccessKey=...&X-Mos-Expires=...&X-Mos-Signature=...
```

## Best Practices

### Choosing the Right Approach

```go
// Decision Tree
func getFileURL(fileType string, isPrivate bool, needsExpiration bool) string {
    if !isPrivate && !needsExpiration {
        // Public asset - use public URL
        return client.GetPublicObjectURL(filename)
    }
    
    if isPrivate || needsExpiration {
        // Private or temporary - use presigned URL
        return client.GetObjectURL(filename, expirationTime)
    }
}

// Examples
logoURL := getFileURL("logo", false, false)           // → Public URL
userFileURL := getFileURL("document", true, true)     // → Presigned URL
```

### Hybrid Approach

Use both types strategically:

```go
type File struct {
    ID           string
    Filename     string
    IsPublic     bool
    OwnerID      string
}

func (f *File) GetURL(client *sdk.Client, requestingUserID string) (string, error) {
    // Public files → public URL
    if f.IsPublic {
        return client.GetPublicObjectURL(f.Filename), nil
    }

    // Private files → verify access then generate presigned URL
    if f.OwnerID != requestingUserID {
        return "", errors.New("unauthorized")
    }

    return client.GetObjectURL(f.Filename, 30*time.Minute), nil
}
```

### Migration Strategy

Migrating from presigned to public (or vice versa):

```go
// V1: Everything presigned (more secure but slower)
func getLegacyURL(filename string) string {
    return client.GetObjectURL(filename, 1*time.Hour)
}

// V2: Public for static assets, presigned for user content
func getOptimizedURL(filename string, category string) string {
    staticCategories := map[string]bool{
        "logo": true,
        "icon": true,
        "banner": true,
    }

    if staticCategories[category] {
        return client.GetPublicObjectURL(filename)
    }

    return client.GetObjectURL(filename, 1*time.Hour)
}
```

## Summary

**Use Public URLs when:**
- ✅ Content is truly public
- ✅ Need CDN caching
- ✅ Want simple, permanent links
- ✅ Performance is critical
- ✅ Beta/development mode

**Use Presigned URLs when:**
- ✅ Content is private
- ✅ Need access control
- ✅ Want time-limited access
- ✅ Compliance requirements
- ✅ Production with private data

**In Production:**
- Use public URLs for static assets (logos, images, CSS, JS)
- Use presigned URLs for user data (documents, photos, videos)
- Implement proper access control before generating URLs
- Monitor and log URL generation for security audits

## Next Steps

- ⚡ [Best Practices](best-practices.md) - Security, performance, error handling
- 📖 [Upload Guide](upload-guide.md) - Learn about uploading files
- 🏠 [Getting Started](getting-started.md) - Return to basics
