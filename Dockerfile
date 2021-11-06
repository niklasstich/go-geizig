# syntax=docker/dockerfile:1.3
#Build
FROM golang:latest AS build

WORKDIR /app
ENV CGO_ENABLED=0
COPY go.mod .
COPY go.sum .
RUN go mod download

#build
COPY *.go .
RUN --mount=type=cache,target=/root/.cache/go-build \
    GOOS=linux go build -o go-geizig -v .

FROM ubuntu:latest AS app
RUN apt-get update && apt-get install ca-certificates -y && update-ca-certificates
COPY --from=build /app/go-geizig /app/go-geizig
WORKDIR /app
CMD ["/app/go-geizig"]