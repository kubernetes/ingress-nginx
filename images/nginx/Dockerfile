# Define Alpine version.
ARG ALPINE=3.21.0

# Start from Alpine.
FROM alpine:${ALPINE}

# Install packages.
RUN apk add cmake make gcc g++ libc-dev gtest-dev benchmark-dev

# Copy files.
COPY main.c /build/main/main.c

# Build main.
RUN gcc /build/main/main.c -o /build/main/main

# Add LuaJIT.
ADD https://github.com/openresty/luajit2.git#v2.1-20241113 /build/luajit2

# Build LuaJIT.
RUN make --directory /build/luajit2 --jobs $(getconf _NPROCESSORS_ONLN)

# Add OpenTelemetry C++.
ADD https://github.com/open-telemetry/opentelemetry-cpp.git#v1.18.0 /build/opentelemetry-cpp

# Configure OpenTelemetry C++.
RUN cmake -S /build/opentelemetry-cpp -B /build/opentelemetry-cpp/build

# Build OpenTelemetry C++.
RUN cmake --build /build/opentelemetry-cpp/build --parallel $(getconf _NPROCESSORS_ONLN)
