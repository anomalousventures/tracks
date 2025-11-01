# Storage

**[← Back to Summary](./0_summary.md)**

## Overview

Tracks provides a unified storage interface supporting multiple backends (S3, R2, local filesystem). The system handles file uploads with streaming to avoid memory issues, content validation, and optional virus scanning integration.

## Goals

- Unified interface for multiple storage backends
- Stream large files directly to storage without memory buffering
- Validate file types by content, not extension
- Virus scanning integration ready
- Presigned URLs for secure direct uploads
- Automatic retry and circuit breaking
- Progress tracking for large uploads

## User Stories

- As a developer, I want to switch storage providers without changing code
- As a user, I want to upload large files without the app crashing
- As a developer, I want files streamed to S3 without loading in memory
- As a security engineer, I want file type validation by content
- As a developer, I want virus scanning before accepting files
- As a user, I want to see upload progress for large files
- As a DevOps engineer, I want to use S3, R2, or local storage

## Storage Interface

```go
// internal/adapters/storage/storage.go
package storage

import (
    "context"
    "io"
    "time"
)

type Storage interface {
    // Save streams data directly to storage
    Save(ctx context.Context, path string, r io.Reader, metadata Metadata) error

    // Open returns a reader for the file
    Open(ctx context.Context, path string) (io.ReadCloser, error)

    // Delete removes the file
    Delete(ctx context.Context, path string) error

    // URL generates a presigned URL for direct access
    URL(ctx context.Context, path string, expires time.Duration) (string, error)

    // UploadURL generates a presigned URL for direct upload
    UploadURL(ctx context.Context, path string, expires time.Duration) (string, error)

    // Exists checks if a file exists
    Exists(ctx context.Context, path string) (bool, error)

    // Size returns the file size
    Size(ctx context.Context, path string) (int64, error)

    // List returns files matching the prefix
    List(ctx context.Context, prefix string) ([]FileInfo, error)
}

type Metadata struct {
    ContentType  string
    CacheControl string
    Tags         map[string]string
}

type FileInfo struct {
    Path         string
    Size         int64
    LastModified time.Time
    ContentType  string
}
```

## S3/R2 Implementation

```go
// internal/adapters/storage/s3.go
package storage

import (
    "context"
    "fmt"
    "io"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go-v2/service/s3/types"
    "github.com/sony/gobreaker"
)

type S3Storage struct {
    client   *s3.Client
    uploader *manager.Uploader
    bucket   string
    cb       *gobreaker.CircuitBreaker[bool]
    region   string
}

func NewS3Storage(client *s3.Client, bucket, region string, cb *gobreaker.CircuitBreaker[bool]) *S3Storage {
    return &S3Storage{
        client:   client,
        uploader: manager.NewUploader(client),
        bucket:   bucket,
        region:   region,
        cb:       cb,
    }
}

func (s *S3Storage) Save(ctx context.Context, path string, r io.Reader, metadata Metadata) error {
    _, err := s.cb.Execute(func() (bool, error) {
        input := &s3.PutObjectInput{
            Bucket: aws.String(s.bucket),
            Key:    aws.String(path),
            Body:   r,
        }

        if metadata.ContentType != "" {
            input.ContentType = aws.String(metadata.ContentType)
        }

        if metadata.CacheControl != "" {
            input.CacheControl = aws.String(metadata.CacheControl)
        }

        if len(metadata.Tags) > 0 {
            tagging := encodeS3Tags(metadata.Tags)
            input.Tagging = aws.String(tagging)
        }

        // Use uploader for automatic multipart upload on large files
        _, err := s.uploader.Upload(ctx, input)
        return true, err
    })

    return err
}

func (s *S3Storage) Open(ctx context.Context, path string) (io.ReadCloser, error) {
    result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(path),
    })
    if err != nil {
        return nil, err
    }

    return result.Body, nil
}

func (s *S3Storage) Delete(ctx context.Context, path string) error {
    _, err := s.cb.Execute(func() (bool, error) {
        _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
            Bucket: aws.String(s.bucket),
            Key:    aws.String(path),
        })
        return true, err
    })

    return err
}

func (s *S3Storage) URL(ctx context.Context, path string, expires time.Duration) (string, error) {
    presigner := s3.NewPresignClient(s.client)
    req, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(path),
    }, func(opts *s3.PresignOptions) {
        opts.Expires = expires
    })

    if err != nil {
        return "", err
    }

    return req.URL, nil
}

func (s *S3Storage) UploadURL(ctx context.Context, path string, expires time.Duration) (string, error) {
    presigner := s3.NewPresignClient(s.client)
    req, err := presigner.PresignPutObject(ctx, &s3.PutObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(path),
    }, func(opts *s3.PresignOptions) {
        opts.Expires = expires
    })

    if err != nil {
        return "", err
    }

    return req.URL, nil
}

func (s *S3Storage) Exists(ctx context.Context, path string) (bool, error) {
    _, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(path),
    })

    if err != nil {
        // Check if it's a not found error
        if isNotFoundError(err) {
            return false, nil
        }
        return false, err
    }

    return true, nil
}

func (s *S3Storage) Size(ctx context.Context, path string) (int64, error) {
    result, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(path),
    })

    if err != nil {
        return 0, err
    }

    return result.ContentLength, nil
}

func (s *S3Storage) List(ctx context.Context, prefix string) ([]FileInfo, error) {
    var files []FileInfo

    paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
        Bucket: aws.String(s.bucket),
        Prefix: aws.String(prefix),
    })

    for paginator.HasMorePages() {
        page, err := paginator.NextPage(ctx)
        if err != nil {
            return nil, err
        }

        for _, obj := range page.Contents {
            files = append(files, FileInfo{
                Path:         *obj.Key,
                Size:         obj.Size,
                LastModified: *obj.LastModified,
            })
        }
    }

    return files, nil
}
```

## Local Filesystem Implementation

```go
// internal/adapters/storage/local.go
package storage

import (
    "context"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "time"
)

type LocalStorage struct {
    basePath string
    baseURL  string
}

func NewLocalStorage(basePath, baseURL string) *LocalStorage {
    return &LocalStorage{
        basePath: basePath,
        baseURL:  baseURL,
    }
}

func (l *LocalStorage) Save(ctx context.Context, path string, r io.Reader, metadata Metadata) error {
    fullPath := filepath.Join(l.basePath, path)

    // Create directory if it doesn't exist
    dir := filepath.Dir(fullPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("create directory: %w", err)
    }

    // Create file
    file, err := os.Create(fullPath)
    if err != nil {
        return fmt.Errorf("create file: %w", err)
    }
    defer file.Close()

    // Copy data
    if _, err := io.Copy(file, r); err != nil {
        return fmt.Errorf("write file: %w", err)
    }

    // Store metadata as extended attributes or separate file
    if metadata.ContentType != "" {
        metaPath := fullPath + ".meta"
        metaFile, err := os.Create(metaPath)
        if err == nil {
            defer metaFile.Close()
            fmt.Fprintf(metaFile, "Content-Type: %s\n", metadata.ContentType)
        }
    }

    return nil
}

func (l *LocalStorage) Open(ctx context.Context, path string) (io.ReadCloser, error) {
    fullPath := filepath.Join(l.basePath, path)
    return os.Open(fullPath)
}

func (l *LocalStorage) Delete(ctx context.Context, path string) error {
    fullPath := filepath.Join(l.basePath, path)
    os.Remove(fullPath + ".meta") // Remove metadata if exists
    return os.Remove(fullPath)
}

func (l *LocalStorage) URL(ctx context.Context, path string, expires time.Duration) (string, error) {
    // For local storage, return a URL that the app will serve
    return fmt.Sprintf("%s/files/%s", l.baseURL, path), nil
}

func (l *LocalStorage) UploadURL(ctx context.Context, path string, expires time.Duration) (string, error) {
    // Local storage doesn't support presigned uploads
    // Return an endpoint that the app handles
    return fmt.Sprintf("%s/upload?path=%s", l.baseURL, path), nil
}

func (l *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
    fullPath := filepath.Join(l.basePath, path)
    _, err := os.Stat(fullPath)
    if err != nil {
        if os.IsNotExist(err) {
            return false, nil
        }
        return false, err
    }
    return true, nil
}

func (l *LocalStorage) Size(ctx context.Context, path string) (int64, error) {
    fullPath := filepath.Join(l.basePath, path)
    info, err := os.Stat(fullPath)
    if err != nil {
        return 0, err
    }
    return info.Size(), nil
}

func (l *LocalStorage) List(ctx context.Context, prefix string) ([]FileInfo, error) {
    var files []FileInfo
    searchPath := filepath.Join(l.basePath, prefix)

    err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if !info.IsDir() {
            relPath, _ := filepath.Rel(l.basePath, path)
            files = append(files, FileInfo{
                Path:         relPath,
                Size:         info.Size(),
                LastModified: info.ModTime(),
            })
        }

        return nil
    })

    return files, err
}
```

## File Upload Handler

```go
// internal/http/handlers/upload_handler.go
package handlers

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/gofrs/uuid/v5"  // Fixed: Using correct UUID package
    "github.com/h2non/filetype"
    "myapp/internal/adapters/storage"
)

type UploadHandler struct {
    storage      storage.Storage
    virusScanner VirusScanner // Optional
}

func NewUploadHandler(storage storage.Storage) *UploadHandler {
    return &UploadHandler{
        storage: storage,
    }
}

func (h *UploadHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
    // Limit request size
    r.Body = http.MaxBytesReader(w, r.Body, 100<<20) // 100MB

    // Parse multipart form with 10MB memory limit
    if err := r.ParseMultipartForm(10 << 20); err != nil {
        http.Error(w, "File too large", http.StatusBadRequest)
        return
    }

    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Error getting file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // ✅ FIXED: Validate file type using magic numbers
    // Read enough bytes for complete magic number detection
    head := make([]byte, 512)  // Changed from 261 to 512 for better detection
    n, err := file.Read(head)
    if err != nil && err != io.EOF {
        http.Error(w, "Error reading file", http.StatusBadRequest)
        return
    }
    head = head[:n]

    // Check file type
    kind, _ := filetype.Match(head)
    if !isAllowedFileType(kind) {
        http.Error(w, "File type not allowed", http.StatusBadRequest)
        return
    }

    // Optional: Virus scanning
    if h.virusScanner != nil {
        // Create temp reader for scanning
        tempFile := io.MultiReader(bytes.NewReader(head), file)
        if infected, err := h.virusScanner.Scan(tempFile); infected || err != nil {
            http.Error(w, "File failed security scan", http.StatusBadRequest)
            return
        }
    }

    // Reset file pointer
    file.Seek(0, io.SeekStart)

    // Generate unique key with proper path structure
    userID := r.Context().Value("user_id").(string)
    fileID := uuid.Must(uuid.NewV7()).String()  // Fixed: Using UUIDv7
    ext := filepath.Ext(header.Filename)
    key := fmt.Sprintf("uploads/%s/%s/%s%s",
        time.Now().Format("2006/01/02"),
        userID,
        fileID,
        ext,
    )

    // Prepare metadata
    metadata := storage.Metadata{
        ContentType:  kind.MIME.Value,
        CacheControl: "max-age=31536000",
        Tags: map[string]string{
            "user_id":       userID,
            "original_name": header.Filename,
            "uploaded_at":   time.Now().Format(time.RFC3339),
        },
    }

    // Stream to storage (no memory buffering)
    if err := h.storage.Save(r.Context(), key, file, metadata); err != nil {
        http.Error(w, "Upload failed", http.StatusInternalServerError)
        return
    }

    // Generate access URL (expires in 7 days)
    url, err := h.storage.URL(r.Context(), key, 7*24*time.Hour)
    if err != nil {
        url = fmt.Sprintf("/files/%s", key)
    }

    // Return response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "id":   fileID,
        "key":  key,
        "url":  url,
        "size": header.Size,
        "type": kind.MIME.Value,
    })
}

func (h *UploadHandler) HandlePresignedUpload(w http.ResponseWriter, r *http.Request) {
    // Generate presigned URL for direct browser upload
    filename := r.URL.Query().Get("filename")
    if filename == "" {
        http.Error(w, "Filename required", http.StatusBadRequest)
        return
    }

    // Validate file extension
    ext := filepath.Ext(filename)
    if !isAllowedExtension(ext) {
        http.Error(w, "File type not allowed", http.StatusBadRequest)
        return
    }

    // Generate key
    userID := r.Context().Value("user_id").(string)
    fileID := uuid.Must(uuid.NewV7()).String()
    key := fmt.Sprintf("uploads/%s/%s/%s%s",
        time.Now().Format("2006/01/02"),
        userID,
        fileID,
        ext,
    )

    // Generate presigned upload URL (expires in 1 hour)
    uploadURL, err := h.storage.UploadURL(r.Context(), key, time.Hour)
    if err != nil {
        http.Error(w, "Failed to generate upload URL", http.StatusInternalServerError)
        return
    }

    // Generate download URL
    downloadURL, _ := h.storage.URL(r.Context(), key, 7*24*time.Hour)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "upload_url":   uploadURL,
        "download_url": downloadURL,
        "key":          key,
        "id":           fileID,
    })
}

func isAllowedFileType(kind filetype.Type) bool {
    allowed := map[string]bool{
        "image/jpeg": true,
        "image/png":  true,
        "image/gif":  true,
        "image/webp": true,
        "application/pdf": true,
    }
    return allowed[kind.MIME.Value]
}

func isAllowedExtension(ext string) bool {
    allowed := map[string]bool{
        ".jpg": true, ".jpeg": true,
        ".png": true, ".gif": true,
        ".webp": true, ".pdf": true,
    }
    return allowed[strings.ToLower(ext)]
}
```

## Virus Scanner Integration

```go
// internal/interfaces/scanner.go
package interfaces

import (
    "io"
)

type VirusScanner interface {
    Scan(r io.Reader) (infected bool, err error)
}
```

```go
// internal/adapters/scanner/clamav.go
package scanner

import (
    "io"
    "myapp/internal/interfaces"
)

// ClamAV implementation
type ClamAVScanner struct {
    endpoint string
}

func NewClamAVScanner(endpoint string) interfaces.VirusScanner {
    return &ClamAVScanner{
        endpoint: endpoint,
    }
}

func (c *ClamAVScanner) Scan(r io.Reader) (bool, error) {
    // Implementation for ClamAV scanning
    // Can use clamd network protocol or command line
    return false, nil
}
```

```go
// internal/adapters/scanner/cloud.go
package scanner

// Placeholder for cloud-based scanning (e.g., AWS Malware Protection)
type CloudScanner struct {
    // Implementation
}
```

## Configuration

```go
// internal/config/storage.go
package config

import (
    "context"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/sony/gobreaker"
    "myapp/internal/adapters/storage"
)

func NewStorage(cfg StorageConfig, cb *gobreaker.CircuitBreaker[bool]) (storage.Storage, error) {
    switch cfg.Provider {
    case "s3":
        awsCfg, err := config.LoadDefaultConfig(context.Background())
        if err != nil {
            return nil, err
        }
        client := s3.NewFromConfig(awsCfg)
        return storage.NewS3Storage(client, cfg.S3Bucket, cfg.S3Region, cb), nil

    case "r2":
        // R2 uses S3 API with custom endpoint
        awsCfg, err := config.LoadDefaultConfig(context.Background())
        if err != nil {
            return nil, err
        }
        client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
            o.BaseEndpoint = aws.String(cfg.S3Endpoint)
        })
        return storage.NewS3Storage(client, cfg.S3Bucket, "auto", cb), nil

    case "local":
        return storage.NewLocalStorage(cfg.LocalPath, cfg.BaseURL), nil

    default:
        return nil, fmt.Errorf("unknown storage provider: %s", cfg.Provider)
    }
}
```

## Progress Tracking

```go
// internal/http/handlers/upload_progress.go
package handlers

// For large file uploads with progress tracking
func (h *UploadHandler) HandleChunkedUpload(w http.ResponseWriter, r *http.Request) {
    // Implementation for chunked upload with progress
    // Uses resumable upload protocols
}
```

## Testing

```go
// internal/adapters/storage/storage_test.go
package storage_test

import (
    "bytes"
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestLocalStorage_SaveAndOpen(t *testing.T) {
    storage := NewLocalStorage("/tmp/test", "http://localhost")

    // Save file
    data := []byte("test content")
    err := storage.Save(context.Background(), "test.txt", bytes.NewReader(data), Metadata{})
    assert.NoError(t, err)

    // Open file
    reader, err := storage.Open(context.Background(), "test.txt")
    assert.NoError(t, err)
    defer reader.Close()

    // Read content
    buf := new(bytes.Buffer)
    buf.ReadFrom(reader)
    assert.Equal(t, "test content", buf.String())

    // Clean up
    storage.Delete(context.Background(), "test.txt")
}
```

## Best Practices

1. **Always stream large files** - Never load entire file into memory
2. **Validate by content, not extension** - Check magic numbers
3. **Use presigned URLs for large uploads** - Direct browser-to-S3 uploads
4. **Set appropriate cache headers** - Immutable content with long cache
5. **Implement virus scanning** - Scan before accepting files
6. **Use circuit breakers** - Protect against storage service failures
7. **Structure paths properly** - Include dates for easier management
8. **Tag uploads** - Track user, time, and purpose

## Next Steps

- Continue to [Observability →](./12_observability.md)
- Back to [← Background Jobs](./10_background_jobs.md)
- Return to [Summary](./0_summary.md)
