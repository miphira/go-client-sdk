# Examples

This directory contains practical examples demonstrating how to use the Miphira Object Storage Go SDK.

## Prerequisites

Set up your environment variables in a `.env` file or export them:

```bash
export STORAGE_BASE_URL="https://storage.miphiraapis.com"
export STORAGE_PROJECT_ID="your-project-id"
export STORAGE_BUCKET="your-bucket-name"
export STORAGE_ACCESS_KEY="MOS_your_access_key"
export STORAGE_SECRET_KEY="your_secret_key"
```

## Running Examples

### Basic Upload

Upload a file and get the public URL with server-generated UUID filename:

```bash
cd examples
go run basic-upload.go /path/to/your/file.jpg
```

**Output:**
```
Uploading file: photo.jpg

✅ Upload successful!
File ID: 8aabd7f7-1dbf-4ea4-8918-db66069746e7
Original Name: photo.jpg
Server Filename: 8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg
Size: 1.2 MB
MIME Type: image/jpeg

📍 Public URL:
https://storage.miphiraapis.com/api/v1/public/projects/550e8400-e29b-41d4-a716-446655440000/buckets/images/8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg

💡 Important:
   - Server generated UUID filename: 8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg
   - Always use resp.URL or resp.Name for accessing the file
   - Don't use your original filename (photo.jpg) - it won't work!
```

### What You'll Learn

- **basic-upload.go**: Simple file upload with automatic response parsing
  - Shows the difference between original filename and server-generated UUID filename
  - Demonstrates why you must use `resp.URL` instead of your original filename
  - Displays all response fields (ID, original name, server filename, size, URL)

## Key Takeaways

1. **Server Generates UUID Filenames**: When you upload `photo.jpg`, the server stores it as `8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg`

2. **Always Use Response URL**: The `FileResponse.URL` contains the correct URL with UUID filename:
   ```go
   resp, err := client.Upload("photo.jpg", nil)
   // ✅ CORRECT: Use resp.URL
   fmt.Println(resp.URL)
   
   // ❌ WRONG: Don't use original filename
   wrongURL := client.GetPublicObjectURL("photo.jpg") // Will return 404!
   ```

3. **Store in Database**: Save these fields in your database:
   - `resp.ID` - File ID for API operations
   - `resp.URL` - Complete URL (recommended for direct use)
   - `resp.Name` - Server-generated UUID filename
   - `resp.OriginalName` - Original filename (for display only)

## Common Mistakes

### ❌ Mistake 1: Using Original Filename

```go
resp, err := client.Upload("photo.jpg", nil)
// Wrong: Using original filename
publicURL := client.GetPublicObjectURL("photo.jpg") // 404 Not Found!
```

### ✅ Solution 1: Use Response URL

```go
resp, err := client.Upload("photo.jpg", nil)
// Correct: Use response URL or server filename
publicURL := resp.URL // Works!
// Or extract filename from response
publicURL := client.GetPublicObjectURL(resp.Name) // Works!
```

### ❌ Mistake 2: Not Storing Server Filename

```go
// Upload file
resp, err := client.Upload("photo.jpg", nil)
// Wrong: Only storing original name
db.Save(User{Avatar: "photo.jpg"}) // Can't access later!
```

### ✅ Solution 2: Store Server Filename or URL

```go
// Upload file
resp, err := client.Upload("photo.jpg", nil)
// Correct: Store server-generated filename or full URL
db.Save(User{
    Avatar: resp.Name, // Store UUID filename
    AvatarURL: resp.URL, // Or store full URL
    OriginalName: resp.OriginalName, // For display
})
```

## Next Steps

After running these examples:

1. **Read the Docs**: Check out [../docs/](../docs/) for comprehensive guides
2. **Upload Guide**: Learn advanced features in [../docs/upload-guide.md](../docs/upload-guide.md)
3. **Best Practices**: Production patterns in [../docs/best-practices.md](../docs/best-practices.md)
