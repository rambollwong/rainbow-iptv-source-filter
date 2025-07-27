package main

import (
	"fmt"
	"os"

	"github.com/rambollwong/rainbow-iptv-source-filter/conf"
	"github.com/rambollwong/rainbow-iptv-source-filter/internal/filex"
	"github.com/rambollwong/rainbow-iptv-source-filter/internal/httpx"
	"github.com/rambollwong/rainbow-iptv-source-filter/internal/logx"
	"github.com/rambollwong/rainbow-iptv-source-filter/internal/m3u8x"
	"github.com/rambollwong/rainbowlog/log"
)

func main() {
	logx.InitGlobalLogger()

	log.Info().Msg("Starting rainbow-iptv-source-filter").Done()

	if err := conf.InitConfig(); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize config").Done()
	}
	if conf.Config.CustomUA != "" {
		httpx.UA = conf.Config.CustomUA
		log.Info().Msg("Use custom UA.").Str("ua", conf.Config.CustomUA).Done()
	}

	// 1. build target Source
	targetSource := m3u8x.BuildTargetSource(conf.Config.GroupList)
	filteredSources := make([]*m3u8x.ProgramListSource, 0, 16)
	tvgNameGroupMap := m3u8x.MapTvgNameGroup(conf.Config.GroupList)

	// 2. test each of sources
	localPath := conf.Config.ProgramListSourceFileLocalPath
	if localPath != "" {
		// search local files
		log.Info().Msg("Searching local m3u8 files...").Str("path", localPath).Done()
		files, err := filex.SearchFilesBySuffix(localPath, ".m3u8")
		if err != nil {
			log.Error().Err(err).Msg("Failed to search files, ignore").Done()
		}
		if len(files) > 0 {
			log.Info().Msg("Found files").Strs("files", files...).Done()
		}
		for _, file := range files {
			log.Info().Msg("Processing local m3u8 file...").Str("file", file).Done()
			// read file
			fileBytes, err := os.ReadFile(file)
			if err != nil {
				log.Error().Err(err).Msg("Failed to read file, ignore").Done()
				continue
			}
			//parse file to source
			newSource := m3u8x.NewProgramListSource()
			if err := newSource.ParseProgramListSource(fileBytes); err != nil {
				log.Error().Err(err).Msg("Failed to parse file, ignore").Done()
				continue
			}

			// test new source
			filteredSource := m3u8x.TestAndFilterProgramListSource(
				newSource, tvgNameGroupMap,
				conf.Config.TestPingMinLatency,
				conf.Config.TestLoadMinSpeed,
				conf.Config.RetryTimes)
			filteredSources = append(filteredSources, filteredSource)
		}
	}

	sourceUrls := conf.Config.ProgramListSourceUrls
	if len(sourceUrls) > 0 {
		log.Info().Msg("Testing remote m3u8 files...").Strs("urls", sourceUrls...).Done()
	}
	for _, sourceUrl := range sourceUrls {
		log.Info().Msg("Processing remote m3u8 file...").Str("url", sourceUrl).Done()
		sourceContent, err := httpx.LoadUrlContentWithRetry(sourceUrl, conf.Config.RetryTimes)
		if err != nil {
			log.Error().Err(err).Msg("Failed to load url, ignore").Done()
			continue
		}
		log.Debug().Msg("Loaded url content").Done()

		// parse url to source
		newSource := m3u8x.NewProgramListSource()
		if err := newSource.ParseProgramListSource(sourceContent); err != nil {
			log.Error().Err(err).Msg("Failed to parse url, ignore").Done()
			continue
		}
		log.Debug().Msg("Parsed url content").Done()

		// test new source
		filteredSource := m3u8x.TestAndFilterProgramListSource(
			newSource, tvgNameGroupMap,
			conf.Config.TestPingMinLatency,
			conf.Config.TestLoadMinSpeed,
			conf.Config.RetryTimes)
		fmt.Printf("%+v\n", filteredSource)
		filteredSources = append(filteredSources, filteredSource)
	}

	// 3. merger all filtered sources to target
	m3u8x.MergeProgramListSources(filteredSources, &targetSource)
	// 4. output to the result file
	fmt.Printf("%+v\n", targetSource)
}
