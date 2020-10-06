FROM golang:latest as build-stage

RUN apt-get update && apt-get -y install upx

WORKDIR /src
ENV GO111MODULE=on CGO_ENABLED=0

COPY . .

ARG SOURCE
RUN go build \
  -a \
  -trimpath \
  -ldflags "-s -w -extldflags '-static'" \
  -installsuffix cgo \
  -tags netgo \
  -o /bin/action \
  actions/$SOURCE/main.go

RUN strip /bin/action

RUN upx -q -9 /bin/action

FROM scratch
COPY --from=build-stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build-stage /bin/action /bin/action
ENTRYPOINT ["/bin/action"]
