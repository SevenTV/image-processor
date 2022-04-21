#ifndef IMAGE_PROCESSOR_CONVERT_PNG_HPP
#define IMAGE_PROCESSOR_CONVERT_PNG_HPP

#include <filesystem>
#include <map>
#include <opencv2/opencv.hpp>
#include <string>

enum OutputType {
    GIF = 1,
    WEBP = 2,
    AVIF = 3,
};

struct File {
    int delay;
    cv::Mat data;
};

struct Output {
    OutputType type;
    std::filesystem::path path;
};

#define ERROR_STR_MAX_LENGTH 100

typedef struct {
    int x_offset_, y_offset_, width_, height_;
} FrameRectangle;

typedef struct {
    WebPMuxFrameInfo sub_frame_; // Encoded frame rectangle.
    WebPMuxFrameInfo key_frame_; // Encoded frame if it is a key-frame.
    int is_key_frame_; // True if 'key_frame' has been chosen.
} EncodedFrame;

struct WebPAnimEncoder {
    const int canvas_width_; // Canvas width.
    const int canvas_height_; // Canvas height.
    const WebPAnimEncoderOptions options_; // Global encoding options.

    FrameRectangle prev_rect_; // Previous WebP frame rectangle.
    WebPConfig last_config_; // Cached in case a re-encode is needed.
    WebPConfig last_config_reversed_; // If 'last_config_' uses lossless, then
        // this config uses lossy and vice versa;
        // only valid if 'options_.allow_mixed'
        // is true.

    WebPPicture* curr_canvas_; // Only pointer; we don't own memory.

    // Canvas buffers.
    WebPPicture curr_canvas_copy_; // Possibly modified current canvas.
    int curr_canvas_copy_modified_; // True if pixels in 'curr_canvas_copy_'
        // differ from those in 'curr_canvas_'.

    WebPPicture prev_canvas_; // Previous canvas.
    WebPPicture prev_canvas_disposed_; // Previous canvas disposed to background.

    // Encoded data.
    EncodedFrame* encoded_frames_; // Array of encoded frames.
    size_t size_; // Number of allocated frames.
    size_t start_; // Frame start index.
    size_t count_; // Number of valid frames.
    size_t flush_count_; // If >0, 'flush_count' frames starting from
        // 'start' are ready to be added to mux.

    // key-frame related.
    int64_t best_delta_; // min(canvas size - frame size) over the frames.
        // Can be negative in certain cases due to
        // transparent pixels in a frame.
    int keyframe_; // Index of selected key-frame relative to 'start_'.
    int count_since_key_frame_; // Frames seen since the last key-frame.

    int first_timestamp_; // Timestamp of the first frame.
    int prev_timestamp_; // Timestamp of the last added frame.
    int prev_candidate_undecided_; // True if it's not yet decided if previous
        // frame would be a sub-frame or a key-frame.

    // Misc.
    int is_first_frame_; // True if first frame is yet to be added/being added.
    int got_null_frame_; // True if WebPAnimEncoderAdd() has already been called
        // with a NULL frame.

    size_t in_frame_count_; // Number of input frames processed so far.
    size_t out_frame_count_; // Number of frames added to mux so far. This may be
        // different from 'in_frame_count_' due to merging.

    WebPMux* mux_; // Muxer to assemble the WebP bitstream.
    char error_str_[ERROR_STR_MAX_LENGTH]; // Error string. Empty if no error.
};

#endif
