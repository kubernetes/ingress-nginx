# Start from Alpine.
FROM alpine:3.21.0

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
