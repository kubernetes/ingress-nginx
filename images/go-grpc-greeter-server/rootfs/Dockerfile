FROM golang:1.20.1-alpine3.17 as build

WORKDIR /go/src/greeter-server

COPY main.go .
RUN go mod init greeter-server && \
  go mod tidy && \
  go build -o /greeter-server main.go

FROM gcr.io/distroless/base-debian10

COPY --from=build /greeter-server /

EXPOSE 50051

CMD ["/greeter-server"]
