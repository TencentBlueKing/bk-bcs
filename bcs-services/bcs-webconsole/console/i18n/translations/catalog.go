// Code generated by running "go generate" in golang.org/x/text. DO NOT EDIT.

package translations

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

type dictionary struct {
	index []uint32
	data  string
}

func (d *dictionary) Lookup(key string) (data string, ok bool) {
	p, ok := messageKeyToIndex[key]
	if !ok {
		return "", false
	}
	start, end := d.index[p], d.index[p+1]
	if start == end {
		return "", false
	}
	return d.data[start:end], true
}

func init() {
	dict := map[string]catalog.Dictionary{
		"en":      &dictionary{index: enIndex, data: enData},
		"zh_Hans": &dictionary{index: zh_HansIndex, data: zh_HansData},
	}
	fallback := language.MustParse("zh-Hans")
	cat, err := catalog.NewFromMap(dict, catalog.Fallback(fallback))
	if err != nil {
		panic(err)
	}
	message.DefaultCatalog = cat
}

var messageKeyToIndex = map[string]int{
	"获取session失败: %s": 2,
	"获取session成功":     3,
	"获取集群成功":          1,
	"请求参数错误: %s":      4,
	"项目不正确":           0,
}

var enIndex = []uint32{ // 6 elements
	0x00000000, 0x00000015, 0x0000002d, 0x00000047,
	0x0000005e, 0x0000005e,
} // Size: 48 bytes

const enData string = "" + // Size: 94 bytes
	"\x02Project_id Incorrect\x02Get Clusters successful\x02Get session faile" +
	"d: %[1]s\x02Get session successful"

var zh_HansIndex = []uint32{ // 6 elements
	0x00000000, 0x00000010, 0x00000023, 0x0000003e,
	0x00000052, 0x0000006c,
} // Size: 48 bytes

const zh_HansData string = "" + // Size: 108 bytes
	"\x02项目不正确\x02获取集群成功\x02获取session失败: %[1]s\x02获取session成功\x02请求参数错误: %[1]" +
	"s"

	// Total table size 298 bytes (0KiB); checksum: 2C901CFD