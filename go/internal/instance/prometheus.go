package instance

import "github.com/prometheus/client_golang/prometheus"

type Prometheus interface {
	StartTask() func(success bool)

	DownloadFile() func()
	ExportFrames() func()
	ResizeFrames() func()
	MakeResults() func()
	UploadResults() func()

	TotalFramesProcessed(int)
	TotalBytesDownloaded(int)
	TotalBytesUploaded(int)

	InputFileType(string)

	Register(prometheus.Registerer)
}
