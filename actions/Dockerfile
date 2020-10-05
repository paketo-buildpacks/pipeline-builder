FROM golang:latest as build-stage
WORKDIR /build
COPY . .
ARG SOURCE
RUN GOOS="linux" go build -ldflags='-s -w' -o /bin/action actions/$SOURCE/main.go

FROM ubuntu:latest
COPY --from=build-stage /bin/action /bin/action
ENTRYPOINT /bin/action
