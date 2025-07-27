package m3u8x

import (
	"strings"

	"github.com/rambollwong/rainbowlog/log"
)

// MergeProgramListSources merges multiple ProgramListSource instances into a single target.
// It combines XTvgUrls from all sources into the target, and merges channels with matching TvgNames.
// Channels are only merged if their TvgName and Title pass validation via checkTvgAndTitle.
// Existing channels in the target with the same TvgName will be appended with matching channels from other sources.
func MergeProgramListSources(sources []*ProgramListSource, target *ProgramListSource) {
	for _, programListSource := range sources {
		// Merge program list URLs
		target.XTvgUrls = append(target.XTvgUrls, programListSource.XTvgUrls...)

		// Merge channels by TvgName
		for tvgName, channels := range target.TvgNameChannels {
			ch, ok := programListSource.TvgNameChannels[tvgName]
			if !ok {
				log.Warn().Msg("Channel not found for tvg").Str("tvg_name", tvgName).Done()
				continue
			}

			// Validate and append each channel
			for _, c := range ch {
				c := c // Create copy to avoid loop variable capture
				if !checkTvgAndTitle(c.TvgName, c.Title) {
					log.Warn().Msg("TvgName does not match Title, ignoring").
						Str("tvg_name", c.TvgName).
						Str("title", c.Title).
						Done()
					continue
				}
				c.Title = c.TvgName
				channels = append(channels, c)
			}
		}
	}
}

// checkTvgAndTitle validates if a channel's TvgName and Title are semantically equivalent.
// It normalizes both strings by converting to uppercase and removing hyphens before comparison.
// Returns true if normalized strings match, false otherwise.
func checkTvgAndTitle(tvgName, title string) bool {
	return strings.ReplaceAll(
		strings.ToUpper(tvgName),
		"-",
		"") == strings.ReplaceAll(strings.ToUpper(title), "-", "")
}
