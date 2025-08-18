package txtx

import (
	"strings"
	"testing"

	"github.com/rambollwong/rainbow-iptv-source-filter/internal/m3u8x"
	"github.com/rambollwong/rainbow-iptv-source-filter/pkg/proto"
)

func TestToM3u(t *testing.T) {
	// Test case 1: Basic conversion functionality
	t.Run("basic conversion", func(t *testing.T) {
		// Create test data
		txt := NewTvgNameChannels()
		txt["Channel1"] = []*Channel{
			{TvgName: "Channel1", Group: "Group1", Url: "url1"},
		}
		txt["Channel2"] = []*Channel{
			{TvgName: "Channel2", Group: "Group1", Url: "url2"},
		}

		// Perform conversion
		result := ToM3u(txt)

		// Verify results
		if result == nil {
			t.Fatal("ToM3u() should not return nil")
		}

		if len(result.TvgNameChannels) != 2 {
			t.Errorf("ToM3u() should have 2 channels, got %d", len(result.TvgNameChannels))
		}

		// Check Channel1
		channels1, ok := result.TvgNameChannels["Channel1"]
		if !ok {
			t.Error("ToM3u() should contain Channel1")
		} else {
			if len(channels1) != 1 {
				t.Errorf("ToM3u() Channel1 should have 1 entry, got %d", len(channels1))
			} else {
				channel := channels1[0]
				if channel.TvgName != "Channel1" || channel.Group != "Group1" || channel.Url != "url1" || channel.Title != "Channel1" {
					t.Errorf("ToM3u() Channel1 has wrong data: %+v", channel)
				}
				// TvgLogo should be empty as set in the function
				if channel.TvgLogo != "" {
					t.Errorf("ToM3u() Channel1 TvgLogo should be empty, got %s", channel.TvgLogo)
				}
			}
		}

		// Check Channel2
		channels2, ok := result.TvgNameChannels["Channel2"]
		if !ok {
			t.Error("ToM3u() should contain Channel2")
		} else {
			if len(channels2) != 1 {
				t.Errorf("ToM3u() Channel2 should have 1 entry, got %d", len(channels2))
			} else {
				channel := channels2[0]
				if channel.TvgName != "Channel2" || channel.Group != "Group1" || channel.Url != "url2" || channel.Title != "Channel2" {
					t.Errorf("ToM3u() Channel2 has wrong data: %+v", channel)
				}
			}
		}
	})

	// Test case 2: Empty input
	t.Run("empty input", func(t *testing.T) {
		txt := NewTvgNameChannels()
		result := ToM3u(txt)

		if result == nil {
			t.Fatal("ToM3u() should not return nil for empty input")
		}

		if len(result.TvgNameChannels) != 0 {
			t.Errorf("ToM3u() with empty input should return empty TvgNameChannels, got %d", len(result.TvgNameChannels))
		}
	})

	// Test case 3: Multiple channels with the same name
	t.Run("multiple channels with same name", func(t *testing.T) {
		txt := NewTvgNameChannels()
		txt["Channel1"] = []*Channel{
			{TvgName: "Channel1", Group: "Group1", Url: "url1"},
			{TvgName: "Channel1", Group: "Group2", Url: "url2"},
		}

		result := ToM3u(txt)

		channels, ok := result.TvgNameChannels["Channel1"]
		if !ok {
			t.Error("ToM3u() should contain Channel1")
		} else {
			if len(channels) != 2 {
				t.Errorf("ToM3u() Channel1 should have 2 entries, got %d", len(channels))
			}
		}
	})
}

func TestFromM3u(t *testing.T) {
	// Test case 1: Basic conversion functionality
	t.Run("basic conversion", func(t *testing.T) {
		// Create test data
		pls := m3u8x.NewProgramListSource()
		pls.TvgNameChannels["Channel1"] = []*m3u8x.Channel{
			{TvgName: "Channel1", Group: "Group1", Url: "url1", Title: "Title1", TvgLogo: "logo1"},
		}
		pls.TvgNameChannels["Channel2"] = []*m3u8x.Channel{
			{TvgName: "Channel2", Group: "Group1", Url: "url2", Title: "Title2", TvgLogo: "logo2"},
		}

		// Perform conversion
		result := FromM3u(pls)

		// Verify results
		if result == nil {
			t.Fatal("FromM3u() should not return nil")
		}

		if len(result) != 2 {
			t.Errorf("FromM3u() should have 2 channels, got %d", len(result))
		}

		// Check Channel1
		channels1, ok := result["Channel1"]
		if !ok {
			t.Error("FromM3u() should contain Channel1")
		} else {
			if len(channels1) != 1 {
				t.Errorf("FromM3u() Channel1 should have 1 entry, got %d", len(channels1))
			} else {
				channel := channels1[0]
				// Note: FromM3u function does not process Title and TvgLogo, only processes TvgName, Group, and Url
				if channel.TvgName != "Channel1" || channel.Group != "Group1" || channel.Url != "url1" {
					t.Errorf("FromM3u() Channel1 has wrong data: %+v", channel)
				}
			}
		}

		// Check Channel2
		channels2, ok := result["Channel2"]
		if !ok {
			t.Error("FromM3u() should contain Channel2")
		} else {
			if len(channels2) != 1 {
				t.Errorf("FromM3u() Channel2 should have 1 entry, got %d", len(channels2))
			} else {
				channel := channels2[0]
				if channel.TvgName != "Channel2" || channel.Group != "Group1" || channel.Url != "url2" {
					t.Errorf("FromM3u() Channel2 has wrong data: %+v", channel)
				}
			}
		}
	})

	// Test case 2: Empty input
	t.Run("empty input", func(t *testing.T) {
		pls := m3u8x.NewProgramListSource()
		result := FromM3u(pls)

		if result == nil {
			t.Fatal("FromM3u() should not return nil for empty input")
		}

		if len(result) != 0 {
			t.Errorf("FromM3u() with empty input should return empty TvgNameChannels, got %d", len(result))
		}
	})

	// Test case 3: Multiple channels with the same name
	t.Run("multiple channels with same name", func(t *testing.T) {
		pls := m3u8x.NewProgramListSource()
		pls.TvgNameChannels["Channel1"] = []*m3u8x.Channel{
			{TvgName: "Channel1", Group: "Group1", Url: "url1", Title: "Title1", TvgLogo: "logo1"},
			{TvgName: "Channel1", Group: "Group2", Url: "url2", Title: "Title2", TvgLogo: "logo2"},
		}

		result := FromM3u(pls)

		channels, ok := result["Channel1"]
		if !ok {
			t.Error("FromM3u() should contain Channel1")
		} else {
			if len(channels) != 2 {
				t.Errorf("FromM3u() Channel1 should have 2 entries, got %d", len(channels))
			}
		}
	})
}

func TestOutputTvgNameChannelsToTxtBz(t *testing.T) {
	// Test basic functionality
	t.Run("basic output", func(t *testing.T) {
		// Create test data
		txt := NewTvgNameChannels()
		txt["Channel1"] = []*Channel{
			{TvgName: "Channel1", Group: "Group1", Url: "url1"},
		}
		txt["Channel2"] = []*Channel{
			{TvgName: "Channel2", Group: "Group2", Url: "url2"},
		}

		// Create group list
		groupList := []*proto.GroupList{
			{
				Group:   "Group1",
				TvgName: []string{"Channel1"},
			},
			{
				Group:   "Group2",
				TvgName: []string{"Channel2"},
			},
		}

		// Execute function
		result := OutputTvgNameChannelsToTxtBz(txt, groupList)

		// Verify result is not empty
		if len(result) == 0 {
			t.Error("OutputTvgNameChannelsToTxtBz() should return non-empty byte slice")
		}

		// Convert to string for checking
		resultStr := string(result)

		// Check if necessary elements are included
		if resultStr != "" {
			// Check if update time line is included
			if strings.Index(resultStr, "更新时间,#genre#") == -1 {
				t.Error("OutputTvgNameChannelsToTxtBz() should contain update time header")
			}

			// Check if group titles are included
			if strings.Index(resultStr, "Group1,#genre#") == -1 ||
				strings.Index(resultStr, "Group2,#genre#") == -1 {
				t.Error("OutputTvgNameChannelsToTxtBz() should contain group headers")
			}

			// Check if channel data is included
			if strings.Index(resultStr, "Channel1,url1") == -1 ||
				strings.Index(resultStr, "Channel2,url2") == -1 {
				t.Error("OutputTvgNameChannelsToTxtBz() should contain channel data")
			}
		}
	})

	// Test empty input
	t.Run("empty inputs", func(t *testing.T) {
		txt := NewTvgNameChannels()
		var groupList []*proto.GroupList

		result := OutputTvgNameChannelsToTxtBz(txt, groupList)

		if len(result) == 0 {
			t.Error("OutputTvgNameChannelsToTxtBz() should return non-empty byte slice even with empty inputs")
		}

		resultStr := string(result)
		if strings.Index(resultStr, "更新时间,#genre#") == -1 {
			t.Error("OutputTvgNameChannelsToTxtBz() should contain update time header even with empty inputs")
		}
	})

	// Test partial matching
	t.Run("partial matching", func(t *testing.T) {
		// Only some channels are in the group list
		txt := NewTvgNameChannels()
		txt["Channel1"] = []*Channel{
			{TvgName: "Channel1", Group: "Group1", Url: "url1"},
		}
		txt["Channel2"] = []*Channel{
			{TvgName: "Channel2", Group: "Group2", Url: "url2"},
		}

		// Group list containing only Channel1
		groupList := []*proto.GroupList{
			{
				Group:   "Group1",
				TvgName: []string{"Channel1"},
			},
		}

		result := OutputTvgNameChannelsToTxtBz(txt, groupList)
		resultStr := string(result)

		// Should contain Channel1 but not Channel2
		if strings.Index(resultStr, "Channel1,url1") == -1 {
			t.Error("OutputTvgNameChannelsToTxtBz() should contain Channel1 data")
		}

		// Should not contain Channel2 data since it's not in groupList
		if strings.Index(resultStr, "Channel2,url2") != -1 {
			t.Error("OutputTvgNameChannelsToTxtBz() should not contain Channel2 data as it's not in groupList")
		}
	})
}
