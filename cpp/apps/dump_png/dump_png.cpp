#include <avif/avif.h>
#include <filesystem>
#include <fstream>
#include <iostream>
#include <opencv2/opencv.hpp>
#include <string>
#include <thread>
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
    std::cerr << "Syntax: dump_png -i input.webp -o output"
              << std::endl
              << "Options:" << std::endl
              << "  -h,--help                   : Shows syntax help" << std::endl
              << "  -i,--input FILENAME         : Input file location (supported "
                 "types are webp and avif)."
              << std::endl
              << "  -o,--output FOLDER          : Output folder"
              << std::endl
              << std::endl;
}

int ReadFile(const std::string fileName, const uint8_t** data, size_t* size)
{
    int ok;

    if (data == NULL || size == NULL) {
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

    bool isAvif, isWebp;

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
            isWebp = std::filesystem::path(arg).extension() == ".webp";
            isAvif = std::filesystem::path(arg).extension() == ".avif";
            if (!isWebp && !isAvif) {
                std::cerr << "\"" << arg
                          << "\" is an unsupported file type for an input image. (supported types are avif and webp)"
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

    if (!input.size() || !output.size()) {
        syntax();
        return EXIT_FAILURE;
    }

    auto frameIndex = 0;
    char buffer[10];
    auto prevFrameTimestamp = 0;

    if (isWebp) {
        WebPData webpData;
        if (!ReadFile(input, &webpData.bytes, &webpData.size)) {
            std::cerr << "\"" << input << "\" failed to read input file." << std::endl;
            return EXIT_FAILURE;
        }

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

        cv::Mat frame(animInfo.canvas_height, animInfo.canvas_width, CV_8UC4);
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
    } else if (isAvif) {
        avifRGBImage rgb;

        auto decoder = avifDecoderCreate();
        decoder->maxThreads = std::thread::hardware_concurrency();

        auto result = avifDecoderSetIOFile(decoder, input.c_str());
        if (result != AVIF_RESULT_OK) {
            std::cerr << "\"" << input << "\" failed to read input file." << std::endl;
            return EXIT_FAILURE;
        }

        result = avifDecoderParse(decoder);
        if (result != AVIF_RESULT_OK) {
            std::cerr << "\"" << input << "\" failed to decode file." << std::endl;
            return EXIT_FAILURE;
        }

        std::cout << "frame_count: " << decoder->imageCount << " width: " << decoder->image->width << " height: " << decoder->image->height << std::endl;
        std::cout << "idx,delay" << std::endl;

        cv::Mat frame(decoder->image->height, decoder->image->width, CV_8UC4);

        avifRGBImageSetDefaults(&rgb, decoder->image);
        rgb.rowBytes = decoder->image->width * 4 * 1;
        rgb.pixels = frame.data;

        while (avifDecoderNextImage(decoder) == AVIF_RESULT_OK) {
            if (avifImageYUVToRGB(decoder->image, &rgb) != AVIF_RESULT_OK) {
                std::cerr << "\"" << input << "\" failed to decode file." << std::endl;
                return EXIT_FAILURE;
            }

            std::cout << frameIndex << "," << decoder->imageTiming.duration * 1000 << std::endl;

            sprintf(buffer, "%04d.png", frameIndex);
            auto filename = std::filesystem::path(output) / buffer;

            cv::cvtColor(frame, frame, cv::COLOR_RGBA2BGRA);
            cv::imwrite(filename, frame);

            frameIndex++;
        }

        frame.release();
        avifDecoderDestroy(decoder);
    }

    return EXIT_SUCCESS;
}
