#ifndef IMAGE_PROCESSOR_CONVERT_PNG_HPP
#define IMAGE_PROCESSOR_CONVERT_PNG_HPP

#include <filesystem>
#include <map>
#include <opencv2/opencv.hpp>
#include <string>

struct File {
    bool used;
    cv::Mat data;
};

struct Output {
    int width;
    int height;
    int resizeRatio;
    File input;
    std::filesystem::path path;
};

#endif
