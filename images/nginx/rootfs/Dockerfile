# Start from Alpine.
FROM alpine:3.21.3 AS builder

# Install dependencies.
RUN apk add --no-cache musl-dev gcc

# Change directory.
WORKDIR /build

# Copy sources.
COPY main.c .

# Build binary.
RUN gcc main.c -static -o main

# Start from scratch.
FROM scratch

# Copy binary.
COPY --from=builder /build/main /main

# Run binary.
ENTRYPOINT ["/main"]
