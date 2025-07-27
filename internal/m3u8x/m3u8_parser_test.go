package m3u8x

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var programListSource = []byte(`#EXTM3U x-tvg-url="https://gh.catmak.name/https://raw.githubusercontent.com/Guovin/iptv-api/refs/heads/master/output/epg/epg.gz"
#EXTINF:-1 tvg-name="CCTV1" tvg-logo="https://gh.catmak.name/https://raw.githubusercontent.com/fanmingming/live/main/tv/CCTV1.png" group-title="🕘️更新时间",2025-07-09 09:06:36
http://php.jdshipin.com/TVOD/iptv.php?id=rthk33
#EXTINF:-1 tvg-name="CCTV1" tvg-logo="https://gh.catmak.name/https://raw.githubusercontent.com/fanmingming/live/main/tv/CCTV1.png" group-title="📺央视频道",CCTV-1
http://php.jdshipin.com/TVOD/iptv.php?id=rthk33
#EXTINF:-1 tvg-name="CCTV1" tvg-logo="https://gh.catmak.name/https://raw.githubusercontent.com/fanmingming/live/main/tv/CCTV1.png" group-title="📺央视频道",CCTV-1
http://iptv.huuc.edu.cn/hls/cctv1hd.m3u8`)

var baseUrl = "http://iptv.huuc.edu.cn/hls/"
var liveStreamSource = []byte(`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:368
#EXT-X-TARGETDURATION:5
#EXTINF:5.000,
cctv1hd-368.ts
#EXTINF:5.000,
cctv1hd-369.ts
#EXTINF:5.000,
cctv1hd-370.ts
#EXTINF:5.000,
cctv1hd-371.ts
#EXTINF:5.000,
cctv1hd-372.ts
#EXTINF:5.000,
cctv1hd-373.ts
`)

func TestParseLiveStreamSource(t *testing.T) {
	source := NewLiveStreamSource(baseUrl)
	err := source.ParseLiveStreamSource(liveStreamSource)
	require.NoError(t, err, "ParseLiveStreamSource failed")
}

func TestSource_ParseProgramListSource(t *testing.T) {
	source := NewProgramListSource()
	err := source.ParseProgramListSource(programListSource)
	require.NoError(t, err, "ParseProgramListSource failed")
}
