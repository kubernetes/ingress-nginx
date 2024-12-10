# Start from Alpine.
FROM alpine:3.21.0

# Install packages.
RUN apk add cmake make gcc g++ libc-dev gtest-dev benchmark-dev

# Copy files.
COPY main.c /build/main/main.c

# Build main.
RUN gcc /build/main/main.c -o /build/main/main
