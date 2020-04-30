package gokong

func String(v string) *string {
	return &v
}

func Bool(v bool) *bool {
	return &v
}

func Int(v int) *int {
	return &v
}

func ToId(v string) *Id {
	id := Id(v)
	return &id
}

func IdToString(v *Id) string {
	if v == nil {
		return ""
	}
	return string(*v)
}

func IpPortSliceSlice(src []IpPort) []*IpPort {
	dst := make([]*IpPort, len(src))
	for i := 0; i < len(src); i++ {
		dst[i] = &(src[i])
	}
	return dst
}

func StringSlice(src []string) []*string {
	dst := make([]*string, len(src))
	for i := 0; i < len(src); i++ {
		dst[i] = &(src[i])
	}
	return dst
}

func StringValueSlice(src []*string) []string {
	dst := make([]string, len(src))
	for i := 0; i < len(src); i++ {
		if src[i] != nil {
			dst[i] = *(src[i])
		}
	}
	return dst
}
