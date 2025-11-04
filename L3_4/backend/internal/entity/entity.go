package entity

type ImageTask struct {
	ImageID    string      `json:"image_id"`
	InputPath  string      `json:"input_path"`
	OutputPath string      `json:"output_path"`
	Operations *Operations `json:"operations"`
}

type Operations struct {
	Resize    *Resize    `json:"resize"`
	WaterMark *Watermark `json:"watermark"`
}

type Resize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Watermark struct {
	Text string `json:"text,omitempty"`
}

type ImageStatus struct {
	Status     string `json:"status"`
	OutputPath string `json:"output_path,omitempty"`
}
