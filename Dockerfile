# syntax=docker/dockerfile:1

# creating a multi-stage build, we can selectively copy artifacts from one stage to another.
# any intermediate artifacts not copied are left behind, and not saved in the final image.

# use a golang base image to build our app.
FROM golang:1.21 AS builder

WORKDIR /build

# copy everything in repo to the working directory in image.
# note that in the image, ./ refer to working directory while / refers to root directory.
# also, our app is a go module so we can build using the vendor folder without needing to download the dependencies.
COPY ./ ./

# build the binary, here we are cross building for linux.
RUN CGO_ENABLED=0 GOOS=linux go build -o ./unravel-be ./cmd/main.go

# for the 2nd stage, we are using the thin alpine image.
FROM alpine:latest

WORKDIR /app

# copy only the binary from the previous stage.
COPY --from=builder /build/unravel-be ./unravel-be

# copy the config directory
COPY --from=builder /build/config ./config

EXPOSE 8080

# run the binary
CMD ["./unravel-be"]