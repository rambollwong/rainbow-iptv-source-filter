package main

import (
	"context"
	"fmt"
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
	"github.com/spf13/pflag"
)

var (
	version string
)

func main() {
	if version == "" {
		version = "dev"
	}
	vFlag := pflag.BoolP("version", "v", false, "show version")
	hFlag := pflag.BoolP("help", "h", false, "show help")
	pflag.Parse()
	if *vFlag {
		fmt.Printf("version: %s\n", version)
		os.Exit(0)
	}
	if *hFlag {
		printHelp()
		os.Exit(0)
	}

	printLogo()

	logx.InitGlobalLogger()
	defer func() {
		if err := log.Logger.Flush(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to flush log: %v\n", err)
		}
	}()

	log.Info().Msg("Starting rainbow-iptv-source-filter").Done()

	if err := conf.InitConfig(); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize config").Done()
	}
	if conf.Config.CustomUA != "" {
		httpx.UA = conf.Config.CustomUA
		log.Info().Msg("Use custom UA.").Str("ua", conf.Config.CustomUA).Done()
	}

	ctx, cancel := context.WithCancel(context.Background())
	workerPool := pool.NewWorkerPool(int(conf.Config.ParallelExecutorNum), pool.WithContext(ctx))
	defer workerPool.Close()

	go mainLogic(ctx, cancel, workerPool)

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

func mainLogic(ctx context.Context, cancel context.CancelFunc, workerPool *pool.WorkerPool) {
	defer cancel()
	// worker pool
	log.Info().Int64("parallel_executor_num", conf.Config.ParallelExecutorNum).Done()
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
				log.Debug().Err(err).Msg("Failed to submit task func").Done()
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
			sourceContent, err := httpx.LoadUrlContentWithRetry(ctx, sourceUrl, conf.Config.RetryTimes)
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
			log.Debug().Err(err).Msg("Failed to submit task func").Done()
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
		ctx,
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
	log.Info().Msg("All done.").Done()
}

func printLogo() {
	logo := `

▗▄▄▄▖▗▄▄▖▗▄▄▄▖▗▖  ▗▖     ▗▄▄▖ ▗▄▖ ▗▖ ▗▖▗▄▄▖  ▗▄▄▖▗▄▄▄▖    ▗▄▄▄▖▗▄▄▄▖▗▖ ▗▄▄▄▖▗▄▄▄▖▗▄▄▖ 
  █  ▐▌ ▐▌ █  ▐▌  ▐▌    ▐▌   ▐▌ ▐▌▐▌ ▐▌▐▌ ▐▌▐▌   ▐▌       ▐▌     █  ▐▌   █  ▐▌   ▐▌ ▐▌
  █  ▐▛▀▘  █  ▐▌  ▐▌     ▝▀▚▖▐▌ ▐▌▐▌ ▐▌▐▛▀▚▖▐▌   ▐▛▀▀▘    ▐▛▀▀▘  █  ▐▌   █  ▐▛▀▀▘▐▛▀▚▖
▗▄█▄▖▐▌    █   ▝▚▞▘     ▗▄▄▞▘▝▚▄▞▘▝▚▄▞▘▐▌ ▐▌▝▚▄▄▖▐▙▄▄▖    ▐▌   ▗▄█▄▖▐▙▄▄▖█  ▐▙▄▄▖▐▌ ▐▌
--------------------------------------------------------------------------------------
Version: %s


`
	fmt.Printf(logo, version)
}

func printHelp() {
	helpText := fmt.Sprintf(`Usage: rainbow-iptv-source-filterd [options]

Options:
  -v, --version          Show version information
  -h, --help             Show this help message
  -c, --config.path      Config file path (default "./conf")
  -l, --local-path       Path of local program list source file
  -o, --output           Output file path

Description:
  This tool filters and processes IPTV source lists in M3U8 format. It can read from local files or remote URLs, test stream availability, and generate a merged, filtered output.

Configuration:
  The tool reads configuration from a YAML file. Please ensure the config file is properly set up before running.

Examples:
  rainbow-iptv-source-filterd              				# Run with default settings
  rainbow-iptv-source-filterd -v           				# Show version
  rainbow-iptv-source-filterd -h           				# Show this help
  rainbow-iptv-source-filterd -c ./config  				# Specify config path
  rainbow-iptv-source-filterd -l ./sources 				# Specify local source path
  rainbow-iptv-source-filterd -o ./result  				# Specify output path
  rainbow-iptv-source-filterd -c ./config -o ./result  	# Specify config path and output path

For more information, please visit the project repository.
`)
	fmt.Print(helpText)
}
