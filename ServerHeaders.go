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

type ServerData struct {
	Filename      string
	Filesize      int64
	Filetype      string
	AcceptsRanges bool
	FinalURL      string
}

func getServerData(downloadURL string) (*ServerData, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// We'll let redirect happen and record the last URL later
			return nil
		},
	}

	// Try HEAD first (cheap)
	req, err := http.NewRequest("HEAD", downloadURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode >= 400 {

		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Status Code: %d\n", resp.StatusCode)
		fmt.Printf("Trying GET...\n")

		// fallback to GET
		reqGet, err := http.NewRequest("GET", downloadURL, nil)
		if err != nil {
			return nil, err
		}
		resp, err = client.Do(reqGet)
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()

	finalURL := resp.Request.URL.String() // after redirect

	// Initialize struct
	data := &ServerData{
		FinalURL: finalURL,
	}

	// 1. Filename from Content-Disposition
	cd := resp.Header.Get("Content-Disposition")
	if cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			if name, ok := params["filename"]; ok {
				data.Filename = name
			} else if name, ok := params["filename*"]; ok {
				// decode UTF-8 format
				if strings.HasPrefix(name, "UTF-8''") {
					decoded, err := url.QueryUnescape(strings.TrimPrefix(name, "UTF-8''"))
					if err == nil {
						data.Filename = decoded
					}
				}
			}
		}
	}

	// 2. Fallback to URL path
	if data.Filename == "" {
		if parsed, err := url.Parse(finalURL); err == nil {
			base := path.Base(parsed.Path)
			if base != "" && strings.Contains(base, ".") {
				data.Filename = base
			}
		}
	}

	// 3. Filesize
	cl := resp.Header.Get("Content-Length")
	if cl != "" {
		var size int64
		fmt.Sscanf(cl, "%d", &size)
		data.Filesize = size
	}

	// 4. Filetype
	ct := resp.Header.Get("Content-Type")
	if ct != "" {
		data.Filetype = ct
	}

	// 5. Accept-Ranges
	if strings.Contains(resp.Header.Get("Accept-Ranges"), "bytes") {
		data.AcceptsRanges = true
	}

	// 6. Final fallback filename (if nothing worked)
	if data.Filename == "" {
		ext := mimeExtensionFromContentType(data.Filetype)
		data.Filename = "downloaded_file" + ext
	}

	// Optional: Discard body if GET was used (to avoid partial download impact)
	if resp.Request.Method == "GET" {
		io.Copy(io.Discard, resp.Body)
	}

	return data, nil
}

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

/*
func main() {
	url := "https://drive.usercontent.google.com/download?id=1d1EBTcLHYQiv93O4nyBBjbK_Wc-2f5qX&export=download&authuser=0&confirm=t&uuid=5cccf6aa-fc97-4bff-89f3-e4339a189778&at=AN8xHoo83pMm2eQ2GwbC6YHA5eK0:1753245767095"
	info, err := getServerData(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Filename: %s\n", info.Filename)
	fmt.Printf("Size: %d bytes\n", info.Filesize)
	fmt.Printf("Filetype: %s\n", info.Filetype)
	fmt.Printf("Accepts Range Requests: %v\n", info.AcceptsRanges)
	fmt.Printf("Final URL after redirect: %s\n", info.FinalURL)
}

*/
