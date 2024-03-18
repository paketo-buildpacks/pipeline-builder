#!/usr/bin/env bash
set -euo pipefail

GOMOD=$(head -1 go.mod | awk '{print $2}')

{{- range $key, $value := .}}
GOOS="linux" go build -ldflags='-s -w' -o "linux/amd64/{{ $key }}" "{{ $value }}"
GOOS="linux" GOARCH="arm64" go build -ldflags='-s -w' -o "linux/arm64/{{ $key }}" "{{ $value }}"
{{- end }}
GOOS="linux" go build -ldflags='-s -w' -o linux/amd64/bin/main "$GOMOD/cmd/main"
GOOS="linux" GOARCH="arm64" go build -ldflags='-s -w' -o linux/arm64/bin/main "$GOMOD/cmd/main"

if [ "${STRIP:-false}" != "false" ]; then
  {{- range $key, $value := .}}
  strip linux/amd64/{{ $key }} linux/arm64/{{ $key }}
  {{- end }}
  strip linux/amd64/bin/main linux/arm64/bin/main
fi

if [ "${COMPRESS:-none}" != "none" ]; then
  {{- range $key, $value := .}}
  $COMPRESS linux/amd64/{{ $key }} linux/arm64/{{ $key }}
  {{- end }}
  $COMPRESS linux/amd64/bin/main linux/arm64/bin/main
fi

ln -fs main linux/amd64/bin/build
ln -fs main linux/arm64/bin/build
ln -fs main linux/amd64/bin/detect
ln -fs main linux/arm64/bin/detect