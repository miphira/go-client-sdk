package main

import (
	"fmt"
	"log"
	"os"

	sdk "github.com/miphira/go-client-sdk"
)

func main() {
	// Create client from environment variables
	client := sdk.NewClient(
		os.Getenv("STORAGE_BASE_URL"),
		os.Getenv("STORAGE_PROJECT_ID"),
		os.Getenv("STORAGE_BUCKET"),
		os.Getenv("STORAGE_ACCESS_KEY"),
		os.Getenv("STORAGE_SECRET_KEY"),
	)

	// Check if a file was provided as argument
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run basic-upload.go <file-path>")
	}
	filePath := os.Args[1]

	// Upload file
	fmt.Printf("Uploading file: %s\n", filePath)
	resp, err := client.Upload(filePath, nil)
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}

	// Display results
	fmt.Println("\n✅ Upload successful!")
	fmt.Printf("File ID: %s\n", resp.ID)
	fmt.Printf("Original Name: %s\n", resp.OriginalName)
	fmt.Printf("Server Filename: %s\n", resp.Name)
	fmt.Printf("Size: %s\n", resp.SizeFormatted)
	fmt.Printf("MIME Type: %s\n", resp.MimeType)
	fmt.Printf("\n📍 Public URL:\n%s\n", resp.URL)
	fmt.Println("\n💡 Important:")
	fmt.Printf("   - Server generated UUID filename: %s\n", resp.Name)
	fmt.Printf("   - Always use resp.URL or resp.Name for accessing the file\n")
	fmt.Printf("   - Don't use your original filename (%s) - it won't work!\n", resp.OriginalName)
}
