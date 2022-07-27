package task

import (
	"encoding/json"
	"time"
)

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

type ResizeRatio int32

const (
	ResizeRatioNothing ResizeRatio = iota
	ResizeRatioStretch
	ResizeRatioPaddingRightBottom
	ResizeRatioPaddingLeftBottom
	ResizeRatioPaddingRightTop
	ResizeRatioPaddingLeftTop
	ResizeRatioPaddingCenter
)

type Task struct {
	ID                string          `json:"id"`
	Flags             TaskFlag        `json:"flags"`
	Input             TaskInput       `json:"input"`
	Output            TaskOutput      `json:"output"`
	SmallestMaxWidth  int             `json:"smallest_max_width"`  // 96
	SmallestMaxHeight int             `json:"smallest_max_height"` // 32
	ResizeRatio       ResizeRatio     `json:"resize_ratio"`
	Scales            []int           `json:"scales"` // 1, 2, 3, 4 for 1x, 2x, 3x, 4x
	Limits            TaskLimits      `json:"limits"`
	Metadata          json.RawMessage `json:"metadata"`
}

type TaskLimits struct {
	MaxProcessingTime time.Duration `json:"max_processing_time"`
	MaxFrameCount     int           `json:"max_frame_count"`
	MaxWidth          int           `json:"max_width"`
	MaxHeight         int           `json:"max_height"`
}

type TaskInput struct {
	Bucket   string            `json:"bucket"`
	Key      string            `json:"key"`
	Reupload TaskInputReupload `json:"reupload"`
}

type TaskInputReupload struct {
	Enabled      bool   `json:"enabled"`
	Key          string `json:"key"`
	Bucket       string `json:"bucket"`
	ACL          string `json:"acl"`
	CacheControl string `json:"cache_control"`
}

type TaskOutput struct {
	Prefix               string `json:"prefix"`
	ACL                  string `json:"acl"`
	Bucket               string `json:"bucket"`
	CacheControl         string `json:"cache_control"`
	ExcludeFileExtension bool   `json:"exclude_file_extension"` // Temporary compatibility workaround, this omits the file extension for WEBP
}
