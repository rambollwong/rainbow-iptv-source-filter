package txtx

import (
	"reflect"
	"testing"
)

func TestNewTvgNameChannels(t *testing.T) {
	tvc := NewTvgNameChannels()
	if tvc == nil {
		t.Error("NewTvgNameChannels() should not return nil")
	}

	if reflect.TypeOf(tvc).Kind() != reflect.Map {
		t.Error("NewTvgNameChannels() should return a map type")
	}
}

func TestTvgNameChannels_ParseTxt(t *testing.T) {
	tests := []struct {
		name     string
		source   []byte
		expected TvgNameChannels
	}{
		{
			name: "basic parsing",
			source: []byte(`Group1,#genre#
Channel1,url1
Channel2,url2
Group2,#genre#
Channel3,url3`),
			expected: TvgNameChannels{
				"CHANNEL1": []*Channel{
					{TvgName: "CHANNEL1", Group: "Group1", Url: "url1"},
				},
				"CHANNEL2": []*Channel{
					{TvgName: "CHANNEL2", Group: "Group1", Url: "url2"},
				},
				"CHANNEL3": []*Channel{
					{TvgName: "CHANNEL3", Group: "Group2", Url: "url3"},
				},
			},
		},
		{
			name:     "empty source",
			source:   []byte(""),
			expected: TvgNameChannels{},
		},
		{
			name: "source with empty lines",
			source: []byte(`
Group1,#genre#

Channel1,url1

Channel2,url2

`),
			expected: TvgNameChannels{
				"CHANNEL1": []*Channel{
					{TvgName: "CHANNEL1", Group: "Group1", Url: "url1"},
				},
				"CHANNEL2": []*Channel{
					{TvgName: "CHANNEL2", Group: "Group1", Url: "url2"},
				},
			},
		},
		{
			name: "source with invalid lines",
			source: []byte(`Group1,#genre#
Channel1,url1
invalid_line
Channel2,url2
another,invalid,line`),
			expected: TvgNameChannels{
				"CHANNEL1": []*Channel{
					{TvgName: "CHANNEL1", Group: "Group1", Url: "url1"},
				},
				"CHANNEL2": []*Channel{
					{TvgName: "CHANNEL2", Group: "Group1", Url: "url2"},
				},
			},
		},
		{
			name: "multiple channels with same name",
			source: []byte(`Group1,#genre#
Channel1,url1
Group2,#genre#
Channel1,url2`),
			expected: TvgNameChannels{
				"CHANNEL1": []*Channel{
					{TvgName: "CHANNEL1", Group: "Group1", Url: "url1"},
					{TvgName: "CHANNEL1", Group: "Group2", Url: "url2"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tvc := NewTvgNameChannels()
			tvc.ParseTxt(tt.source)

			if len(tvc) != len(tt.expected) {
				t.Errorf("ParseTxt() got %d channel names, want %d", len(tvc), len(tt.expected))
			}

			for channelName, channels := range tt.expected {
				gotChannels, exists := tvc[channelName]
				if !exists {
					t.Errorf("ParseTxt() missing channel name %s", channelName)
					continue
				}

				if len(gotChannels) != len(channels) {
					t.Errorf("ParseTxt() got %d channels for %s, want %d", len(gotChannels), channelName, len(channels))
					continue
				}

				for i, expectedChannel := range channels {
					gotChannel := gotChannels[i]
					if !reflect.DeepEqual(gotChannel, expectedChannel) {
						t.Errorf("ParseTxt() got channel %v, want %v", gotChannel, expectedChannel)
					}
				}
			}
		})
	}
}
