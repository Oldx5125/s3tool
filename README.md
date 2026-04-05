# S3 Tool

[简体中文](README_zh.md) | English

A **single-file, dependency-free, ready-to-use** S3 compatible storage management tool.

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## Features

### Core Functions
- :white_check_mark: Connect to any S3-compatible storage (AWS S3, MinIO, Huawei OBS, etc.)
- :white_check_mark: List all buckets in your account
- :white_check_mark: Bucket selection (no manual entry needed)
- :white_check_mark: Quick bucket switching (no reconnection required)
- :white_check_mark: Folder navigation (breadcrumbs, double-click to enter)
- :white_check_mark: Self-signed SSL certificate support
- :white_check_mark: Auto-restore connection on page refresh
- :white_check_mark: Logout to clear credentials
- :white_check_mark: Command-line argument configuration

### File Operations
- :white_check_mark: Upload files (with folder path support)
- :white_check_mark: Download files
- :white_check_mark: Delete files
- :white_check_mark: Create folders
- :white_check_mark: Delete folders (empty folders only)
- :white_check_mark: Rename files
- :white_check_mark: Move files (cross-directory)
- :white_check_mark: Copy files

### Batch Operations
- :white_check_mark: Multi-select files (checkbox + select all)
- :white_check_mark: Batch upload (multi-file selection + queue progress)
- :white_check_mark: Batch download (download selected files one by one)
- :white_check_mark: Batch delete (concurrent deletion)

### File Preview
- :white_check_mark: Image preview (JPEG, PNG, GIF, WebP, SVG)
- :white_check_mark: Text preview (auto-detect encoding)
- :white_check_mark: Video playback (MP4, WebM)
- :white_check_mark: Audio playback (MP3, WAV, OGG)
- :white_check_mark: PDF preview

### Bucket Management
- :white_check_mark: Configure bucket CORS rules

## Installation

### Build from Source

```bash
# GitHub
git clone https://github.com/Oldx5125/s3tool.git
# or Gitee (faster in China)
git clone https://gitee.com/kylingx/s3tool.git
cd s3tool
go build -o s3tool .

## Quick Start

### Start the Server

```bash
# Default: listen on 0.0.0.0:9090
./s3tool

# Specify port
./s3tool -p 8080

# Localhost only
./s3tool -addr 127.0.0.1

# Specify both port and address
./s3tool -p 8080 -addr 127.0.0.1
```

### Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `-p <port>` | Listen port | 9090 |
| `-addr <addr>` | Listen address | 0.0.0.0 |
| `-h, -help` | Show help | - |
| `-v, -version` | Show version | - |

### Access the UI

Open http://localhost:9090 in your browser.

### Connect to Storage

Enter the following in the interface:

| Field | Description | Example |
|-------|-------------|---------|
| Endpoint | S3 service address | `s3.amazonaws.com` or `10.50.2.45:5080` |
| AK | Access Key | - |
| SK | Secret Key | - |
| SSL | Use HTTPS | Check based on your setup |

**Connection Steps:**
1. Fill in Endpoint, AK, SK
2. Click "List Buckets"
3. Select a bucket from the dropdown
4. Click "Connect"

> Supports URL-format Endpoint, e.g., `https://s3.amazonaws.com`, will be auto-parsed.

## Internationalization

This tool supports **Chinese** and **English** languages (frontend implementation).

### Language Detection

| Method | Description |
|--------|-------------|
| URL Parameter | Add `?lang=en` or `?lang=zh` to the URL |
| Browser Setting | Auto-detect from browser language |

### Examples

For frontend, open:
- Chinese (default): http://localhost:9090
- English: http://localhost:9090?lang=en

## Usage Guide

### File List Operations

- **Single Click** - Select file/folder
- **Double Click Folder** - Enter folder
- **Header Sorting** - Click header to sort by name, size, or modified time (click again to toggle asc/desc)
- **Breadcrumb Navigation** - Click path level for quick jump
- **Go Back** - Return to parent directory

### File Actions

| Action | Description |
|--------|-------------|
| List Buckets | Get all buckets in the account |
| Switch Bucket | Switch to another bucket without reconnecting |
| Upload | Upload to current directory (multi-select supported) |
| Batch Download | Download all selected files |
| Batch Delete | Concurrently delete all selected files |
| New Folder | Create subfolder in current directory |
| Rename | Rename file or folder |
| Move | Move file to another directory |
| Copy | Copy file to another directory |
| Preview | Preview images, text, video, audio, PDF |
| CORS Config | Configure cross-origin rules for the bucket |

### Multi-Select Operations

- Click the checkbox on the left of a file row to select
- Click the checkbox in the header to select all files in current directory
- Selected files can be batch downloaded or deleted

### CORS Configuration

Click "CORS Config" button to manage bucket cross-origin rules:

1. **View Config** - Current CORS config loads automatically when opening
2. **Load Example** - Fill with common CORS configuration template
3. **Save Config** - Save the JSON configuration in the editor
4. **Delete Config** - Clear the bucket's CORS configuration

**CORS Configuration JSON Example:**
```json
{
    "CORSRules": [
        {
            "AllowedOrigin": ["*"],
            "AllowedMethod": ["PUT", "POST", "DELETE"],
            "AllowedHeader": ["*"]
        },
        {
            "AllowedOrigin": ["*"],
            "AllowedMethod": ["GET", "HEAD"]
        }
    ]
}
```

> **Note**: Use singular field names: `AllowedOrigin`, `AllowedMethod`, `AllowedHeader`

## Security Features

- Credentials stored only in browser localStorage
- No credentials stored on server
- Auto-restore connection on page refresh
- Logout clears all credentials

## Tech Stack

- **Go 1.21+** - Runtime
- **Gin** - HTTP framework
- **minio-go v7** - S3 client
- **embed** - Frontend resource embedding
- **slog** - Structured logging

## Build

```bash
# Compile
go build -o s3tool .

# Run tests
go test -v ./...

# Run tests with coverage
go test -v -cover -coverprofile=coverage.out .
go tool cover -func=coverage.out

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o s3tool-linux-amd64 .
GOOS=darwin GOARCH=amd64 go build -o s3tool-darwin-amd64 .
GOOS=windows GOARCH=amd64 go build -o s3tool-windows-amd64.exe .
```

## Project Structure

```
s3tool/
├── main.go           # Application entry point
├── frontend.html     # Frontend UI (embedded)
├── internal/
│   ├── api/          # HTTP handlers & middleware
│   ├── config/       # Configuration management
│   ├── i18n/         # Internationalization
│   ├── models/       # Data models
│   ├── s3client/     # S3 client wrapper
│   └── utils/        # Utility functions
├── go.mod
├── go.sum
└── README.md
```

## License

MIT License - see [LICENSE](LICENSE) file for details.
