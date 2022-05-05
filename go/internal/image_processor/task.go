package image_processor

type TaskFlag int32

const (
	TaskFlagGIF TaskFlag = 1 << iota
	TaskFlagWEBP
	TaskFlagAVIF
	TaskFlagPNG
	TaskFlagPNG_STATIC
	TaskFlagWEBP_STATIC
	TaskFlagAVIF_STATIC
	TaskFlagALL TaskFlag = (1 << iota) - 1
)

type Task struct {
	ID                string     `json:"id"`
	Flags             TaskFlag   `json:"flags"`
	Input             TaskInput  `json:"input"`
	Output            TaskOutput `json:"output"`
	SmallestMaxWidth  int        `json:"max_width"`  // 128
	SmallestMaxHeight int        `json:"max_height"` // 32
	Scales            []int      `json:"scales"`     // 1, 2, 3, 4 for 1x, 2x, 3x, 4x
}

type TaskInput struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

type TaskOutput struct {
	Bucket string `json:"bucket"`
	Prefix string `json:"prefix"`
}
