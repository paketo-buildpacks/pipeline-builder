#!/usr/bin/env bash
set -euo pipefail

GOMOD=$(head -1 go.mod | awk '{print $2}')

{{- range $key, $value := .Helpers}}
{{- range $.Architectures}}
GOOS="linux" GOARCH="{{.}}" go build -ldflags='-s -w' -o "linux/{{.}}/{{ $key }}" "{{ $value }}"
{{- end}}
{{- end}}
{{- range $.Architectures}}
GOOS="linux" GOARCH="{{.}}" go build -ldflags='-s -w' -o "linux/{{.}}/bin/main" "$GOMOD/cmd/main"
{{- end}}

if [ "${STRIP:-false}" != "false" ]; then
  {{- range $key, $value := .Helpers}}
  strip {{- range $.Architectures}} linux/{{.}}/{{ $key }}{{- end}}
  {{- end}}
  strip {{- range $.Architectures}} linux/{{.}}/bin/main{{- end}}
fi

if [ "${COMPRESS:-none}" != "none" ]; then
  {{- range $key, $value := .Helpers}}
  $COMPRESS {{- range $.Architectures}} linux/{{.}}/{{ $key }}{{- end}}
  {{- end}}
  $COMPRESS {{- range $.Architectures}} linux/{{.}}/bin/main{{- end}}
fi

{{- range $.Architectures}}
ln -fs main linux/{{.}}/bin/build
ln -fs main linux/{{.}}/bin/detect
{{- end}}
