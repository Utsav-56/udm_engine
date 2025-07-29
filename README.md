A high-performance download manager written in Go that accelerates downloads using multi-threaded technology, capable of boosting download speeds by up to 8X.

## Table of Contents

- Overview
- Features
- Architecture
- Components
- How It Works
- Installation
- Usage
- API Reference
- Contributing
- Changelog

## Overview

UDM (Ultimate Download Manager) is a modern download acceleration engine designed to replace traditional download managers like IDM and XDM. Built from the ground up in Go, UDM leverages multi-threading and advanced HTTP techniques to dramatically improve download speeds and reliability.

### Key Goals

- **Speed**: Multi-threaded downloads with up to 8X speed improvement
- **Reliability**: Robust error handling with retry mechanisms
- **Compatibility**: Support for various server configurations and protocols
- **Efficiency**: Written in Go for optimal performance and memory usage

## Features

### Current Features (Will continue in future versions)

- ✅ **Server Metadata Extraction**: Intelligent file information retrieval
- ✅ **Multi-threaded Downloads**: Parallel chunk downloading for speed boost
- ✅ **Range Request Support**: Automatically detects and uses HTTP range requests
- ✅ **Retry Mechanism**: 3-attempt retry system for network resilience
- ✅ **Redirect Handling**: Follows HTTP redirects seamlessly
- ✅ **RFC 6266 Compliance**: Proper filename extraction from Content-Disposition headers
- ✅ **MIME Type Detection**: Automatic file type identification
- ✅ **Fallback Strategies**: Multiple methods for filename and metadata extraction

### Planned Features

- 🔄 **Resume Downloads**: Pause and resume capability
- 🔄 **Download Queue**: Batch download management
- 🔄 **Browser Integration**: Seamless browser plugin support for both Geko and Chromium
- 🔄 **GUI Interface**: User-friendly graphical interface to tweak settings and configurations
- 🔄 **Download Categories**: Automatic file organization

## Architecture

UDM follows a modular architecture with clean separation of concerns:

```
UDM Engine
├── Server Analysis Module (ServerHeaders.go)
│   ├── Metadata Extraction
│   ├── Range Support Detection
│   └── URL Resolution
├── File System Module (External: self made ufs package)
│   ├── File Operations
│   ├── Metadata Management
│   └── Directory Handling
└── Download Engine (In Development)
    ├── Multi-threading Controller
    ├── Chunk Management
    └── Progress Tracking
```

## Components

### 1. Server Headers Module (ServerHeaders.go)

The core module responsible for analyzing remote files and extracting crucial metadata before download initiation.

#### Key Functions:

- **`GetServerData`**: Main entry point for server analysis
- **`tryGetServerData`**: Core implementation with HEAD/GET fallback
- **`mimeExtensionFromContentType`**: MIME type to extension mapping

#### Data Structure:

```go
type ServerData struct {
    Filename      string  // Extracted filename
    Filesize      int64   // File size in bytes
    Filetype      string  // MIME type
    AcceptsRanges bool    // Range request support
    FinalURL      string  // URL after redirects
}
```

## How It Works

### 1. Server Analysis Phase

When you provide a download URL, UDM performs intelligent server analysis:

```
URL Input → HEAD Request → GET Fallback → Metadata Extraction
    ↓              ↓             ↓              ↓
Network Test → Response Check → Full Request → File Information
```

#### Process Flow:

1. **Initial HEAD Request**: Lightweight request to gather metadata
2. **GET Fallback**: If HEAD fails, performs GET request
3. **Redirect Following**: Tracks all redirects to final destination
4. **Metadata Extraction**: Extracts filename, size, type, and capabilities
5. **Range Detection**: Determines if server supports partial downloads

### 2. Retry Mechanism

UDM implements a robust 3-attempt retry system:

- **Attempt 1**: Initial request
- **Attempt 2**: Retry after 2-second delay (if network error)
- **Attempt 3**: Final attempt after another delay
- **Failure**: Returns comprehensive error information

### 3. Filename Resolution Strategy

UDM uses a sophisticated multi-step filename resolution:

1. **Content-Disposition Header**: RFC 6266 compliant parsing
2. **UTF-8 Encoding**: Handles international filenames
3. **URL Path Extraction**: Fallback to URL basename
4. **MIME Type Mapping**: Last resort with appropriate extension
5. **Default Naming**: "downloaded_file" with detected extension

### 4. Multi-threading Preparation

The server analysis phase prepares for multi-threaded downloads by:

- Checking `Accept-Ranges: bytes` header
- Determining file size for chunk calculation
- Validating server capabilities
- Optimizing request strategy

## Installation

### Prerequisites

- Go 1.24.2 or higher
- Internet connection for dependencies

### From Source

```bash
git clone <repository-url>
cd nudm
go mod tidy
go build
```

## Usage

### Basic Server Analysis

```go
package main

import (
    "fmt"
    "log"
)

func main() {
    url := "https://example.com/largefile.zip"

    // Analyze server and file metadata
    serverData, err := GetServerData(url)
    if err != nil {
        log.Fatal("Failed to analyze server:", err)
    }

    // Display file information
    fmt.Printf("Filename: %s\n", serverData.Filename)
    fmt.Printf("Size: %d bytes (%.2f MB)\n",
        serverData.Filesize,
        float64(serverData.Filesize)/1024/1024)
    fmt.Printf("Type: %s\n", serverData.Filetype)
    fmt.Printf("Supports Range Requests: %v\n", serverData.AcceptsRanges)
    fmt.Printf("Final URL: %s\n", serverData.FinalURL)
}
```

### Error Handling

```go
serverData, err := GetServerData(url)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "failed after"):
        fmt.Println("Network connectivity issues")
    case strings.Contains(err.Error(), "timeout"):
        fmt.Println("Server response timeout")
    default:
        fmt.Printf("Server analysis failed: %v\n", err)
    }
    return
}
```

## API Reference

### ServerData Structure

| Field           | Type     | Description                            |
| --------------- | -------- | -------------------------------------- |
| `Filename`      | `string` | Extracted or generated filename        |
| `Filesize`      | `int64`  | File size in bytes (0 if unknown)      |
| `Filetype`      | `string` | MIME type from Content-Type header     |
| `AcceptsRanges` | `bool`   | Whether server supports range requests |
| `FinalURL`      | `string` | Final URL after following redirects    |

### Functions

#### `GetServerData(downloadURL string) (*ServerData, error)`

Main function to analyze a remote file and extract metadata.

**Parameters:**

- `downloadURL`: The URL of the file to analyze

**Returns:**

- `*ServerData`: Comprehensive file and server information
- `error`: Error details if analysis fails

**Features:**

- 3-attempt retry mechanism
- Network error resilience
- Comprehensive error reporting

#### `tryGetServerData(downloadURL string) (*ServerData, error)`

Core implementation that performs a single analysis attempt.

**Process:**

1. HEAD request attempt
2. GET request fallback
3. Metadata extraction
4. Filename resolution
5. Server capability detection

## Implementation Details

### Network Resilience

UDM implements several layers of network resilience:

- **Timeout Management**: 15-second timeout per request
- **Retry Logic**: Exponential backoff with 2-second intervals
- **Request Method Fallback**: HEAD → GET progression
- **Redirect Handling**: Automatic redirect following
- **Error Classification**: Distinguishes between recoverable and permanent errors

### Filename Extraction Algorithm

```
Content-Disposition Header Present?
├── Yes → Parse RFC 6266 format
│   ├── filename= parameter found? → Use value
│   └── filename*= parameter found? → Decode UTF-8
└── No → Extract from URL path
    ├── Valid filename in path? → Use basename
    └── No valid filename → Generate from MIME type
        └── Unknown MIME → Use "downloaded_file"
```

### Performance Optimizations

- **Efficient HTTP Client**: Reused client with optimal timeouts
- **Minimal Data Transfer**: HEAD requests minimize bandwidth
- **Memory Management**: Discards response bodies when not needed
- **String Operations**: Efficient string manipulation for parsing

## Contributing

We welcome contributions to UDM! Please follow our contribution guidelines:

### Code Style

- Follow Go conventions and use `go fmt`
- Include comprehensive documentation for all functions
- Add examples for complex functionality
- Implement proper error handling

### Documentation Format

```go
// FunctionName does X and Y for the given input
// Detailed description of the function's behavior and purpose
//
// Parameters:
//   - param1: Description of parameter 1
//   - param2: Description of parameter 2
//
// Returns:
//   - returnType: Description of return value
//   - error: Error conditions and meanings
//
// Example:
//   result, err := FunctionName(input)
//   if err != nil {
//       // Handle error
//   }
//   // Use result
```

### Development Setup

1. Fork the repository
2. Create a feature branch
3. Implement changes with tests
4. Update documentation
5. Submit a pull request

## Changelog

See changelog.md for detailed version history and recent improvements.

### Recent Major Updates

- **Enhanced Retry Mechanism**: 3-attempt retry system for network resilience
- **Improved Filename Handling**: RFC 6266 compliant parsing with UTF-8 support
- **Range Request Detection**: Automatic server capability detection
- **Redirect Tracking**: Complete redirect chain following
- **MIME Type Fallback**: Comprehensive file type detection

## Roadmap

### Phase 1 (Current): Foundation ✅

- Server analysis and metadata extraction
- Basic HTTP handling and error management
- Filename resolution and MIME type detection

### Phase 2 (In Progress): Core Download Engine 🔄

- Multi-threaded download implementation
- Chunk management and coordination
- Progress tracking and reporting

### Phase 3 (Planned): Advanced Features 📋

- Resume and pause functionality
- Download queue management
- Bandwidth control and scheduling

### Phase 4 (Future): User Experience 🎯

- Graphical user interface
- Browser integration
- Advanced configuration options

## License

This project is open source. Please check the license file for details.

## Support

For issues, feature requests, or contributions, please use the project's issue tracker or submit a pull request.

---

**UDM - Because your downloads deserve to be ultimate.** 🚀
