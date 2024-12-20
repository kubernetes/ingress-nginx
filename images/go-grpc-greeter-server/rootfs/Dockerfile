ARG GOLANG_VERSION

FROM golang:${GOLANG_VERSION}-alpine3.21 as build

WORKDIR /go/src/greeter-server

COPY main.go .
RUN go mod init greeter-server && \
  go mod tidy && \
  go build -o /greeter-server main.go

FROM gcr.io/distroless/base-debian10

COPY --from=build /greeter-server /

EXPOSE 50051

CMD ["/greeter-server"]
