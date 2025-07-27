package m3u8x

import (
	"github.com/rambollwong/rainbow-iptv-source-filter/internal/httpx"
	"github.com/rambollwong/rainbowlog/log"
)

// TestAndFilterProgramListSource tests and filters program list source based on latency and download speed
// source: the original program list source to be tested and filtered
// tvgNameGroupMap: map of tvg names to groups for filtering channels
// minLatency: maximum allowed latency in milliseconds
// loadMinSpeed: minimum required download speed in kbps
// retryTimes: number of retries when loading channel URL
// Returns a filtered program list source containing only valid entries
func TestAndFilterProgramListSource(
	source *ProgramListSource,
	tvgNameGroupMap map[string]string,
	minLatency, loadMinSpeed, retryTimes int64) (filteredSource *ProgramListSource) {
	filteredSource = NewProgramListSource()

	// Test x tvg urls for latency
	for _, tvgUrl := range source.XTvgUrls {
		latency, err := httpx.PingURL(tvgUrl)
		if err != nil {
			log.Error().Err(err).Msg("Failed to ping tvg url, ignore.").Done()
			continue
		}
		if latency > minLatency {
			log.Info().Msg("tvg url latency is too long, ignore.").
				Str("tvg_url", tvgUrl).
				Int64("latency", latency).
				Done()
			continue
		}
		log.Info().Msg("tvg url latency is ok.").
			Str("tvg_url", tvgUrl).
			Int64("latency", latency).
			Done()
		filteredSource.XTvgUrls = append(filteredSource.XTvgUrls, tvgUrl)
	}

	// Test program list channels
	for tvgName, _ := range tvgNameGroupMap {
		chs, ok := source.TvgNameChannels[tvgName]
		if !ok {
			log.Info().Msg("channel not found for tvg name").Str("tvgName", tvgName).Done()
			continue
		}

		// Initialize the channel slice if not exists
		if _, exists := filteredSource.TvgNameChannels[tvgName]; !exists {
			filteredSource.TvgNameChannels[tvgName] = make([]*Channel, 0, 8)
		}

		for idx, ch := range chs {
			log.Info().Msg("Testing channel url...").
				Str("tvg_name", tvgName).
				Int("index", idx).
				Str("channel_url", ch.Url).
				Done()

			latency, err := httpx.PingURL(ch.Url)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to ping channel url, skip.").Done()
			} else {
				if latency > minLatency {
					log.Info().Msg("Channel url latency is too long, ignore.").
						Str("channel_url", ch.Url).
						Int64("latency", latency).
						Done()
					continue
				}
				log.Debug().Msg("Channel url latency is ok.").
					Str("channel_url", ch.Url).
					Int64("latency", latency).
					Done()
			}

			urlContent, err := httpx.LoadUrlContentWithRetry(ch.Url, retryTimes)
			if err != nil {
				log.Error().Err(err).Msg("Failed to load channel url, ignore.").Done()
				continue // Add continue to skip to next channel when load fails
			}

			liveStreamSource := NewLiveStreamSource(ch.Url)
			if err := liveStreamSource.ParseLiveStreamSource(urlContent); err != nil {
				log.Error().Err(err).Msg("Failed to parse live stream source for channel url, ignore.").Done()
				continue
			}
			filesCount := len(liveStreamSource.Files)
			startIdx := 0
			if filesCount > 2 {
				// only the latest two files are tested here. If both fail, skip the source.
				startIdx = filesCount - 2
			}
			for i := startIdx; i < filesCount; i++ {
				file := liveStreamSource.Files[i]
				fileURL := liveStreamSource.FileURI + "/" + file.FileName
				log.Debug().Msg("Testing file url...").Str("file_url", fileURL).Done()
				kbps, err := httpx.TestDownloadSpeed(fileURL)
				if err != nil {
					log.Error().Err(err).Msg("Failed to test download speed for file url, ignore.").Done()
					continue
				}
				if kbps < loadMinSpeed {
					log.Info().Msg("File download speed is too low, ignore.").
						Str("file_url", fileURL).
						Int64("kbps", kbps).
						Str("channel_url", ch.Url).
						Done()
					continue
				}
				log.Info().Msg("File download speed is ok.").
					Str("file_url", fileURL).
					Int64("kbps", kbps).
					Str("channel_url", ch.Url).
					Done()

				filteredSource.TvgNameChannels[tvgName] = append(filteredSource.TvgNameChannels[tvgName], ch)
				log.Info().Msg("Channel is ok.").
					Str("channel_url", ch.Url).
					Str("tvg_name", tvgName).
					Done()
				break
			}
		}
	}

	return filteredSource
}
