package m3u8x

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rambollwong/rainbow-iptv-source-filter/internal/httpx"
	"github.com/rambollwong/rainbowlog/log"
)

const (
	TagExtm3u             = "#EXTM3U"
	TagExtinf             = "#EXTINF"
	TagExtXVersion        = "#EXT-X-VERSION"
	TagExtXMediaSequence  = "#EXT-X-MEDIA-SEQUENCE"
	TagExtXTargetDuration = "#EXT-X-TARGETDURATION"
)

type Channel struct {
	TvgName string // TvgName is the name of channel
	TvgLogo string // TvgLogo is the logo url of channel
	Group   string // Group of the channel
	Title   string // Title of the channel, usually consistent with TvgName
	Url     string // Url of the channel's live source
}

type ProgramListSource struct {
	XTvgUrls        []string              // XTvgUrls means the Live Program List
	TvgNameChannels map[string][]*Channel // TvgNameChannels are channels that grouped by TvgName
}

func NewProgramListSource() *ProgramListSource {
	return &ProgramListSource{
		XTvgUrls:        nil,
		TvgNameChannels: make(map[string][]*Channel),
	}
}

func (s *ProgramListSource) ParseProgramListSource(source []byte) (err error) {
	scanner := bufio.NewScanner(bytes.NewReader(source))
	lineNo := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNo++
		if lineNo == 1 {
			// read program list config
			s.XTvgUrls, err = readXTvgUrlsFromLine(line)
			if err != nil {
				return err
			}
			continue
		}

		// read info of channel
		channel := &Channel{}
		if !channel.readInfoFromLine(line) {
			log.Warn().Msg("Can not read tvg info from line.").
				Int("line_no", lineNo).Str("line", line).Done()
			continue
		}
		if scanner.Scan() {
			line = scanner.Text()
			lineNo++
			// read live stream url
			channel.Url = strings.TrimSpace(line)
			if dlIdx := strings.Index(channel.Url, "$"); dlIdx != -1 {
				channel.Url = channel.Url[:dlIdx]
			}
			s.TvgNameChannels[channel.TvgName] = append(s.TvgNameChannels[channel.TvgName], channel)

		}
	}
	return scanner.Err()
}

func readXTvgUrlsFromLine(line string) ([]string, error) {
	arr := strings.Split(line, " ")
	if len(arr) != 2 {
		return nil, fmt.Errorf("invalid x-tvg-url line: %s", line)
	}
	if arr[0] != TagExtm3u {
		return nil, fmt.Errorf("invalid tag of x-tvg-url line: %s", line)
	}
	arr = strings.Split(arr[1], "=")
	if strings.ToLower(arr[0]) != "x-tvg-url" {
		return nil, fmt.Errorf("invalid x-tvg-url line: %s", line)
	}
	return strings.Split(strings.ReplaceAll(arr[1], "\"", ""), ","), nil
}

func (c *Channel) readInfoFromLine(line string) bool {
	if !strings.HasPrefix(line, TagExtinf) {
		return false
	}
	arr := strings.Split(line, " ")
	for i, s := range arr {
		if i == 0 {
			continue
		}

		infoArr := strings.Split(s, "=")
		switch infoArr[0] {
		case "tvg-name":
			c.TvgName = strings.ReplaceAll(strings.ToUpper(infoArr[1]), "-", "")
			c.TvgName = strings.ReplaceAll(c.TvgName, "\"", "")
		case "tvg-logo":
			c.TvgLogo = strings.ReplaceAll(infoArr[1], "\"", "")
		case "tvg-id":
			// ignore
		case "group-title":
			groupTitle := strings.Split(infoArr[1], ",")
			c.Group, c.Title = groupTitle[0], strings.ToUpper(groupTitle[1])
			c.Group = strings.ReplaceAll(c.Group, "\"", "")
			c.Title = strings.ReplaceAll(c.Title, "\"", "")
		case "":
			// ignore
		default:
			//log.Debug().Msg("Unknown parameter of EXTINF").Str("parameter", infoArr[0]).Done()
		}
	}
	return true
}

type LiveStreamFile struct {
	ExtInf   string // ExtInf is the value of #EXTINF
	FileName string // FileName of the live stream file
}

type LiveStreamSource struct {
	ExtXVersion        string           // ExtXVersion is the value of #EXT-X-VERSION
	ExtXMediaSequence  string           // ExtXMediaSequence is the value of #EXT-X-MEDIA-SEQUENCE
	ExtXTargetDuration string           // ExtXTargetDuration is the value of #EXT-X-TARGETDURATION
	FileURI            string           // FileURI of the live stream file
	Files              []LiveStreamFile // Files for the live stream
}

func NewLiveStreamSource(sourceURL string) *LiveStreamSource {
	return &LiveStreamSource{
		FileURI: sourceURL[:strings.LastIndex(sourceURL, "/")],
	}
}

func (s *LiveStreamSource) ParseLiveStreamSource(ctx context.Context, source []byte) error {
	scanner := bufio.NewScanner(bytes.NewReader(source))
	lineNo := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNo++
		if strings.HasPrefix(line, TagExtm3u) {
			continue
		} else if strings.HasPrefix(line, TagExtXVersion) {
			s.ExtXVersion = strings.TrimPrefix(line, TagExtXVersion+":")
		} else if strings.HasPrefix(line, TagExtXMediaSequence) {
			s.ExtXMediaSequence = strings.TrimPrefix(line, TagExtXVersion+":")
		} else if strings.HasPrefix(line, TagExtXTargetDuration) {
			s.ExtXTargetDuration = strings.TrimPrefix(line, TagExtXTargetDuration+":")
		} else if strings.HasPrefix(line, TagExtinf) {
			file := LiveStreamFile{}
			file.ExtInf = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, TagExtinf+":"), ","))
			if !scanner.Scan() {
				return errors.New("source information is incomplete")
			}
			line = scanner.Text()
			lineNo++
			file.FileName = strings.TrimSpace(line)
			s.Files = append(s.Files, file)
		} else if strings.HasPrefix(line, "#") {
			log.Debug().Msg("unknown tag of line").Int("line_no", lineNo).Str("line", line).Done()
		} else if strings.HasSuffix(line, ".m3u8") {
			urlContent, err := httpx.LoadUrlContent(ctx, line)
			if err != nil {
				log.Error().Msg("Failed to load channel url, ignore.").
					Str("channel_url", line).
					Err(err).
					Done()
				return err
			}
			s.FileURI = line[:strings.LastIndex(line, "/")]
			return s.ParseLiveStreamSource(ctx, urlContent)
		} else {
			return errors.New("source information is incomplete")
		}
	}
	return nil
}
