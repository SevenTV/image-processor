#include <filesystem>
#include <fstream>
#include <iostream>
#include <opencv2/opencv.hpp>
#include <string>
#include <thread>
#include <vector>

#include "resize_png.hpp"

#define NEXTARG()                                                     \
    if (((argIndex + 1) == argc) || (argv[argIndex + 1][0] == '-')) { \
        std::cerr << arg << " requires an argument." << std::endl;    \
        return EXIT_FAILURE;                                          \
    }                                                                 \
    arg = std::string(argv[++argIndex])

void syntax()
{
    std::cerr << "Syntax: resize_png [options] -i input.png -r 100 100 -o out.png -r 50 50 -o out2.png"
              << std::endl
              << "Options:" << std::endl
              << "  -h,--help                   : Shows syntax help" << std::endl
              << "  -i,--input FILENAME         : Input file location (supported "
                 "types are png)."
              << std::endl
              << "  -r,--resize 100 100         : The width and height"
              << std::endl
              << "  -o,--output FILENAME        : Output filename"
                 " (supported types are png)."
              << std::endl
              << std::endl;
}

int main(int argc, char* argv[])
{
    std::vector<Output> outputs;

    File currentInput;
    int currentWidth, currentHeight;

    int argIndex = 1;
    while (argIndex < argc) {
        std::string arg = argv[argIndex];

        if (arg == "--help" || arg == "-h") {
            syntax();
            return EXIT_FAILURE;
        } else if (arg == "--output" || arg == "-o") {
            if (!currentInput.data.data) {
                std::cerr << "\"" << arg
                          << "\" You must provide an input before specifying an output."
                          << std::endl;
                return EXIT_FAILURE;
            }
            if (currentWidth <= 0 || currentHeight <= 0) {
                std::cerr << "\"" << arg
                          << "\" You must provide a resize before specifying an output."
                          << std::endl;
                return EXIT_FAILURE;
            }
            NEXTARG();
            Output output;
            output.path = arg;
            output.input = currentInput;
            output.height = currentHeight;
            output.width = currentWidth;

            currentHeight = 0;
            currentWidth = 0;
            currentInput.used = true;

            outputs.push_back(output);
        } else if (arg == "--input" || arg == "-i") {
            NEXTARG();
            if (std::filesystem::path(arg).extension() != ".png") {
                std::cerr << "\"" << arg
                          << "\" is an unsupported file type for an input image."
                          << std::endl;
                return EXIT_FAILURE;
            }
            if (!currentInput.used && currentInput.data.data) {
                std::cerr << "\"" << arg
                          << "\" Unconnected input image."
                          << std::endl;
                return EXIT_FAILURE;
            }

            currentInput.used = false;
            currentInput.data = cv::imread(arg, cv::IMREAD_UNCHANGED);
            if (!currentInput.data.data) {
                std::cerr << "Invalid input image: " << arg << std::endl;
                return EXIT_FAILURE;
            }
        } else if (arg == "--resize" || arg == "-r") {
            NEXTARG();
            currentWidth = std::stoi(arg);
            if (currentWidth <= 0) {
                std::cerr << "Invalid resize width: " << arg << std::endl;
                return EXIT_FAILURE;
            }
            NEXTARG();
            currentHeight = std::stoi(arg);
            if (currentHeight <= 0) {
                std::cerr << "Invalid resize width: " << arg << std::endl;
                return EXIT_FAILURE;
            }
        } else {
            std::cerr << "\"" << arg << "\" is an unknown argument." << std::endl;
            return EXIT_FAILURE;
        }

    loop:
        argIndex++;
    }

    if (outputs.size() == 0) {
        std::cerr << "0 output files provided, at least 1 output file is required."
                  << std::endl;
        return EXIT_FAILURE;
    }

    if (!currentInput.used && currentInput.data.data) {
        std::cerr << "Unconnected input image."
                  << std::endl;
        return EXIT_FAILURE;
    }

    for (auto output : outputs) {
        cv::Mat img;
        cv::resize(output.input.data, img, cv::Size(output.width, output.height), 0, 0, cv::INTER_NEAREST_EXACT);
        cv::imwrite(output.path, img);
        img.release();
    }

    for (auto output : outputs) {
        if (output.input.data.data) {
            output.input.data.release();
        }
    }

    if (currentInput.data.data) {
        currentInput.data.release();
    }

    return EXIT_SUCCESS;
}
