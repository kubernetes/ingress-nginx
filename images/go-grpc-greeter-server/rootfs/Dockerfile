ARG GOLANG_VERSION

FROM golang:${GOLANG_VERSION}-alpine3.21 AS build

WORKDIR /go/src/greeter-server

COPY . .
RUN go build -o /greeter-server

FROM gcr.io/distroless/base-debian10

COPY --from=build /greeter-server /

EXPOSE 50051

CMD ["/greeter-server"]
