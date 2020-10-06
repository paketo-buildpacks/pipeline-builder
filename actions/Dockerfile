FROM golang:latest as build-stage
WORKDIR /build
COPY . .
ARG SOURCE
RUN GOOS="linux" go build -ldflags='-s -w' -o /bin/action actions/$SOURCE/main.go

FROM scratch
COPY --from=build-stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-stage /bin/action /bin/
ENTRYPOINT ["/bin/action"]
