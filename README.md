# ImageProcessor

## 1. For building and compiling this application

You can checkout he respective READMEs

- [C++](cpp/README.md)
- [Go](go/README.md)

## Design

![Diagram](./diagram.png)

The API Pods will send messages via RMQ in the format of a [Task](./go/internal/image_processor/task.go#L16)
The RMQ message must have the ReplyTo field set so that the image processor can know where to send the result back to.

The result payload will be in the structure of a [Result](./go/internal/image_processor/result.go#L24)
