package explorer

type MediaItem struct {
	Id         string   `json:"id"`
	Name       string   `json:"name"`
	Mode       string   `json:"mode"`
	Path       string   `json:"path"`
	HlsPath    string   `json:"hlsPath"`
	Suffix     string   `json:"type"`
	UpdateTime int64    `json:"updateTime"`
	Size       int64    `json:"size"`
}
