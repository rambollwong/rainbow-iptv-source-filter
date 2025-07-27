package m3u8x

import (
	"github.com/rambollwong/rainbow-iptv-source-filter/pkg/proto"
)

// MapTvgNameGroup maps the group list to a mapping relationship from tvg names to group names
// Parameters:
//
//	groupList []*proto.GroupList - The list of groups, containing group information and corresponding tvg name lists
//
// Returns:
//
//	map[string]string - Mapping from tvg names to group names, where key is the tvg name and value is the corresponding group name
func MapTvgNameGroup(groupList []*proto.GroupList) map[string]string {
	tvgNameGroup := make(map[string]string)
	for _, gl := range groupList {
		for _, tvgName := range gl.TvgName {
			tvgNameGroup[tvgName] = gl.Group
		}
	}
	return tvgNameGroup
}

// BuildTargetSource constructs a target source data structure based on the provided group list
// Parameters:
//
//	groupList - A slice of pointers to GroupList containing grouping information with tvg names
//
// Returns:
//
//	target - The constructed ProgramListSource structure containing channels mapped by tvg names
func BuildTargetSource(groupList []*proto.GroupList) (target ProgramListSource) {
	// Initialize the TvgNameChannels map, creating an empty channel slice for each tvg name
	target.TvgNameChannels = make(map[string][]*Channel)
	for _, gl := range groupList {
		for _, tvgName := range gl.TvgName {
			target.TvgNameChannels[tvgName] = make([]*Channel, 0, 16)
		}
	}
	return target
}
