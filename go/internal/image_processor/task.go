package image_processor

type TaskFlag int32

const (
	TaskFlagGIF TaskFlag = 1 << iota
	TaskFlagWEBP
	TaskFlagAVIF
	TaskFlagPNG_STATIC
	TaskFlagWEBP_STATIC
	TaskFlagAVIF_STATIC
)

type Task struct {
	ID    string   `json:"id"`
	Flags TaskFlag `json:"flags"`
	Input struct {
		Bucket string `json:"bucket"`
		Key    string `json:"key"`
	} `json:"input"`
	Output struct {
		Bucket string `json:"bucket"`
		Prefix string `json:"prefix"`
	} `json:"output"`
}
