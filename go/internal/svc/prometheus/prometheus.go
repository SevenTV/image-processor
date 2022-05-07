package prometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/seventv/image-processor/go/internal/instance"
)

type Options struct {
	Labels prometheus.Labels
}

func New(o Options) instance.Prometheus {
	m := &Instance{
		totalTasks: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   "image_processor",
			Name:        "total_tasks",
			Help:        "The total number of successful tasks",
			ConstLabels: o.Labels,
		}, []string{"state"}),
		currentTasks: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "image_processor",
			Name:        "current_tasks",
			Help:        "The current number of request",
			ConstLabels: o.Labels,
		}),
		taskDurationSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   "image_processor",
			Name:        "task_duration_seconds",
			Help:        "The seconds spent running tasks",
			ConstLabels: o.Labels,
		}, []string{"action"}),
		totalFramesProcessed: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "image_processor",
			Name:        "total_frames",
			Help:        "The total number of frames processed",
			ConstLabels: o.Labels,
		}),
		totalBytes: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   "image_processor",
			Name:        "total_bytes",
			Help:        "Total bytes moved from s3",
			ConstLabels: o.Labels,
		}, []string{"direction"}),
		totalFiles: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   "image_processor",
			Name:        "total_files",
			Help:        "Total files moved from s3",
			ConstLabels: o.Labels,
		}, []string{"direction"}),
		taskInputType: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   "image_processor",
			Name:        "input_type",
			Help:        "The mime type of input images",
			ConstLabels: o.Labels,
		}, []string{"content_type"}),
		registry: prometheus.NewRegistry(),
	}

	m.registry.MustRegister(
		m.currentTasks,
		m.taskDurationSeconds,
		m.totalTasks,
		m.taskInputType,

		m.totalBytes,
		m.totalFiles,

		m.totalFramesProcessed,
	)

	return m
}

type Instance struct {
	totalTasks          *prometheus.CounterVec
	currentTasks        prometheus.Gauge
	taskDurationSeconds *prometheus.HistogramVec

	totalBytes           *prometheus.CounterVec
	totalFiles           *prometheus.CounterVec
	totalFramesProcessed prometheus.Counter

	taskInputType *prometheus.CounterVec

	registry *prometheus.Registry
}

func (m *Instance) Registry() *prometheus.Registry {
	return m.registry
}

func (m *Instance) StartTask() func(success bool) {
	start := time.Now()
	m.currentTasks.Inc()

	return func(success bool) {
		if success {
			m.totalTasks.With(prometheus.Labels{"state": "success"}).Inc()
		} else {
			m.totalTasks.With(prometheus.Labels{"state": "failed"}).Inc()
		}
		m.currentTasks.Dec()
		m.taskDurationSeconds.With(prometheus.Labels{"action": "complete"}).Observe(float64(time.Since(start)/time.Millisecond) / 1000)
	}
}

func (m *Instance) TotalBytesDownloaded(bytes int) {
	m.totalFiles.With(prometheus.Labels{"direction": "downloaded"}).Inc()
	m.totalBytes.With(prometheus.Labels{"direction": "downloaded"}).Add(float64(bytes))
}

func (m *Instance) TotalBytesUploaded(bytes int) {
	m.totalFiles.With(prometheus.Labels{"direction": "uploaded"}).Inc()
	m.totalBytes.With(prometheus.Labels{"direction": "uploaded"}).Add(float64(bytes))
}

func (m *Instance) TotalFramesProcessed(frames int) {
	m.totalFramesProcessed.Add(float64(frames))
}

func (m *Instance) DownloadFile() func() {
	start := time.Now()

	return func() {
		m.taskDurationSeconds.With(prometheus.Labels{"action": "download_file"}).Observe(float64(time.Since(start)/time.Millisecond) / 1000)
	}
}

func (m *Instance) ResizeFrames() func() {
	start := time.Now()

	return func() {
		m.taskDurationSeconds.With(prometheus.Labels{"action": "resize_frames"}).Observe(float64(time.Since(start)/time.Millisecond) / 1000)
	}
}

func (m *Instance) ExportFrames() func() {
	start := time.Now()

	return func() {
		m.taskDurationSeconds.With(prometheus.Labels{"action": "export_frames"}).Observe(float64(time.Since(start)/time.Millisecond) / 1000)
	}
}

func (m *Instance) MakeResults() func() {
	start := time.Now()

	return func() {
		m.taskDurationSeconds.With(prometheus.Labels{"action": "make_results"}).Observe(float64(time.Since(start)/time.Millisecond) / 1000)
	}
}

func (m *Instance) UploadResults() func() {
	start := time.Now()

	return func() {
		m.taskDurationSeconds.With(prometheus.Labels{"action": "upload_results"}).Observe(float64(time.Since(start)/time.Millisecond) / 1000)
	}
}

func (m *Instance) InputFileType(contentType string) {
	m.taskInputType.With(prometheus.Labels{"content_type": contentType}).Inc()
}
