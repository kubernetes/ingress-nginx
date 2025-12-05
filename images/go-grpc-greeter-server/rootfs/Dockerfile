ARG GOLANG_VERSION

FROM golang:${GOLANG_VERSION}-alpine3.23 AS build

WORKDIR /go/src/greeter-server

COPY . .
RUN go build -o /greeter-server

FROM scratch

COPY --from=build /greeter-server /

EXPOSE 50051

CMD ["/greeter-server"]
