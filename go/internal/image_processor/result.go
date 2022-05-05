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
	ID      string         `json:"id"`
	State   ResultState    `json:"state"`
	Message string         `json:"message"`
	Outputs []ResultOutput `json:"outputs"`
}

type ResultOutputFormatType int32

const (
	_ ResultOutputFormatType = iota
	ResultOutputFormatTypeWEBP
	ResultOutputFormatTypeAVIF
	ResultOutputFormatTypeGIF
	ResultOutputFormatTypePNG
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

type ResultOutput struct {
	Name       string                 `json:"name"`
	FrameCount int                    `json:"frame_count"`
	Format     ResultOutputFormatType `json:"format"`
	Width      int                    `json:"width"`
	Height     int                    `json:"height"`
	Key        string                 `json:"key"`
	Bucket     string                 `json:"bucket"`
	Size       int                    `json:"size"`
}
