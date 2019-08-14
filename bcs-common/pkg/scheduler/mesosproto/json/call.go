package json

type Call struct {
	Type      string     `json:"type"`
	Subscribe *Subscribe `json:"subscribe"`
}
type Subscribe struct {
	FrameworkInfo *FrameworkInfo `json:"framework_info"`
}
type FrameworkInfo struct {
	User         string `json:"user"`
	Name         string `json:"name"`
	HostName     string `json:"hostname"`
	*FrameworkID `json:"framework_id"`
}
type FrameworkID struct {
	Value string `json:"value"`
}
