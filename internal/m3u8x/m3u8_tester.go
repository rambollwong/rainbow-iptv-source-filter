package m3u8x

import (
	"sync"

	"github.com/rambollwong/rainbow-iptv-source-filter/internal/httpx"
	"github.com/rambollwong/rainbowcat/pool"
	"github.com/rambollwong/rainbowlog/log"
)

// ParallelTestProgramListSource filters the given ProgramListSource by testing the latency of XTvgUrls
// and the download speed of channel streams in parallel using a worker pool.
// It returns a new ProgramListSource containing only the URLs and channels that pass the tests.
func ParallelTestProgramListSource(
	source *ProgramListSource,
	minLatency, loadMinSpeed, retryTimes int64,
	workerPool *pool.WorkerPool) (filteredSource *ProgramListSource) {
	// Initialize the filtered source and synchronization primitives
	filteredSource = NewProgramListSource()
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}

	// Test each XTvgUrl for latency
	for _, tvgUrl := range source.XTvgUrls {
		wg.Add(1)
		testFunc := func() {
			defer wg.Done()
			// Ping the URL to measure latency
			latency, err := httpx.PingURL(tvgUrl)
			if err != nil {
				log.Error().Err(err).Msg("Failed to ping tvg url, ignore.").Done()
				return
			}
			// Check if latency exceeds the minimum allowed
			if latency > minLatency {
				log.Info().Msg("tvg url latency is too long, ignore.").
					Str("tvg_url", tvgUrl).
					Int64("latency", latency).
					Done()
				return
			}
			// Log successful latency test and add URL to filtered source
			log.Info().Msg("tvg url latency is ok.").
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
	for tvgName, chs := range source.TvgNameChannels {
		// Initialize the channel slice in the filtered source if it doesn't exist
		if _, exists := filteredSource.TvgNameChannels[tvgName]; !exists {
			filteredSource.TvgNameChannels[tvgName] = make([]*Channel, 0, 8)
		}

		// Iterate over each channel for the current tvgName
		for idx, ch := range chs {
			wg.Add(1)
			testFunc := func() {
				defer wg.Done()
				// Log the start of channel URL testing
				log.Info().Msg("Testing channel url...").
					Str("tvg_name", tvgName).
					Int("index", idx).
					Str("channel_url", ch.Url).
					Done()

				// Ping the channel URL to check latency
				latency, err := httpx.PingURL(ch.Url)
				if err != nil {
					log.Warn().Err(err).Msg("Failed to ping channel url, skip.").
						Str("channel_url", ch.Url).
						Done()
				} else {
					// If latency is too high, skip this channel
					if latency > minLatency {
						log.Info().Msg("Channel url latency is too long, ignore.").
							Str("channel_url", ch.Url).
							Int64("latency", latency).
							Done()
						return
					}
					// Log successful latency test
					log.Debug().Msg("Channel url latency is ok.").
						Str("channel_url", ch.Url).
						Int64("latency", latency).
						Done()
				}

				// Load the content of the channel URL with retries
				urlContent, err := httpx.LoadUrlContentWithRetry(ch.Url, retryTimes)
				if err != nil {
					log.Error().Err(err).Msg("Failed to load channel url, ignore.").
						Str("channel_url", ch.Url).
						Done()
					return
				}

				// Parse the live stream source from the loaded content
				liveStreamSource := NewLiveStreamSource(ch.Url)
				if err := liveStreamSource.ParseLiveStreamSource(urlContent); err != nil {
					log.Error().Err(err).Msg("Failed to parse live stream source for channel url, ignore.").
						Str("channel_url", ch.Url).
						Done()
					return
				}
				// Check if any files were found in the live stream source
				filesCount := len(liveStreamSource.Files)
				if filesCount == 0 {
					log.Info().Msg("No files found in live stream source, ignore.").
						Str("channel_url", ch.Url).
						Done()
					return
				}

				// Only test the first and last file for download speed
				var files []LiveStreamFile
				if filesCount == 1 {
					files = liveStreamSource.Files
				} else {
					files = []LiveStreamFile{liveStreamSource.Files[0], liveStreamSource.Files[filesCount-1]}
				}
				// Test each selected file for download speed
				for _, file := range files {
					fileURL := liveStreamSource.FileURI + "/" + file.FileName
					log.Debug().Msg("Testing file url...").Str("file_url", fileURL).Done()
					kbps, err := httpx.TestDownloadSpeed(fileURL)
					if err != nil {
						log.Error().Err(err).Msg("Failed to test download speed for file url, ignore.").Done()
						continue
					}
					// If download speed is too low, skip this file
					if kbps < loadMinSpeed {
						log.Info().Msg("File download speed is too low, ignore.").
							Str("file_url", fileURL).
							Int64("kbps", kbps).
							Str("channel_url", ch.Url).
							Done()
						continue
					}
					// Log successful download speed test
					log.Info().Msg("File download speed is ok.").
						Str("file_url", fileURL).
						Int64("kbps", kbps).
						Str("channel_url", ch.Url).
						Done()

					// Add the channel to the filtered source and break to avoid duplicates
					mu.Lock()
					filteredSource.TvgNameChannels[tvgName] = append(filteredSource.TvgNameChannels[tvgName], ch)
					mu.Unlock()
					log.Info().Msg("Channel is ok.").
						Str("channel_url", ch.Url).
						Str("tvg_name", tvgName).
						Done()
					break
				}
			}
			// Submit the channel test function to the worker pool
			if err := workerPool.Submit(testFunc); err != nil {
				log.Fatal().Msg("Failed to submit test task").
					Err(err).
					Done()
			}
		}
	}

	// Wait for all tests to complete
	wg.Wait()

	return filteredSource
}
