FROM golang:buster as build

WORKDIR /go/src/greeter-server

RUN curl -o main.go https://raw.githubusercontent.com/grpc/grpc-go/master/examples/features/reflection/server/main.go && \
  go mod init greeter-server && \
  go mod tidy && \
  go build -o /greeter-server main.go

FROM gcr.io/distroless/base-debian10

COPY --from=build /greeter-server /

EXPOSE 50051

CMD ["/greeter-server"]
