#include <avif/avif.h>
#include <filesystem>
#include <fstream>
#include <gifski.h>
#include <iostream>
#include <opencv2/opencv.hpp>
#include <string>
#include <thread>
#include <vector>
#include <webp/decode.h>
#include <webp/encode.h>
#include <webp/mux.h>

#include "convert_png.hpp"

#define NEXTARG()                                                     \
    if (((argIndex + 1) == argc) || (argv[argIndex + 1][0] == '-')) { \
        std::cerr << arg << " requires an argument." << std::endl;    \
        return EXIT_FAILURE;                                          \
    }                                                                 \
    arg = std::string(argv[++argIndex])

void syntax()
{
    std::cerr << "Syntax: convert_png [options] -i input.png -o output.webp -o "
                 "output.gif -o output.avif"
              << std::endl
              << "Options:" << std::endl
              << "  -h,--help                   : Shows syntax help" << std::endl
              << "  -i,--input FILENAME         : Input file location (supported "
                 "types are png)."
              << std::endl
              << "  -t,--threads THREADS        : The number of threads to use." << std::endl
              << "  -o,--output FILENAME        : Output file location "
                 " (supported types are webp, avif, gif)."
              << std::endl
              << "  -d,--delay D                : Delay of the next frame in "
                 "100s of a second. (default 4 = 40ms)"
              << std::endl
              << std::endl;
}

bool equal(const cv::Mat& a, const cv::Mat& b)
{
    if ((a.rows != b.rows) || (a.cols != b.cols))
        return false;

    cv::Mat c;
    cv::bitwise_xor(a, b, c);

    auto s = cv::sum(c);
    return (s[0] == 0) && (s[1] == 0) && (s[2] == 0) && (s[3] == 0);
}

int main(int argc, char* argv[])
{
    int delay = 4;

    std::vector<File> inputs;
    std::vector<Output> outputs;

    int width, height = 0;

    int threads = std::thread::hardware_concurrency();

    int argIndex = 1;
    while (argIndex < argc) {
        std::string arg = argv[argIndex];

        if (arg == "--delay" || arg == "-d") {
            NEXTARG();
            delay = std::atoi(arg.c_str());
            if (delay <= 0) {
                std::cerr << "\"" << arg << "\" is not a valid value for delay."
                          << std::endl;
                return EXIT_FAILURE;
            }
        } else if (arg == "--threads" || arg == "-t") {
            NEXTARG();
            threads = std::atoi(arg.c_str());
            if (threads <= 0) {
                std::cerr << "\"" << arg << "\" is not a valid value for threads."
                          << std::endl;
                return EXIT_FAILURE;
            }
        } else if (arg == "--help" || arg == "-h") {
            syntax();
            return EXIT_FAILURE;
        } else if (arg == "--output" || arg == "-o") {
            NEXTARG();
            Output output;
            output.path = arg;
            auto ext = output.path.extension();
            if (ext == ".webp") {
                output.type = OutputType::WEBP;
            } else if (ext == ".avif") {
                output.type = OutputType::AVIF;
            } else if (ext == ".gif") {
                output.type = OutputType::GIF;
            } else {
                std::cerr << "\"" << arg
                          << "\" is an unsupported file type for an output image."
                          << std::endl;
                return EXIT_FAILURE;
            }

            outputs.push_back(output);
        } else if (arg == "--input" || arg == "-i") {
            NEXTARG();
            if (std::filesystem::path(arg).extension() != ".png") {
                std::cerr << "\"" << arg
                          << "\" is an unsupported file type for an input image."
                          << std::endl;
                return EXIT_FAILURE;
            }

            File f;
            f.delay = delay;
            f.data = cv::imread(arg, cv::IMREAD_UNCHANGED);
            cv::cvtColor(f.data, f.data, cv::COLOR_BGRA2RGBA);
            if (!f.data.data) {
                std::cerr << "Invalid input image: " << arg << std::endl;
                return EXIT_FAILURE;
            }
            if (inputs.size() != 0) {
                auto lastInput = inputs[inputs.size() - 1];
                if (equal(lastInput.data, f.data)) {
                    lastInput.delay += f.delay;
                    inputs[inputs.size() - 1] = lastInput;
                    goto loop;
                }
            }

            auto size = f.data.size();
            width = size.width;
            height = size.height;
            inputs.push_back(f);
        } else {
            std::cerr << "\"" << arg << "\" is an unknown argument." << std::endl;
            return EXIT_FAILURE;
        }

    loop:
        argIndex++;
    }

    if (inputs.size() == 0) {
        std::cerr << "0 input files provided, at least 1 input file is required."
                  << std::endl;
        return EXIT_FAILURE;
    }

    if (outputs.size() == 0) {
        std::cerr << "0 output files provided, at least 1 output file is required."
                  << std::endl;
        return EXIT_FAILURE;
    }

    for (auto output : outputs) {
        if (output.type == OutputType::AVIF) {
            auto encoder = avifEncoderCreate();

            encoder->maxThreads = threads;
            encoder->minQuantizer = 5;
            encoder->maxQuantizer = 20;
            encoder->minQuantizerAlpha = 0;
            encoder->maxQuantizerAlpha = 10;
            encoder->tileColsLog2 = 2;
            encoder->tileRowsLog2 = 2;
            encoder->speed = 4;
            encoder->timescale = 100;
            encoder->keyframeInterval = 0;

            auto image = avifImageCreateEmpty();
            image->colorPrimaries = AVIF_COLOR_PRIMARIES_BT709;
            image->transferCharacteristics = AVIF_TRANSFER_CHARACTERISTICS_SRGB;
            image->matrixCoefficients = AVIF_MATRIX_COEFFICIENTS_BT601;
            image->yuvRange = AVIF_RANGE_FULL;
            image->alphaPremultiplied = false;
            image->width = width;
            image->height = height;
            image->depth = 8;
            image->yuvFormat = AVIF_PIXEL_FORMAT_YUV444;

            avifRGBImage rgb;
            avifRWData avifOutput = AVIF_DATA_EMPTY;
            avifRGBImageSetDefaults(&rgb, image);

            rgb.format = avifRGBFormat::AVIF_RGB_FORMAT_RGBA;
            for (int i = 0; i < inputs.size(); i++) {
                auto input = inputs[i];
                rgb.rowBytes = 4 * image->width;
                rgb.pixels = input.data.data;
                rgb.depth = 8;
                auto res = avifImageRGBToYUV(image, &rgb);
                if (res != AVIF_RESULT_OK) {
                    std::cerr << "Failed to convert to YUV(A): " << avifResultToString(res) << std::endl;
                    return EXIT_FAILURE;
                }

                res = avifEncoderAddImage(encoder, image, input.delay, avifAddImageFlag::AVIF_ADD_IMAGE_FLAG_NONE);
                if (res != AVIF_RESULT_OK) {
                    std::cerr << "Failed to add image to encoder: " << avifResultToString(res) << std::endl;
                    return EXIT_FAILURE;
                }
            }

            auto res = avifEncoderFinish(encoder, &avifOutput);
            if (res != AVIF_RESULT_OK) {
                std::cerr << "Failed to finish encode: " << avifResultToString(res) << std::endl;
                return EXIT_FAILURE;
            }

            std::ofstream fout;
            fout.open(output.path, std::ios::binary | std::ios::out);
            fout.write((const char*)avifOutput.data, avifOutput.size);
            fout.close();

            if (image) {
                avifImageDestroy(image);
            }
            if (encoder) {
                avifEncoderDestroy(encoder);
            }
            avifRWDataFree(&avifOutput);
        } else if (output.type == OutputType::WEBP) {
            WebPAnimEncoderOptions anim_config;
            WebPConfig config;
            WebPPicture pic;
            WebPData webp_data;

            WebPDataInit(&webp_data);
            if (!WebPAnimEncoderOptionsInit(&anim_config) || !WebPConfigInit(&config) || !WebPPictureInit(&pic)) {
                std::cerr << "Library version mismatch." << std::endl;
                return EXIT_FAILURE;
            }

            anim_config.allow_mixed = 1;
            config.quality = 95;
            config.lossless = 1;
            config.thread_level = 0;

            auto ok = WebPValidateConfig(&config);
            if (!ok) {
                std::cerr << "Invalid WebP Config" << std::endl;
                return EXIT_FAILURE;
            }

            pic.use_argb = 1;
            pic.width = width;
            pic.height = height;

            auto enc = WebPAnimEncoderNew(width, height, &anim_config);
            if (!enc) {
                std::cerr << "Could not create WebPAnimEncoder object." << std::endl;
                return EXIT_FAILURE;
            }

            auto ts = 0;
            for (int i = 0; i < inputs.size(); i++) {
                auto input = inputs[i];

                ok = WebPPictureImportRGBA(&pic, input.data.data, 4 * width * sizeof(*input.data.data));
                ok = ok && WebPAnimEncoderAdd(enc, &pic, ts, &config);
                if (!ok) {
                    std::cerr << "WebP error while adding frame #" << i << ": " << pic.error_code << std::endl;
                    return EXIT_FAILURE;
                }

                WebPPictureFree(&pic);
                ts += input.delay * 10;
            }

            ok = ok && WebPAnimEncoderAdd(enc, NULL, ts, NULL);
            ok = ok && WebPAnimEncoderAssemble(enc, &webp_data);
            if (!ok) {
                std::cerr << "Error during final animation assembly." << std::endl;
                return EXIT_FAILURE;
            }

            WebPAnimEncoderDelete(enc);

            if (inputs.size() > 1) {
                auto mux = WebPMuxCreate(&webp_data, 1);
                if (mux == NULL) {
                    std::cerr << "ERROR: Could not re-mux to add loop count/metadata." << std::endl;
                    return EXIT_FAILURE;
                }
                WebPDataClear(&webp_data);

                WebPMuxAnimParams new_params;
                auto err = WebPMuxGetAnimationParams(mux, &new_params);
                if (err != WEBP_MUX_OK) {
                    std::cerr << "ERROR: Could not fetch loop count. " << err << std::endl;
                    return EXIT_FAILURE;
                }

                new_params.loop_count = 0;
                err = WebPMuxSetAnimationParams(mux, &new_params);
                if (err != WEBP_MUX_OK) {
                    std::cerr << "ERROR: Could not update loop count. " << err << std::endl;
                    return EXIT_FAILURE;
                }

                err = WebPMuxAssemble(mux, &webp_data);
                if (err != WEBP_MUX_OK) {
                    std::cerr << "ERROR: Could not assemble when re-muxing to add loop count/metadata. " << err << std::endl;
                    return EXIT_FAILURE;
                }

                WebPMuxDelete(mux);
            }

            std::ofstream fout;
            fout.open(output.path, std::ios::binary | std::ios::out);
            fout.write((const char*)webp_data.bytes, webp_data.size);
            fout.close();

            WebPDataClear(&webp_data);

        } else if (output.type == OutputType::GIF) {
            GifskiSettings settings;
            settings.quality = 95;
            settings.fast = false;
            settings.height = height;
            settings.width = width;
            settings.repeat = 0;

            auto g = gifski_new(&settings);
            if (!g) {
                std::cerr << "GifSki init failed" << std::endl;
                return EXIT_FAILURE;
            }

            auto res = gifski_set_file_output(g, output.path.c_str());
            if (res != GIFSKI_OK) {
                std::cerr << "GifSki Failed 1: " << res << std::endl;
                return EXIT_FAILURE;
            }

            double offset = 0;
            for (int i = 0; i < inputs.size(); i++) {
                auto input = inputs[i];

                auto res = gifski_add_frame_rgba(g, i, width, height, input.data.data, offset / 100);
                if (res != GIFSKI_OK) {
                    std::cerr << "GifSki Failed 2: " << res << std::endl;
                    return EXIT_FAILURE;
                }

                offset += input.delay;
            }

            res = gifski_finish(g);
            if (res != GIFSKI_OK) {
                std::cerr << "GifSki Failed 3: " << res << std::endl;
                return EXIT_FAILURE;
            }
        }
    }

    for (auto input : inputs) {
        input.data.release();
    }

    return EXIT_SUCCESS;
}
