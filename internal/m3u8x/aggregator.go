package m3u8x

import (
	"strings"

	"github.com/rambollwong/rainbowcat/types"
	"github.com/rambollwong/rainbowcat/util"
	"github.com/rambollwong/rainbowlog/log"
)

// MergeProgramListSources merges multiple ProgramListSource instances into a single one.
// It combines XTvgUrls and channels grouped by TvgName, ensuring no duplicate URLs
// and validating that TvgName matches Title before merging.
func MergeProgramListSources(sources []*ProgramListSource) (merged *ProgramListSource) {
	merged = NewProgramListSource()
	existUrl := types.NewSet[string]()
	for _, programListSource := range sources {
		// Merge program list URLs
		merged.XTvgUrls = append(merged.XTvgUrls, programListSource.XTvgUrls...)
		merged.XTvgUrls = util.SliceUnion(merged.XTvgUrls)

		// Merge channels by TvgName
		for tvgName, channels := range programListSource.TvgNameChannels {
			ch, ok := merged.TvgNameChannels[tvgName]
			if !ok {
				ch = make([]*Channel, 0, 16)
				merged.TvgNameChannels[tvgName] = ch
			}
			for _, c := range channels {
				c := c // Create copy to avoid loop variable capture
				// Validate TvgName and Title match before proceeding
				if !checkTvgAndTitle(c.TvgName, c.Title) {
					log.Warn().Msg("TvgName does not match Title, ignoring").
						Str("tvg_name", c.TvgName).
						Str("title", c.Title).
						Done()
					continue
				}
				// Normalize Title to match TvgName
				c.Title = c.TvgName
				// Skip if URL already exists
				if existUrl.Exist(c.Url) {
					log.Warn().Msg("Channel url already exists, skip.").
						Str("url", c.Url).
						Done()
					continue
				}
				// Add URL to set and append channel
				existUrl.Put(c.Url)
				ch = append(ch, c)
				merged.TvgNameChannels[tvgName] = ch
			}
		}
	}
	return merged
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
