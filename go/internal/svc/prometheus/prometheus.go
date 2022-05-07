package prometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/seventv/image-processor/go/internal/instance"
)

type Options struct {
	Labels prometheus.Labels
}

func copyLabels(p prometheus.Labels) prometheus.Labels {
	x := prometheus.Labels{}
	for k, v := range p {
		x[k] = v
	}

	return x
}

func New(o Options) instance.Prometheus {
	totalSuccessfulTasks := copyLabels(o.Labels)
	totalFailedTasks := copyLabels(o.Labels)
	currentTasks := copyLabels(o.Labels)
	taskDurationSeconds := copyLabels(o.Labels)
	totalBytesDownloaded := copyLabels(o.Labels)
	totalBytesUploaded := copyLabels(o.Labels)
	totalFramesProcessed := copyLabels(o.Labels)
	downloadFileDuration := copyLabels(o.Labels)
	resizeFramesDuration := copyLabels(o.Labels)
	exportFramesDuration := copyLabels(o.Labels)
	makeResultsDuration := copyLabels(o.Labels)
	uploadResultsDuration := copyLabels(o.Labels)

	totalSuccessfulTasks["state"] = "successful"
	totalFailedTasks["state"] = "failed"

	totalBytesDownloaded["state"] = "downloaded"
	totalBytesUploaded["state"] = "uploaded"

	return &Instance{
		totalSuccessfulTasks: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "image_processor",
			Name:        "total_tasks",
			Help:        "The total number of successful tasks",
			ConstLabels: totalSuccessfulTasks,
		}),
		totalFailedTasks: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "image_processor",
			Name:        "total_tasks",
			Help:        "The total number of failed tasks",
			ConstLabels: totalFailedTasks,
		}),
		currentTasks: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "image_processor",
			Name:        "current_tasks",
			Help:        "The current number of request",
			ConstLabels: currentTasks,
		}),
		taskDurationSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   "image_processor",
			Name:        "task_duration_seconds",
			Help:        "The seconds spent running tasks",
			ConstLabels: taskDurationSeconds,
		}),
		downloadFileDurationSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   "image_processor",
			Name:        "download_file_duration_seconds",
			Help:        "The seconds spent downloading files",
			ConstLabels: downloadFileDuration,
		}),
		exportFramesDurationSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   "image_processor",
			Name:        "export_frames_duration_seconds",
			Help:        "The seconds spent exporting frames",
			ConstLabels: exportFramesDuration,
		}),
		resizeFramesDurationSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   "image_processor",
			Name:        "resize_frames_duration_seconds",
			Help:        "The seconds spent resizing frames",
			ConstLabels: resizeFramesDuration,
		}),
		makeResultsDurationSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   "image_processor",
			Name:        "make_results_duration_seconds",
			Help:        "The seconds spent downloading files",
			ConstLabels: makeResultsDuration,
		}),
		uploadResultsDurationSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   "image_processor",
			Name:        "upload_results_duration_seconds",
			Help:        "The seconds spent uploading results",
			ConstLabels: uploadResultsDuration,
		}),
		totalBytesDownloaded: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "image_processor",
			Name:        "total_bytes",
			Help:        "The total number of bytes downloaded",
			ConstLabels: totalBytesDownloaded,
		}),
		totalBytesUploaded: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "image_processor",
			Name:        "total_bytes",
			Help:        "The total number of bytes uploaded",
			ConstLabels: totalBytesUploaded,
		}),
		totalFramesProcessed: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "image_processor",
			Name:        "total_frames",
			Help:        "The total number of frames processed",
			ConstLabels: totalFramesProcessed,
		}),
	}
}

type Instance struct {
	totalSuccessfulTasks prometheus.Counter
	totalFailedTasks     prometheus.Counter
	currentTasks         prometheus.Gauge
	taskDurationSeconds  prometheus.Histogram

	downloadFileDurationSeconds  prometheus.Histogram
	resizeFramesDurationSeconds  prometheus.Histogram
	exportFramesDurationSeconds  prometheus.Histogram
	makeResultsDurationSeconds   prometheus.Histogram
	uploadResultsDurationSeconds prometheus.Histogram

	totalBytesDownloaded prometheus.Counter
	totalBytesUploaded   prometheus.Counter
	totalFramesProcessed prometheus.Counter
}

func (m *Instance) Register(r prometheus.Registerer) {
	r.MustRegister(
		m.currentTasks,
		m.taskDurationSeconds,
		m.totalFailedTasks,
		m.totalSuccessfulTasks,

		m.downloadFileDurationSeconds,
		m.resizeFramesDurationSeconds,
		m.exportFramesDurationSeconds,
		m.makeResultsDurationSeconds,
		m.uploadResultsDurationSeconds,

		m.totalBytesDownloaded,
		m.totalBytesUploaded,
		m.totalFramesProcessed,
	)
}

func (m *Instance) StartTask() func(success bool) {
	start := time.Now()
	m.currentTasks.Inc()

	return func(success bool) {
		if success {
			m.totalSuccessfulTasks.Inc()
		} else {
			m.totalFailedTasks.Inc()
		}
		m.currentTasks.Dec()
		m.taskDurationSeconds.Observe(float64(time.Since(start)/time.Millisecond) / 1000)
	}
}

func (m *Instance) TotalBytesDownloaded(bytes int) {
	m.totalBytesDownloaded.Add(float64(bytes))
}

func (m *Instance) TotalBytesUploaded(bytes int) {
	m.totalBytesDownloaded.Add(float64(bytes))
}

func (m *Instance) TotalFramesProcessed(frames int) {
	m.totalFramesProcessed.Add(float64(frames))
}

func (m *Instance) DownloadFile() func() {
	start := time.Now()

	return func() {
		m.downloadFileDurationSeconds.Observe(float64(time.Since(start)/time.Millisecond) / 1000)
	}
}

func (m *Instance) ResizeFrames() func() {
	start := time.Now()

	return func() {
		m.resizeFramesDurationSeconds.Observe(float64(time.Since(start)/time.Millisecond) / 1000)
	}
}

func (m *Instance) ExportFrames() func() {
	start := time.Now()

	return func() {
		m.exportFramesDurationSeconds.Observe(float64(time.Since(start)/time.Millisecond) / 1000)
	}
}

func (m *Instance) MakeResults() func() {
	start := time.Now()

	return func() {
		m.makeResultsDurationSeconds.Observe(float64(time.Since(start)/time.Millisecond) / 1000)
	}
}

func (m *Instance) UploadResults() func() {
	start := time.Now()

	return func() {
		m.uploadResultsDurationSeconds.Observe(float64(time.Since(start)/time.Millisecond) / 1000)
	}
}
