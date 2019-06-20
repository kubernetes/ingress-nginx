# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/ncabatoff/process-exporter

# Build the process-exporter command inside the container.
RUN make -C /go/src/github.com/ncabatoff/process-exporter

USER root

# Run the process-exporter command by default when the container starts.
ENTRYPOINT ["/go/src/github.com/ncabatoff/process-exporter/process-exporter"]

# Document that the service listens on port 9256.
EXPOSE 9256
