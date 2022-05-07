package instance

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Prometheus interface {
	Register(r prometheus.Registerer)

	StartTask() func(success bool)

	DownloadFile() func()
	ExportFrames() func()
	ResizeFrames() func()
	MakeResults() func()
	UploadResults() func()

	TotalFramesProcessed(int)
	TotalBytesDownloaded(int)
	TotalBytesUploaded(int)
}
