package txtx

import (
	"reflect"
	"testing"
)

func TestMergeTvgNameChannels(t *testing.T) {
	// Test case 1: Basic merge functionality
	t.Run("basic merge", func(t *testing.T) {
		// Create test data
		channels1 := NewTvgNameChannels()
		channels1["Channel1"] = []*Channel{
			{TvgName: "Channel1", Group: "Group1", Url: "url1"},
		}
		channels1["Channel2"] = []*Channel{
			{TvgName: "Channel2", Group: "Group1", Url: "url2"},
		}

		channels2 := NewTvgNameChannels()
		channels2["Channel3"] = []*Channel{
			{TvgName: "Channel3", Group: "Group2", Url: "url3"},
		}
		channels2["Channel1"] = []*Channel{
			{TvgName: "Channel1", Group: "Group2", Url: "url4"},
		}

		// Perform merge
		result := MergeTvgNameChannels([]TvgNameChannels{channels1, channels2})

		// Verify results
		expected := TvgNameChannels{
			"Channel1": []*Channel{
				{TvgName: "Channel1", Group: "Group1", Url: "url1"},
				{TvgName: "Channel1", Group: "Group2", Url: "url4"},
			},
			"Channel2": []*Channel{
				{TvgName: "Channel2", Group: "Group1", Url: "url2"},
			},
			"Channel3": []*Channel{
				{TvgName: "Channel3", Group: "Group2", Url: "url3"},
			},
		}

		// Compare results
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("MergeTvgNameChannels() = %v, want %v", result, expected)
		}
	})

	// Test case 2: Empty input
	t.Run("empty input", func(t *testing.T) {
		result := MergeTvgNameChannels([]TvgNameChannels{})
		if len(result) != 0 {
			t.Errorf("MergeTvgNameChannels() with empty input should return empty map")
		}
	})

	// Test case 3: Duplicate URL filtering
	t.Run("duplicate url filtering", func(t *testing.T) {
		channels1 := NewTvgNameChannels()
		channels1["Channel1"] = []*Channel{
			{TvgName: "Channel1", Group: "Group1", Url: "url1"},
			{TvgName: "Channel1", Group: "Group1", Url: "url2"},
		}

		channels2 := NewTvgNameChannels()
		channels2["Channel1"] = []*Channel{
			{TvgName: "Channel1", Group: "Group2", Url: "url1"}, // Duplicate URL
			{TvgName: "Channel1", Group: "Group2", Url: "url3"},
		}

		result := MergeTvgNameChannels([]TvgNameChannels{channels1, channels2})

		// Expected result should only contain 3 URLs, because url1 is duplicated
		expectedUrls := map[string]bool{
			"url1": true,
			"url2": true,
			"url3": true,
		}

		// Check merged results
		actualUrls := make(map[string]bool)
		for _, channels := range result {
			for _, channel := range channels {
				actualUrls[channel.Url] = true
			}
		}

		if !reflect.DeepEqual(actualUrls, expectedUrls) {
			t.Errorf("MergeTvgNameChannels() duplicate filtering failed, got urls %v, want %v", actualUrls, expectedUrls)
		}

		// Ensure Channel1 has 3 entries (url1 from first, url2 from first, url3 from second)
		if len(result["Channel1"]) != 3 {
			t.Errorf("MergeTvgNameChannels() should have 3 channels for Channel1, got %d", len(result["Channel1"]))
		}
	})

	// Test case 4: Single TvgNameChannels input
	t.Run("single input", func(t *testing.T) {
		channels := NewTvgNameChannels()
		channels["Channel1"] = []*Channel{
			{TvgName: "Channel1", Group: "Group1", Url: "url1"},
		}
		channels["Channel2"] = []*Channel{
			{TvgName: "Channel2", Group: "Group1", Url: "url2"},
		}

		result := MergeTvgNameChannels([]TvgNameChannels{channels})

		expected := TvgNameChannels{
			"Channel1": []*Channel{
				{TvgName: "Channel1", Group: "Group1", Url: "url1"},
			},
			"Channel2": []*Channel{
				{TvgName: "Channel2", Group: "Group1", Url: "url2"},
			},
		}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("MergeTvgNameChannels() with single input = %v, want %v", result, expected)
		}
	})

	// Test case 5: Multiple same channel names but different groups
	t.Run("multiple same channel names", func(t *testing.T) {
		channels1 := NewTvgNameChannels()
		channels1["Channel1"] = []*Channel{
			{TvgName: "Channel1", Group: "Sports", Url: "sport1"},
		}

		channels2 := NewTvgNameChannels()
		channels2["Channel1"] = []*Channel{
			{TvgName: "Channel1", Group: "News", Url: "news1"},
		}

		channels3 := NewTvgNameChannels()
		channels3["Channel1"] = []*Channel{
			{TvgName: "Channel1", Group: "Movies", Url: "movie1"},
		}

		result := MergeTvgNameChannels([]TvgNameChannels{channels1, channels2, channels3})

		// Should have 3 Channel1 entries, each from a different group
		if len(result["Channel1"]) != 3 {
			t.Errorf("MergeTvgNameChannels() should have 3 channels for Channel1, got %d", len(result["Channel1"]))
		}

		// Check that all URLs are different
		urls := make(map[string]bool)
		for _, channel := range result["Channel1"] {
			if urls[channel.Url] {
				t.Errorf("Duplicate URL found: %s", channel.Url)
			}
			urls[channel.Url] = true
		}
	})
}
