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

#endif
