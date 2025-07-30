package main

import (
	"context"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

	"github.com/rambollwong/rainbow-iptv-source-filter/conf"
	"github.com/rambollwong/rainbow-iptv-source-filter/internal/filex"
	"github.com/rambollwong/rainbow-iptv-source-filter/internal/httpx"
	"github.com/rambollwong/rainbow-iptv-source-filter/internal/logx"
	"github.com/rambollwong/rainbow-iptv-source-filter/internal/m3u8x"
	"github.com/rambollwong/rainbowcat/pool"
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

	ctx, cancel := context.WithCancel(context.Background())

	go mainLogic(ctx, cancel)

	// Graceful shutdown
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-ctx.Done():
			return
		case <-signals:
			cancel()
		}
	}()

	<-ctx.Done()
}

func mainLogic(ctx context.Context, cancel context.CancelFunc) {
	// worker pool
	log.Info().Int64("parallel_executor_num", conf.Config.ParallelExecutorNum).Done()
	workerPool := pool.NewWorkerPool(int(conf.Config.ParallelExecutorNum), pool.WithContext(ctx))
	wg := &sync.WaitGroup{}

	newFilteredSources := make([]*m3u8x.ProgramListSource, 0, 16)
	newFilteredSourcesMutex := &sync.Mutex{}

	localPath := conf.Config.ProgramListSourceFileLocalPath
	groupList := conf.Config.GroupList
	if localPath != "" {
		// search local files
		log.Info().Msg("Searching local m3u8 files...").Str("path", localPath).Done()
		files, err := filex.SearchFilesBySuffix(localPath, ".m3u8")
		if err != nil {
			log.Error().Msg("Failed to search files, ignore").Err(err).Done()
		}
		if len(files) > 0 {
			log.Info().Msg("Found files").Strs("files", files...).Done()
		}
		for _, file := range files {
			wg.Add(1)
			taskFunc := func() {
				defer wg.Done()
				log.Info().Msg("Processing local m3u8 file...").Str("file", file).Done()
				// read file
				fileBytes, err := os.ReadFile(file)
				if err != nil {
					log.Error().Msg("Failed to read file, ignore").Str("file", file).Err(err).Done()
					return
				}
				log.Debug().Msg("Loaded local file.").Str("file", file).Done()

				//parse file to source
				newSource := m3u8x.NewProgramListSource()
				if err := newSource.ParseProgramListSource(fileBytes); err != nil {
					log.Error().Msg("Failed to parse file, ignore").Str("file", file).Err(err).Done()
					return
				}
				log.Debug().Msg("Parsed local file.").Str("file", file).Done()

				m3u8x.FilterTvgNameOfSource(newSource, groupList)
				newFilteredSourcesMutex.Lock()
				newFilteredSources = append(newFilteredSources, newSource)
				newFilteredSourcesMutex.Unlock()
			}
			err := workerPool.Submit(taskFunc)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to submit task func").Done()
				return
			}

		}
	}

	sourceUrls := conf.Config.ProgramListSourceUrls
	if len(sourceUrls) > 0 {
		log.Info().Msg("Testing remote m3u8 files...").Strs("urls", sourceUrls...).Done()
	}
	for _, sourceUrl := range sourceUrls {
		wg.Add(1)
		taskFunc := func() {
			defer wg.Done()
			log.Info().Msg("Processing remote m3u8 file...").Str("url", sourceUrl).Done()
			sourceContent, err := httpx.LoadUrlContentWithRetry(sourceUrl, conf.Config.RetryTimes)
			if err != nil {
				log.Error().Msg("Failed to load url, ignore").Err(err).Done()
				return
			}
			log.Debug().Msg("Loaded url content").Str("url", sourceUrl).Done()

			// parse url to source
			newSource := m3u8x.NewProgramListSource()
			if err := newSource.ParseProgramListSource(sourceContent); err != nil {
				log.Error().Msg("Failed to parse url, ignore").Str("url", sourceUrl).Err(err).Done()
				return
			}
			log.Debug().Msg("Parsed url content").Str("url", sourceUrl).Done()
			m3u8x.FilterTvgNameOfSource(newSource, groupList)
			newFilteredSourcesMutex.Lock()
			newFilteredSources = append(newFilteredSources, newSource)
			newFilteredSourcesMutex.Unlock()
		}
		err := workerPool.Submit(taskFunc)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to submit task func").Done()
			return
		}
	}
	// wait all tasks done
	wg.Wait()

	// merge all filtered sources
	mergedSource := m3u8x.MergeProgramListSources(newFilteredSources)
	log.Info().Msg("Merge all sources successfully.").Done()

	// test merged source
	targetSource := m3u8x.ParallelTestProgramListSource(
		mergedSource,
		conf.Config.TestPingMinLatency,
		conf.Config.TestLoadMinSpeed,
		conf.Config.RetryTimes,
		workerPool, groupList)
	log.Info().Msg("All source tests are completed.").Done()

	// fix channel group
	m3u8x.FixChannelGroup(targetSource, groupList)

	// output to the result file
	log.Info().Msg("Writing the final source to the file...").
		Str("output_file", conf.Config.OutputFile).
		Done()
	outputBz := m3u8x.OutputProgramListSourceToM3u8Bz(targetSource, groupList)
	outputFile := path.Join(conf.Config.OutputFile)
	if path.Ext(outputFile) != ".m3u8" {
		outputFile += ".m3u8"
	}
	err := filex.WriteBytesToFile(outputBz, outputFile)
	if err != nil {
		log.Fatal().Msg("Failed to write to file.").Err(err).Done()
	}
	log.Info().Msg("The file writing is completed.").Done()

	workerPool.Close()
	cancel()
	log.Info().Msg("All done.").Done()
}
