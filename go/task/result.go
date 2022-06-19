package task

import (
	"fmt"
	"time"
)

type ResultState int32

const (
	_ ResultState = iota
	ResultStateSuccess
	ResultStateFailed
)

func (r ResultState) String() string {
	switch r {
	case ResultStateSuccess:
		return "SUCCESS"
	case ResultStateFailed:
		return "FAILED"
	default:
		return fmt.Sprintf("UNKNOWN TYPE %d", r)
	}
}

type Result struct {
	ID           string          `json:"id"`
	StartedAt    time.Time       `json:"started_at"`
	FinishedAt   time.Time       `json:"finished_at"`
	State        ResultState     `json:"state"`
	Message      string          `json:"message"`
	ImageInput   ResultImage     `json:"image_input"`
	ImageOutputs []ResultImage   `json:"image_outputs"`
	ZipOutput    ResultZipOutput `json:"zip_output"`
}

type ResultImage struct {
	Name         string `json:"name"`
	SHA3         string `json:"sha3"`
	FrameCount   int    `json:"frame_count"`
	ContentType  string `json:"content_type"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Size         int    `json:"size"`
	Key          string `json:"key"`
	Bucket       string `json:"bucket"`
	ACL          string `json:"acl"`
	CacheControl string `json:"cache_control"`
}

type ResultZipOutput struct {
	Name         string `json:"name"`
	SHA3         string `json:"sha3"`
	Size         int    `json:"size"`
	Key          string `json:"key"`
	Bucket       string `json:"bucket"`
	ACL          string `json:"acl"`
	CacheControl string `json:"cache_control"`
}
