package task

import (
	"encoding/json"
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
	ID            string          `json:"id"`
	StartedAt     time.Time       `json:"started_at"`
	FinishedAt    time.Time       `json:"finished_at"`
	State         ResultState     `json:"state"`
	Message       string          `json:"message"`
	ImageInput    ResultFile      `json:"image_input"`
	ImageOutputs  []ResultFile    `json:"image_outputs"`
	ArchiveOutput ResultFile      `json:"archive_output"`
	Metadata      json.RawMessage `json:"metadata"`
}

type ResultFile struct {
	Name         string `json:"name"`
	SHA3         string `json:"sha3"`
	ContentType  string `json:"content_type"`
	Size         int    `json:"size"`
	Key          string `json:"key"`
	Bucket       string `json:"bucket"`
	ACL          string `json:"acl"`
	CacheControl string `json:"cache_control"`

	FrameCount int `json:"frame_count,omitempty"`
	Width      int `json:"width,omitempty"`
	Height     int `json:"height,omitempty"`
}
