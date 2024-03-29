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
              << "  -h,--help                   : Shows syntax help." << std::endl
              << "  --resize-ratio [1-6]        : Resize the ratio." << std::endl
              << "     1. stretch (default)" << std::endl
              << "     2. left-bottom" << std::endl
              << "     3. right-bottom" << std::endl
              << "     4. left-top" << std::endl
              << "     5. right-top" << std::endl
              << "     6. center" << std::endl
              << "  -i,--input FILENAME         : Input file location (supported "
                 "types are png)."
              << std::endl
              << "  -r,--resize 100 100         : The width and height."
              << std::endl
              << "  -o,--output FILENAME        : Output filename."
                 " (supported types are png)."
              << std::endl
              << std::endl;
}

int main(int argc, char* argv[])
{
    std::vector<Output> outputs;

    File currentInput;
    int currentWidth, currentHeight;
    int resizeRatio = 1;

    int argIndex = 1;
    while (argIndex < argc) {
        std::string arg = argv[argIndex];

        if (arg == "--help" || arg == "-h") {
            syntax();
            return EXIT_FAILURE;
        } else if (arg == "--resize-ratio") {
            NEXTARG();

            resizeRatio = std::stoi(arg);
            if (resizeRatio <= 0 || resizeRatio > 6) {
                std::cerr << "Invalid resize ratio: " << arg << std::endl;
                return EXIT_FAILURE;
            }
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
            output.resizeRatio = resizeRatio;

            currentHeight = 0;
            currentWidth = 0;
            currentInput.used = true;
            resizeRatio = 1;

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

            if (currentInput.data.channels() == 4) {
                // we need to correct the transparent pixels on this image so that resizing doesnt make weird artifacts
                std::vector<cv::Mat> channels;
                cv::split(currentInput.data, channels); // break image into channels

                cv::Mat adjacent;

                auto alpha = channels[3]; // get the alpha channel

                cv::Mat noAlpha;

                channels.pop_back();
                cv::merge(channels, noAlpha);

                cv::Mat adj;
                cv::dilate(alpha, adj, cv::Mat(), cv::Point(-1, -1), 3);

                cv::inpaint(noAlpha, alpha == 0 & adj, noAlpha, 1, cv::INPAINT_TELEA); // inpaint the alpha channel
                adj.release();

                cv::split(noAlpha, channels); // split the image back into channels
                noAlpha.release();

                channels.push_back(alpha); // add the alpha channel back into the channels

                cv::merge(channels, currentInput.data); // merge the channels back into an image
                for (auto& channel : channels) {
                    channel.release();
                }
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
        cv::Mat img = output.input.data;
        auto newSize = cv::Size(output.width, output.height);

        if (output.resizeRatio != 1) {
            auto currentSize = output.input.data.size();

            auto currentRatio = currentSize.aspectRatio();
            auto newRatio = newSize.aspectRatio();
            if (currentRatio != newRatio) {
                cv::Mat padded;

                if (currentRatio < newRatio) { // means that width is too small
                    padded.create(img.rows, int(double(img.rows) * newRatio), img.type());
                } else { // means that height is too small
                    padded.create(int(double(img.cols) / newRatio), img.cols, img.type());
                }

                padded.setTo(cv::Scalar::all(0));

                int x;
                int y;

                if (output.resizeRatio == 2) {
                    x = 0;
                    y = 0;
                } else if (output.resizeRatio == 3) {
                    x = padded.cols - img.cols;
                    y = 0;
                } else if (output.resizeRatio == 4) {
                    x = 0;
                    y = padded.rows - img.rows;
                } else if (output.resizeRatio == 5) {
                    x = padded.cols - img.cols;
                    y = padded.rows - img.rows;
                } else if (output.resizeRatio == 6) {
                    x = (padded.cols - img.cols) / 2;
                    y = (padded.rows - img.rows) / 2;
                }

                img.copyTo(padded(cv::Rect(x, y, img.cols, img.rows)));

                output.input.data = padded;
            }
        }

        cv::resize(output.input.data, img, newSize, 0, 0, cv::INTER_AREA);
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
