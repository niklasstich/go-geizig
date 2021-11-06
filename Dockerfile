#Build
FROM golang:latest AS build

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

#build
COPY *.go .
RUN GOOS=linux go build -a -o go-geizig -v .

FROM ubuntu:latest AS app
RUN apt-get update && apt-get install ca-certificates -y && update-ca-certificates
COPY --from=build /app/go-geizig /app/go-geizig
WORKDIR /app
CMD ["/app/go-geizig"]