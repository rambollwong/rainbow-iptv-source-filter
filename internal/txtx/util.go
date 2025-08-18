package txtx

import (
	"strings"
	"time"

	"github.com/rambollwong/rainbow-iptv-source-filter/internal/m3u8x"
	"github.com/rambollwong/rainbow-iptv-source-filter/pkg/proto"
)

func ToM3u(txt TvgNameChannels) *m3u8x.ProgramListSource {
	pls := m3u8x.NewProgramListSource()
	for tvgName, channels := range txt {
		ch, ok := pls.TvgNameChannels[tvgName]
		if !ok {
			ch = make([]*m3u8x.Channel, 0, 16)
		}

		for _, c := range channels {
			ch = append(ch, &m3u8x.Channel{
				TvgName: c.TvgName,
				TvgLogo: "",
				Group:   c.Group,
				Title:   c.TvgName,
				Url:     c.Url,
			})
		}

		pls.TvgNameChannels[tvgName] = ch
	}
	return pls
}

func FromM3u(pls *m3u8x.ProgramListSource) TvgNameChannels {
	txt := NewTvgNameChannels()
	for tvgName, channels := range pls.TvgNameChannels {
		ch, ok := txt[tvgName]
		if !ok {
			ch = make([]*Channel, 0, 16)
		}
		for _, c := range channels {
			ch = append(ch, &Channel{
				TvgName: c.TvgName,
				Group:   c.Group,
				Url:     c.Url,
			})
		}
		txt[tvgName] = ch
	}
	return txt
}

func OutputTvgNameChannelsToTxtBz(txt TvgNameChannels, groupList []*proto.GroupList) []byte {
	b := strings.Builder{}
	b.WriteString("更新时间,")
	b.WriteString(TxtGenre)
	b.WriteString("\n")
	b.WriteString(time.Now().Format("20060102 15:04:05"))
	b.WriteString(",http://tc-tct.douyucdn2.cn/dyliveflv1/122402rK7MO9bXSq_2000.flv?wsAuth=8cea39337984fd3341cc9ec569502e4f&token=cpn-androidmpro-0-122402-0fcea45d2300cfa0ac75fafd8679bb53af10de8c33ae99d9&logo=0&expire=0&did=d010b07dcb997ada9934081c873542f0&origin=tct&vhost=p\n")
	b.WriteString("\n\n")

	for _, list := range groupList {
		group := list.Group
		b.WriteString(group)
		b.WriteString(",")
		b.WriteString(TxtGenre)
		b.WriteString("\n")

		for _, tvgName := range list.TvgName {
			channels, ok := txt[tvgName]
			if !ok {
				continue
			}
			for _, channel := range channels {
				b.WriteString(channel.TvgName)
				b.WriteString(",")
				b.WriteString(channel.Url)
				b.WriteString("\n")
			}
		}
	}

	return []byte(b.String())
}
