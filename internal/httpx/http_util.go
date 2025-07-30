package httpx

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

var HttpClient = &http.Client{
	Timeout: time.Second * 5,
	Transport: &http.Transport{
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		ResponseHeaderTimeout: time.Second * 10,
		DisableCompression:    false,
		Proxy:                 http.ProxyFromEnvironment,
	},
}

var UA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36"

// LoadUrlContent fetches content from the specified URL and returns it as a byte slice.
// It handles HTTP request creation, execution, and response body reading.
// Returns an error if the request fails or the status code is not OK (200).
func LoadUrlContent(url string) (content []byte, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UA)
	req.Header.Set("Accept", "*/*")

	resp, err := HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed, status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// LoadUrlContentWithRetry attempts to fetch content from the specified URL with multiple retries.
// It uses LoadUrlContent internally and retries the request up to retryTimes if errors occur.
// Returns the content if any attempt succeeds, otherwise returns the last error encountered.
func LoadUrlContentWithRetry(url string, retryTimes int64) (content []byte, err error) {
	for i := int64(0); i < retryTimes; i++ {
		content, err = LoadUrlContent(url)
		if err == nil {
			return content, nil
		}
	}
	return nil, err
}

// PingURL measures the latency (response time) of the specified URL using an HTTP HEAD request.
// It calculates the time taken from sending the request to receiving the response headers.
// Returns the latency in milliseconds and any error that occurred during the request.
func PingURL(url string) (latency int64, err error) {
	start := time.Now()
	resp, err := HttpClient.Head(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("request failed, status code: %d", resp.StatusCode)
	}
	latency = int64(time.Since(start) / time.Millisecond)
	return latency, nil
}

// TestDownloadSpeed measures the download speed of a given URL by downloading a portion or the entire file.
// It returns the download speed in kilobytes per second (KB/s) and any error that occurred during the test.
// For files larger than 5MB, it downloads the first 5MB to calculate the speed.
// For smaller files, it downloads the entire file.
func TestDownloadSpeed(url string) (kbps float64, err error) {
	// Determine the size of data to download
	testSize := int64(10 * (1 << 20)) // 10MB

	// Send GET request to start downloading
	getReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	getReq.Header.Set("User-Agent", UA)
	getReq.Header.Set("Accept", "*/*")
	getReq.Header.Set("Cache-Control", "no-cache")

	getResp, err := HttpClient.Do(getReq)
	if err != nil {
		return 0, err
	}
	defer getResp.Body.Close()

	// Validate response status code
	if getResp.StatusCode != http.StatusOK && getResp.StatusCode != http.StatusPartialContent {
		return 0, fmt.Errorf("request failed, status code: %d", getResp.StatusCode)
	}

	// Create temporary buffer
	buffer := make([]byte, 32*1024) // 32KB buffer
	downloaded := int64(0)

	// Start downloading and timing
	start := time.Now()
	for {
		// Check if the test size has been reached
		if downloaded >= testSize {
			break
		}

		// Calculate the remaining bytes to read
		bytesToRead := testSize - downloaded
		if bytesToRead > int64(len(buffer)) {
			bytesToRead = int64(len(buffer))
		}

		// Read data
		n, err := getResp.Body.Read(buffer[:bytesToRead])

		if err != nil {
			if err == io.EOF {
				break // Normal end
			}
			return 0, err
		}
		downloaded += int64(n)
	}

	// Calculate download speed
	elapsedTime := time.Since(start)
	if elapsedTime > 0 {
		seconds := elapsedTime.Seconds()
		kbps = float64(downloaded) / seconds / 1024
	}
	return kbps, nil
}
