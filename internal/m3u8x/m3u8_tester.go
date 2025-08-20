package m3u8x

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/rambollwong/rainbow-iptv-source-filter/internal/httpx"
	"github.com/rambollwong/rainbow-iptv-source-filter/pkg/proto"
	"github.com/rambollwong/rainbowcat/pool"
	"github.com/rambollwong/rainbowlog/log"
)

// ParallelTestProgramListSource filters the given ProgramListSource by testing the latency of XTvgUrls
// and the download speed of channel streams in parallel using a worker pool.
// It returns a new ProgramListSource containing only the URLs and channels that pass the tests.
func ParallelTestProgramListSource(
	ctx context.Context,
	source *ProgramListSource,
	minLatency, loadMinSpeed, retryTimes int64,
	workerPool *pool.WorkerPool,
	groupList []*proto.GroupList,
	hostCustomUA map[string]string,
) (filteredSource *ProgramListSource) {
	// Initialize the filtered source and synchronization primitives
	filteredSource = NewProgramListSource()
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}

	// Test each XTvgUrl for latency
	for _, tvgUrl := range source.XTvgUrls {
		if tvgUrl == "" {
			continue
		}

		wg.Add(1)
		testFunc := func() {
			defer wg.Done()
			// Ping the URL to measure latency
			latency, err := httpx.PingURL(tvgUrl)
			if err != nil {
				log.Error().Msg("Failed to ping tvg url, ignore.").Err(err).Done()
				return
			}
			// Check if latency exceeds the minimum allowed
			if latency > minLatency {
				log.Info().Msg("Tvg url latency is too long, ignore.").
					Str("tvg_url", tvgUrl).
					Int64("latency", latency).
					Done()
				return
			}
			// Log successful latency test and add URL to filtered source
			log.Info().Msg("Tvg url latency is ok.").
				Str("tvg_url", tvgUrl).
				Int64("latency", latency).
				Done()
			mu.Lock()
			filteredSource.XTvgUrls = append(filteredSource.XTvgUrls, tvgUrl)
			mu.Unlock()
		}
		// Submit the test function to the worker pool
		if err := workerPool.Submit(testFunc); err != nil {
			log.Fatal().Err(err).Msg("Failed to submit test func").Done()
		}
	}

	// Test each channel in the program list
	hostGroupChannels := make(map[string]map[string][]*Channel) // host -> tvgName -> channels
	for _, list := range groupList {
		for _, tvgName := range list.TvgName {
			// Initialize the channel slice in the filtered source if it doesn't exist
			if _, exists := filteredSource.TvgNameChannels[tvgName]; !exists {
				filteredSource.TvgNameChannels[tvgName] = make([]*Channel, 0, 8)
			}
			chs, exist := source.TvgNameChannels[tvgName]
			if !exist {
				continue
			}
			log.Info().Int("number_of_channels_waiting_for_testing", len(chs)).Str("tvg_name", tvgName).Done()
			// Iterate over each channel for the current tvgName
			for _, ch := range chs {
				if strings.Contains(ch.Url, "audio") {
					continue
				}

				// Group all channels by host first, then test each host sequentially
				// This approach prevents test failures due to server request rate limiting
				u, err := url.Parse(ch.Url)
				if err != nil {
					log.Error().Msg("Failed to parse channel url, ignore.").
						Str("tvg_name", tvgName).
						Str("channel_url", ch.Url).
						Done()
					return
				}
				host := u.Host
				hostChannels, exist := hostGroupChannels[host]
				if !exist {
					hostChannels = make(map[string][]*Channel)
				}
				hostTvgChs, exist := hostChannels[ch.TvgName]
				if !exist {
					hostTvgChs = make([]*Channel, 0, 16)
				}
				hostTvgChs = append(hostTvgChs, ch)
				hostChannels[ch.TvgName] = hostTvgChs
				hostGroupChannels[host] = hostChannels
			}
		}
	}

	for _, list := range groupList {
		for _, tvgName := range list.TvgName {
			for host, tvgChs := range hostGroupChannels {
				customUA := hostCustomUA[host]
				chs, exist := tvgChs[tvgName]
				if !exist {
					continue
				}

				wg.Add(1)
				testFunc := func() {
					defer wg.Done()

					for _, ch := range chs {
						// Log the start of channel URL testing
						log.Info().Msg("Testing channel url...").
							Str("tvg_name", tvgName).
							Str("channel_url", ch.Url).
							Str("host", host).
							Done()

						u, err := url.Parse(ch.Url)
						if err != nil {
							log.Error().Msg("Failed to parse channel url, ignore.").
								Str("tvg_name", tvgName).
								Str("channel_url", ch.Url).
								Done()
							return
						}
						if strings.HasSuffix(u.Path, ".m3u8") {
							if !TestM3u8DownloadSpeedWithRetry(
								ctx, ch.Url, customUA, float64(loadMinSpeed), retryTimes) {
								return
							}
						} else {
							speed, err := httpx.TestDownloadSpeed(ctx, ch.Url)
							if err != nil {
								if errors.Is(err, context.Canceled) {
									return
								}
								log.Error().Msg("Failed to test channel url load speed, ignore.").
									Str("tvg_name", tvgName).
									Str("channel_url", ch.Url).
									Done()
								return
							}
							if speed < float64(loadMinSpeed) {
								log.Warn().Msg("Channel url load speed is too low, ignore.").
									Str("tvg_name", tvgName).
									Str("channel_url", ch.Url).
									Float64("speed", speed).
									Done()
								return
							}
						}

						// Add the channel to the filtered source and break to avoid duplicates
						mu.Lock()
						filteredSource.TvgNameChannels[tvgName] = append(filteredSource.TvgNameChannels[tvgName], ch)
						mu.Unlock()
						log.Info().Msg("Channel is ok.").
							Str("tvg_name", tvgName).
							Str("channel_url", ch.Url).
							Done()
					}
				}

				// Submit the channel test function to the worker pool
				if err := workerPool.Submit(testFunc); err != nil {
					if !errors.Is(err, pool.ErrWorkerPoolClosed) && !errors.Is(err, pool.ErrWorkerPoolClosing) {
						log.Warn().Msg("Failed to submit test task").
							Err(err).
							Done()
					}
				}
			}

			// Wait for all tests to complete
			wg.Wait()
		}
	}

	return filteredSource
}

// TestM3u8DownloadSpeed tests the download speed of media data corresponding to an m3u8 URL.
// Input: Network URL of the m3u8 file and the required minimum download speed (kb/s).
// Output: Returns true if any ts segment meets the speed requirement, otherwise returns false; along with possible error.
func TestM3u8DownloadSpeed(ctx context.Context, m3u8URL, customUA string, requiredSpeed float64) bool {
	// Download and parse the m3u8 file to get .ts segment URLs (first and last one)
	tsURLs, err := getFirstAndLastTsSegmentURL(ctx, m3u8URL, customUA)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return false
		}
		log.Error().Msg("Failed to download m3u8 file, ignore.").
			Str("m3u8_url", m3u8URL).Err(err).
			Done()
		return false
	}

	// Test the download speed of .ts segments (limit max download to 10MB per segment to avoid resource waste)
	const maxTestSize = 10 * 1024 * 1024 // 10MB
	var totalSpeed float64
	for _, tsURL := range tsURLs {
		speed, err := testFileDownloadSpeed(ctx, tsURL, customUA, maxTestSize)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return false
			}
			log.Error().Msg("Failed to test file download speed, skip this file.").
				Str("m3u8_url", m3u8URL).
				Str("file_url", tsURL).Err(err).
				Done()
			continue
		}
		totalSpeed += speed
		// If any segment meets the speed requirement, return success immediately
		if speed >= requiredSpeed {
			break
		}
	}

	// If multiple segments were tested, calculate the average speed for judgment
	if len(tsURLs) > 1 {
		totalSpeed = totalSpeed / float64(len(tsURLs))
	}
	// Return success if average speed meets the requirement
	if totalSpeed >= requiredSpeed {
		log.Info().Msg("File download speed is ok.").
			Float64("kbps", totalSpeed).
			Str("m3u8_url", m3u8URL).
			Done()
		return true
	}
	// None of the segments meet the speed requirement
	log.Warn().Msg("M3u8 url load speed is too low, ignore.").
		Str("m3u8_url", m3u8URL).
		Float64("kbps", totalSpeed).
		Done()
	return false
}

// TestM3u8DownloadSpeedWithRetry tests the download speed of an m3u8 URL with retry logic.
// It attempts to test the download speed up to retryTimes+1 times (1 initial attempt + retryTimes retries).
// Returns true if the test passes within the required speed at least once, otherwise returns false.
func TestM3u8DownloadSpeedWithRetry(
	ctx context.Context,
	m3u8URL, customUA string,
	requiredSpeed float64,
	retryTimes int64,
) bool {
	var i int64
	for {
		i++
		if TestM3u8DownloadSpeed(ctx, m3u8URL, customUA, requiredSpeed) {
			return true
		}
		if i > retryTimes {
			return false
		}
		log.Debug().Msg("Failed to test m3u8 download speed, retrying...").
			Str("m3u8_url", m3u8URL).
			Int64("retry_times", i).
			Done()
	}
}

// getFirstAndLastTsSegmentURL extracts the first and last valid .ts segment URLs from an m3u8 file.
func getFirstAndLastTsSegmentURL(ctx context.Context, m3u8URL, customUA string) ([]string, error) {
	// Download m3u8 file content
	req, err := http.NewRequestWithContext(ctx, "GET", m3u8URL, nil)
	if err != nil {
		return nil, err
	}
	if customUA == "" {
		req.Header.Set("User-Agent", httpx.UA)
	} else {
		req.Header.Set("User-Agent", customUA)
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := httpx.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed, status code: %d", resp.StatusCode)
	}

	m3u8Content, err := httpx.LoadUrlContent(ctx, m3u8URL)
	if err != nil {
		return nil, fmt.Errorf("failed to load content: %v", err.Error())
	}

	// Parse m3u8 content to extract .ts segment URLs
	tsURLs, err := parseTsSegments(string(m3u8Content), m3u8URL)
	if err != nil {
		return nil, err
	}

	l := len(tsURLs)
	if l == 0 {
		return nil, fmt.Errorf("ts segment not found")
	}
	if l == 1 {
		return []string{tsURLs[0]}, nil
	}

	return []string{tsURLs[0], tsURLs[l-1]}, nil
}

// parseTsSegments parses all .ts segment absolute URLs from m3u8 content.
func parseTsSegments(m3u8Content, baseURL string) ([]string, error) {
	lines := strings.Split(m3u8Content, "\n")
	var tsURLs []string

	// Parse base URL for relative path concatenation
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url failed")
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip m3u8 comment lines (starting with #)
		if strings.HasPrefix(line, "#") {
			continue
		}
		// Handle relative path .ts files
		tsURL, err := parsedBaseURL.Parse(line)
		if err != nil {
			continue // Skip invalid URLs
		}
		tsURLs = append(tsURLs, tsURL.String())
	}

	return tsURLs, nil
}

// testFileDownloadSpeed tests the download speed of a specified URL and returns kb/s.
func testFileDownloadSpeed(ctx context.Context, fileURL, customUA string, maxDownloadSize int64) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fileURL, nil)
	if err != nil {
		return 0, err
	}
	if customUA == "" {
		req.Header.Set("User-Agent", httpx.UA)
	} else {
		req.Header.Set("User-Agent", customUA)
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := httpx.HttpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to load ts file")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("request failed, status code: %d", resp.StatusCode)
	}

	// Start timing and download data
	startTime := time.Now()
	buffer := make([]byte, 32*1024) // 32KB buffer
	var downloadedBytes int64

	for {
		// Check if maximum download size is reached
		if downloadedBytes >= maxDownloadSize {
			break
		}

		// Calculate remaining bytes to download
		remaining := maxDownloadSize - downloadedBytes
		if remaining < int64(len(buffer)) {
			buffer = buffer[:remaining]
		}

		// Read data
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			downloadedBytes += int64(n)
		}

		// Handle read errors
		if err != nil {
			if err == io.EOF {
				break // Normal end (file smaller than max test size)
			}
			return 0, err
		}
	}

	// Calculate download speed (kb/s = (bytes / 1024) / seconds)
	elapsedSeconds := time.Since(startTime).Seconds()
	if elapsedSeconds <= 0 {
		return 0, fmt.Errorf("wrong elapsed seconds")
	}

	speedKbPerSec := float64(downloadedBytes) / elapsedSeconds / 1024
	if speedKbPerSec <= 0 {
		log.Debug().Msg("Wrong speed").Int64("downloaded_bytes", downloadedBytes).Float64("elapsed_seconds", elapsedSeconds).Done()
	}
	return speedKbPerSec, nil
}
