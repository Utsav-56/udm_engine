package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

/*
  File contains:
  The code to get the metadata of a file hosted in a server
*/

// ServerData struct to store the server data
// Stores the data of the metadata of the file hosted in a server
//
// Parameters:
//   - Filename: The name of the file
//   - Filesize: The size of the file in bytes
//   - Filetype: The type of the file
//   - AcceptsRanges: Boolean indicating if the server accepts range requests
//   - FinalURL: The final URL of the file after following redirects
type ServerData struct {
	Filename      string
	Filesize      int64
	Filetype      string
	AcceptsRanges bool
	FinalURL      string
}

/*
	To the contributor::


	Thanks for your contribution to this project!
	and please geenrate documentation of your changes and codes properly
	a format of a documentation is as follows:
	1. A description of the function
	2. A description of the parameters
	3. A description of the return value
	4. An example of how to use the function if possible also show how to handle error,

// FOllowing is a example for it but it has nothing to do with our code just for reference it is there

// GetFileSize returns the size of the given file in bytes.
// This function checks if the path exists and is a file before retrieving its size.
//
// Parameters:
//   - path: The absolute or relative path to the file
//
// Returns:
//   - int64: The size of the file in bytes
//   - Returns 0 if the file doesn't exist, is a directory, or if an error occurs
//
// Example:
//
//	size := ufs.GetFileSize("/path/to/file.txt")
//	fmt.Printf("File size: %d bytes\n", size)



*/

// GetServerData returns the filename, filesize, file type, accepts range requests, and final URL of the server
// It also handles errors and returns the error message
//
// Working:
//   - The function takes a downloadURL as input
//   - The function makes a HEAD request at first to the provided downloadURL
//   - If the request fails, it makes a GET request to the provided downloadURL
//   - If the request is successful, it returns the filename, filesize, file type, accepts range requests, and final URL of the server
//   - If the request fails, it returns an error message
//
// Note:
//   - The function handles errors and returns the error message
//   - The function also includes a retry mechanism which will retry up to 3 times
//
// Parameters:
//   - downloadURL: The URL of the file to download
//
// Returns:
//   - *ServerData: A struct containing the filename, filesize, file type, accepts range requests, and final URL of the server
//   - error: An error message if the function fails
//
// Example:
//
//	func main(){
//		url := "https://example.com/sample.pdf"
//		info, err := getServerData(url)
//
//		if err != nil {
//			fmt.Println("Error:", err)
//			return
//		}
//
//		fmt.Printf("Filename: %s\n", info.Filename)
//		fmt.Printf("Size: %d bytes\n", info.Filesize)
//		fmt.Printf("Filetype: %s\n", info.Filetype)
//		fmt.Printf("Accepts Range Requests: %v\n", info.AcceptsRanges)
//		fmt.Printf("Final URL after redirect: %s\n", info.FinalURL)
//	}
func GetServerData(downloadURL string) (*ServerData, error) {
	const maxRetries = 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		data, err := tryGetServerData(downloadURL)
		if err == nil {
			return data, nil
		}
		lastErr = err
		if attempt < maxRetries {
			time.Sleep(2 * time.Second) // short wait before retry
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %v", maxRetries, lastErr)
}

// tryGetServerData attempts to retrieve server data using a HEAD request, falling back to a GET request if necessary
//
// Working:
//   - The function takes a downloadURL as input
//   - The function makes a HEAD request to the provided downloadURL
//   - If the HEAD request fails, it makes a GET request to the provided downloadURL
//   - If the request is successful, it returns the server data
//   - If the request fails, it returns an error message
//
// Parameters:
//   - downloadURL: The URL of the file to download
//
// Returns:
//   - *ServerData: A struct containing the filename, filesize, file type, accepts range requests, and final URL of the server
//   - error: An error message if the function fails
//
// Example:
//
//	func main(){
//		url := "https://example.com/sample.pdf"
//		data, err := tryGetServerData(url)
//
//		if err != nil {
//			fmt.Println("Error:", err)
//			return
//		}
//
//		fmt.Printf("Filename: %s\n", data.Filename)
//		fmt.Printf("Size: %d bytes\n", data.Filesize)
//		fmt.Printf("Filetype: %s\n", data.Filetype)
//		fmt.Printf("Accepts Range Requests: %v\n", data.AcceptsRanges)
//		fmt.Printf("Final URL after redirect: %s\n", data.FinalURL)
//	}
func tryGetServerData(downloadURL string) (*ServerData, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	// 1. Try HEAD request
	req, err := http.NewRequest("HEAD", downloadURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode >= 400 {
		// 2. Fallback to GET request
		// reqGet, err := http.NewRequest("GET", downloadURL, nil)
		// if err != nil {
		// 	return nil, err
		// }
		// resp, err = client.Do(reqGet)
		// if err != nil {
		// 	return nil, err
		// }
		return nil, err
	}
	defer resp.Body.Close()

	finalURL := resp.Request.URL.String()

	data := &ServerData{
		FinalURL: finalURL,
	}

	// 3. Content-Disposition based filename
	cd := resp.Header.Get("Content-Disposition")
	if cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			if name, ok := params["filename"]; ok {
				data.Filename = name
			} else if name, ok := params["filename*"]; ok {
				if strings.HasPrefix(name, "UTF-8''") {
					decoded, err := url.QueryUnescape(strings.TrimPrefix(name, "UTF-8''"))
					if err == nil {
						data.Filename = decoded
					}
				}
			}
		}
	}

	// 4. Fallback to path in URL
	if data.Filename == "" {
		if parsed, err := url.Parse(finalURL); err == nil {
			base := path.Base(parsed.Path)
			if base != "" && strings.Contains(base, ".") {
				data.Filename = base
			}
		}
	}

	// 5. Content-Length
	cl := resp.Header.Get("Content-Length")
	if cl != "" {
		var size int64
		fmt.Sscanf(cl, "%d", &size)
		data.Filesize = size
	}

	// 6. Content-Type
	data.Filetype = resp.Header.Get("Content-Type")

	// 7. Accept-Ranges
	if strings.Contains(resp.Header.Get("Accept-Ranges"), "bytes") {
		data.AcceptsRanges = true
	}

	// 8. Last fallback for filename
	if data.Filename == "" {
		ext := mimeExtensionFromContentType(data.Filetype)
		data.Filename = "downloaded_file" + ext
	}

	// If GET was used, discard the partial body
	if resp.Request.Method == "GET" {
		io.Copy(io.Discard, resp.Body)
	}

	return data, nil
}

// mimeExtensionFromContentType extracts the file extension from a Content-Type header
//
// Working:
//   - The function takes a Content-Type header value as input
//   - The function checks if the Content-Type header contains a known file extension
//   - If a match is found, it returns the file extension
//   - If no match is found, it returns an empty string
//
// Parameters:
//   - ct: The Content-Type header value
//
// Returns:
//   - string: The file extension, or an empty string if not found
//
// Example:
//
//	extension := mimeExtensionFromContentType("text/html")
//	fmt.Printf("File extension: %s\n", extension)
func mimeExtensionFromContentType(ct string) string {
	// Add more if needed
	mapping := map[string]string{
		"image/jpeg":      ".jpg",
		"image/png":       ".png",
		"image/gif":       ".gif",
		"text/html":       ".html",
		"application/pdf": ".pdf",
	}
	if ext, ok := mapping[ct]; ok {
		return ext
	}
	return ""
}

func extractFilename(resp *http.Response) string {
	cd := resp.Header.Get("Content-Disposition")
	if cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			if name, ok := params["filename"]; ok {
				return name
			} else if name, ok := params["filename*"]; ok {
				if strings.HasPrefix(name, "UTF-8''") {
					decoded, err := url.QueryUnescape(strings.TrimPrefix(name, "UTF-8''"))
					if err == nil {
						return decoded
					}
				}
			}
		}
	}

	parsed, err := url.Parse(resp.Request.URL.String())
	if err == nil {
		base := path.Base(parsed.Path)
		if base != "" && strings.Contains(base, ".") {
			return base
		}
	}
	return ""
}
