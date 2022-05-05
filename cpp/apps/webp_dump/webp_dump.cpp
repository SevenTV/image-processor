#include <filesystem>
#include <fstream>
#include <iostream>
#include <opencv2/opencv.hpp>
#include <string>
#include <vector>
#include <webp/demux.h>

#define NEXTARG()                                                     \
    if (((argIndex + 1) == argc) || (argv[argIndex + 1][0] == '-')) { \
        std::cerr << arg << " requires an argument." << std::endl;    \
        return EXIT_FAILURE;                                          \
    }                                                                 \
    arg = std::string(argv[++argIndex])

void syntax()
{
    std::cerr << "Syntax: webp_dump -i input.webp -o output"
              << std::endl
              << "Options:" << std::endl
              << "  -h,--help                   : Shows syntax help" << std::endl
              << "  -i,--input FILENAME         : Input file location (supported "
                 "types are webp)."
              << std::endl
              << "  -o,--output FOLDER          : Output folder"
              << std::endl
              << std::endl;
}

int ReadFile(const std::string fileName, const uint8_t** data, size_t* size)
{
    int ok;

    if (data == NULL || size == NULL) {
        std::cout << "x" << std::endl;
        return 0;
    }
    *data = NULL;
    *size = 0;

    auto in = std::fopen(fileName.c_str(), "rb");
    if (in == NULL) {
        return 0;
    }

    fseek(in, 0, SEEK_END);
    size_t file_size = ftell(in);
    fseek(in, 0, SEEK_SET);

    auto file_data = (uint8_t*)WebPMalloc(file_size + 1);
    if (file_data == NULL) {
        fclose(in);
        return 0;
    }

    ok = (fread(file_data, file_size, 1, in) == 1);
    fclose(in);

    if (!ok) {
        WebPFree(file_data);
        return 0;
    }

    file_data[file_size] = '\0';
    *data = file_data;
    *size = file_size;

    return 1;
}

int main(int argc, char* argv[])
{
    std::string input;
    std::string output;

    int argIndex = 1;
    while (argIndex < argc) {
        std::string arg = argv[argIndex];

        if (arg == "--help" || arg == "-h") {
            syntax();
            return EXIT_FAILURE;
        } else if (arg == "--output" || arg == "-o") {
            NEXTARG();

            output = arg;
        } else if (arg == "--input" || arg == "-i") {
            NEXTARG();
            if (std::filesystem::path(arg).extension() != ".webp") {
                std::cerr << "\"" << arg
                          << "\" is an unsupported file type for an input image."
                          << std::endl;
                return EXIT_FAILURE;
            }

            input = arg;
        } else {
            std::cerr << "\"" << arg << "\" is an unknown argument." << std::endl;
            return EXIT_FAILURE;
        }

    loop:
        argIndex++;
    }

    WebPData webpData;
    if (!ReadFile(input, &webpData.bytes, &webpData.size)) {
        std::cerr << "\"" << input << "\" failed to read input file." << std::endl;
        return EXIT_FAILURE;
    }

    uint32_t frame_index = 0;
    WebPAnimInfo animInfo;

    auto dec = WebPAnimDecoderNew(&webpData, NULL);
    if (!dec) {
        std::cerr << "\"" << input << "\" failed to decode file." << std::endl;
        return EXIT_FAILURE;
    }

    if (!WebPAnimDecoderGetInfo(dec, &animInfo)) {
        std::cerr << "\"" << input << "\" failed to get info file." << std::endl;
        return EXIT_FAILURE;
    }

    auto prevFrameTimestamp = 0;
    cv::Mat frame(animInfo.canvas_height, animInfo.canvas_width, CV_8UC4);

    auto frameIndex = 0;
    char buffer[10];
    std::cout << "frame_count: " << animInfo.frame_count << " width: " << animInfo.canvas_width << " height: " << animInfo.canvas_height << std::endl;
    std::cout << "idx,delay" << std::endl;
    while (WebPAnimDecoderHasMoreFrames(dec)) {
        int timestamp;

        if (!WebPAnimDecoderGetNext(dec, &frame.data, &timestamp)) {
            std::cerr << "\"" << input << "\" failed to decode frame #" << frameIndex << std::endl;
            return EXIT_FAILURE;
        }

        if (frameIndex > animInfo.frame_count) {
            std::cerr << "\"" << input << "\" out of bouns read detected: " << frameIndex << std::endl;
            return EXIT_FAILURE;
        }

        auto duration = timestamp - prevFrameTimestamp;
        std::cout << frameIndex << "," << duration / 10 << std::endl;

        sprintf(buffer, "%04d.png", frameIndex);

        auto filename = std::filesystem::path(output) / buffer;

        cv::cvtColor(frame, frame, cv::COLOR_RGBA2BGRA);
        cv::imwrite(filename, frame);

        ++frameIndex;
        prevFrameTimestamp = timestamp;
    }

    frame.release();
    WebPAnimDecoderDelete(dec);

    return EXIT_SUCCESS;
}
