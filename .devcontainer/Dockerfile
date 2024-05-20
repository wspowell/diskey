FROM golang:1.22.1-alpine3.19

RUN apk add --no-cache build-base make gcc git

RUN apk add --update --no-cache \
           graphviz \
           ttf-freefont

# Install protoc
RUN apk update && apk add --no-cache make protobuf-dev

RUN go install golang.org/x/tools/...@v0.13.0
RUN go install gotest.tools/gotestsum@v1.10.1
RUN go install github.com/uudashr/gopkgs/v2/cmd/gopkgs@v2.1.2
# The version is correct for go-outline. The tag is not prefixed with 'v'.
RUN go install github.com/ramya-rao-a/go-outline@1.0.0
RUN go install github.com/cweill/gotests/...@v1.6.0
RUN go install github.com/fatih/gomodifytags@v1.16.0
RUN go install github.com/josharian/impl@v1.2.0
RUN go install github.com/haya14busa/goplay/cmd/goplay@v1.0.0
RUN go install github.com/go-delve/delve/cmd/dlv@v1.22.1
RUN go install github.com/vektra/mockery/v2@v2.33.2
RUN go install mvdan.cc/gofumpt@v0.5.0
RUN go install github.com/daixiang0/gci@v0.11.1
RUN go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.15.0
RUN go install golang.org/x/tools/gopls@v0.15.1
RUN go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@v0.13.0

RUN go install github.com/google/pprof@latest

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.54.2

CMD [ "sleep", "infinity" ]