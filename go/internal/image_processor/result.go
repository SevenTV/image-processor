package image_processor

import "fmt"

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
	ID           string              `json:"id"`
	State        ResultState         `json:"state"`
	Message      string              `json:"message"`
	InputSHA3    string              `json:"input_sha3"`
	ImageOutputs []ResultImageOutput `json:"image_outputs"`
	ZipOutput    ResultZipOutput     `json:"zip_output"`
}

type ResultOutputFormatType int32

const (
	_ ResultOutputFormatType = iota
	ResultOutputFormatTypeWEBP
	ResultOutputFormatTypeAVIF
	ResultOutputFormatTypeGIF
	ResultOutputFormatTypePNG
	ResultOutputFormatTypeZIP
)

func (r ResultOutputFormatType) String() string {
	switch r {
	case ResultOutputFormatTypeWEBP:
		return "WEBP"
	case ResultOutputFormatTypeAVIF:
		return "AVIF"
	case ResultOutputFormatTypeGIF:
		return "GIF"
	case ResultOutputFormatTypePNG:
		return "PNG"
	default:
		return fmt.Sprintf("UNKNOWN TYPE %d", r)
	}
}

type ResultImageOutput struct {
	Name         string                 `json:"name"`
	SHA3         string                 `json:"sha3"`
	FrameCount   int                    `json:"frame_count"`
	Format       ResultOutputFormatType `json:"format"`
	ContentType  string                 `json:"content_type"`
	Width        int                    `json:"width"`
	Height       int                    `json:"height"`
	Key          string                 `json:"key"`
	Bucket       string                 `json:"bucket"`
	Size         int                    `json:"size"`
	ACL          string                 `json:"acl"`
	CacheControl string                 `json:"cache_control"`
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
