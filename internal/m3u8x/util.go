package m3u8x

import (
	"fmt"
	"strings"
	"time"

	"github.com/rambollwong/rainbow-iptv-source-filter/pkg/proto"
)

// FilterTvgNameOfSource filters the channels in the source based on the provided group list.
// It creates a new map of channels containing only those that match the tvg names specified in the group list.
// Parameters:
//
//	source *ProgramListSource - The source containing all channels grouped by tvg names
//	groupList []*proto.GroupList - The list of groups containing tvg names to filter by
func FilterTvgNameOfSource(source *ProgramListSource, groupList []*proto.GroupList) {
	newTvgNameGroup := make(map[string][]*Channel)
	// Iterate through all group lists
	for _, gl := range groupList {
		// Iterate through all tvg names within the group
		for _, tvgName := range gl.TvgName {
			// Check if the corresponding channel exists in the source
			if chs, ok := source.TvgNameChannels[tvgName]; ok {
				newTvgNameGroup[tvgName] = chs
			}
		}
	}
	source.TvgNameChannels = newTvgNameGroup
}

// FixChannelGroup updates the group information for channels based on the provided group list.
// It iterates through each group and assigns the group title to all channels under that group's tvg-name.
func FixChannelGroup(source *ProgramListSource, groupList []*proto.GroupList) {
	for _, list := range groupList {
		for _, tvgName := range list.TvgName {
			chs, ok := source.TvgNameChannels[tvgName]
			if !ok {
				continue
			}
			for _, ch := range chs {
				ch.Group = list.Group
			}
		}
	}
}

// OutputProgramListSourceToM3u8Bz converts a ProgramListSource into an M3U8 formatted byte slice.
// It includes metadata such as tvg-url, a timestamp channel, and all filtered channels grouped accordingly.
// Parameters:
//   - source: The source containing channel data.
//   - groupList: A list defining how channels should be grouped.
//
// Returns:
//   - A byte slice representing the M3U8 content.
func OutputProgramListSourceToM3u8Bz(source *ProgramListSource, groupList []*proto.GroupList) []byte {
	b := strings.Builder{}
	// Write the M3U header with tvg-url information
	b.WriteString(TagExtm3u)
	b.WriteString(" x-tvg-url=")
	b.WriteString("\"")
	b.WriteString(strings.Join(source.XTvgUrls, "\",\""))
	b.WriteString("\"\n")

	// Add a special channel indicating the update time
	b.WriteString("#EXTINF:-1 tvg-name=\"更新时间\" tvg-logo=\"https://avatars.githubusercontent.com/u/125233001?v=4\" group-title=\"更新时间\",")
	b.WriteString(time.Now().Format("20060102 15:04:05"))
	b.WriteString("\n")
	b.WriteString("http://tc-tct.douyucdn2.cn/dyliveflv1/122402rK7MO9bXSq_2000.flv?wsAuth=8cea39337984fd3341cc9ec569502e4f&token=cpn-androidmpro-0-122402-0fcea45d2300cfa0ac75fafd8679bb53af10de8c33ae99d9&logo=0&expire=0&did=d010b07dcb997ada9934081c873542f0&origin=tct&vhost=p\n")

	// Process each group and its associated channels
	tvgId := 0
	for _, list := range groupList {
		group := list.Group
		for _, tvgName := range list.TvgName {
			tvgId++
			channels, ok := source.TvgNameChannels[tvgName]
			if !ok {
				continue
			}
			for _, channel := range channels {
				// Format each channel line according to M3U8 specification
				channelLine := fmt.Sprintf("#EXTINF:-1 tvg-id=\"%d\" tvg-name=\"%s\" tvg-logo=\"%s\" group-title=\"%s\",%s\n%s\n",
					tvgId, channel.TvgName, channel.TvgLogo, group, channel.Title, channel.Url)
				b.WriteString(channelLine)
			}
		}
	}
	return []byte(b.String())
}
