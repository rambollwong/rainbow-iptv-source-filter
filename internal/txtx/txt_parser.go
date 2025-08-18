package txtx

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/rambollwong/rainbowlog/log"
)

const (
	TxtGenre = "#genre#"
)

// Channel represents a TV channel with its metadata
type Channel struct {
	TvgName string // TvgName is the name of the channel
	Group   string // Group is the category or group the channel belongs to
	Url     string // Url is the streaming URL of the channel
}

// TvgNameChannels is a map where the key is the channel name and the value is a slice of Channel pointers
type TvgNameChannels map[string][]*Channel

// NewTvgNameChannels creates and returns a new instance of TvgNameChannels
func NewTvgNameChannels() TvgNameChannels {
	return make(TvgNameChannels)
}

// ParseTxt parses the given source byte slice which contains channel data in a specific format
// and populates the TvgNameChannels map with the parsed information.
func (t TvgNameChannels) ParseTxt(source []byte) {
	scanner := bufio.NewScanner(bytes.NewReader(source))
	var currentGroup string // Holds the current group name while parsing

	// Iterate through each line in the source
	lineNo := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNo++
		line = strings.TrimSpace(line) // Remove leading and trailing whitespaces

		// Skip empty lines
		if line == "" {
			continue
		}

		// Split the line by comma to extract channel information
		arr := strings.Split(line, ",")
		if len(arr) != 2 {
			log.Debug().Msg("invalid line.").Int("line_no", lineNo).Str("line", line).Done()
			continue // Skip lines that don't have exactly two parts
		}

		// Check if the second part indicates a genre/group marker
		if arr[1] == TxtGenre {
			currentGroup = arr[0] // Set the current group name
			continue
		}

		// Create a new Channel instance with the parsed data
		channel := &Channel{
			TvgName: strings.ReplaceAll(strings.ToUpper(arr[0]), "-", ""),
			Group:   currentGroup,
			Url:     arr[1],
		}
		channel.Url = strings.TrimSpace(channel.Url)
		if dlIdx := strings.Index(channel.Url, "$"); dlIdx != -1 {
			channel.Url = channel.Url[:dlIdx]
		}

		// Initialize the slice for this channel name if it doesn't exist
		if _, ok := t[channel.TvgName]; !ok {
			t[channel.TvgName] = make([]*Channel, 0, 16)
		}

		// Append the new channel to the slice for this channel name
		t[channel.TvgName] = append(t[channel.TvgName], channel)
	}
}
