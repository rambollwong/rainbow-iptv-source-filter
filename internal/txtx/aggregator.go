package txtx

import (
	"github.com/rambollwong/rainbowcat/types"
	"github.com/rambollwong/rainbowlog/log"
)

// MergeTvgNameChannels merges multiple TvgNameChannels into one,
// ensuring that channels with the same URL are not duplicated.
// It returns a merged TvgNameChannels map.
func MergeTvgNameChannels(tvgNameChannels []TvgNameChannels) (merged TvgNameChannels) {
	// Initialize the merged result and a set to track existing URLs
	merged = NewTvgNameChannels()
	existUrl := types.NewSet[string]()

	// Iterate over each TvgNameChannels collection
	for _, tvgNameChannel := range tvgNameChannels {
		// Iterate over each TVG name and its associated channels
		for tvgName, channels := range tvgNameChannel {
			// Check if the TVG name already exists in the merged result
			ch, ok := merged[tvgName]
			if !ok {
				// If not, initialize a new slice for channels with initial capacity of 16
				ch = make([]*Channel, 0, 16)
			}

			// Iterate over each channel in the current TVG name's channels
			for _, c := range channels {
				c := c // Create copy to avoid loop variable capture
				// Skip the channel if its URL already exists in the merged result
				if existUrl.Exist(c.Url) {
					log.Debug().Msg("Channel url already exists, skip.").
						Str("url", c.Url).
						Done()
					continue
				}

				// Add the channel's URL to the set of existing URLs
				existUrl.Put(c.Url)
				// Append the channel to the merged result
				ch = append(ch, c)
			}
			// Update the merged result with the processed channels for the current TVG name
			merged[tvgName] = ch
		}
	}
	// Return the merged TvgNameChannels
	return merged
}
