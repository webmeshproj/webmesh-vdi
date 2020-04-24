#################
# Compile image #
#################
FROM golang:1.14-alpine as builder

ARG GO_SWAGGER_VERSION=v0.23.0
RUN apk add --update upx curl \
  && mkdir -p /go/src/github.com/tinyzimmer/kvdi \
  && curl -JL -o /usr/local/bin/swagger https://github.com/go-swagger/go-swagger/releases/download/${GO_SWAGGER_VERSION}/swagger_linux_amd64 \
  && chmod +x /usr/local/bin/swagger

# Setup build directory
WORKDIR /go/src/github.com/tinyzimmer/kvdi

# Go build options
ENV GO111MODULE=on
ENV CGO_ENABLED=0

# Fetch deps first as they don't change frequently
COPY go.mod /go/src/github.com/tinyzimmer/kvdi/go.mod
COPY go.sum /go/src/github.com/tinyzimmer/kvdi/go.sum
RUN go mod download

# Copy go code
COPY version/ /go/src/github.com/tinyzimmer/kvdi/version
COPY pkg/     /go/src/github.com/tinyzimmer/kvdi/pkg
COPY cmd/app  /go/src/github.com/tinyzimmer/kvdi/cmd/app

# Build the binary and swagger json
RUN go build -o /tmp/app ./cmd/app \
  && upx /tmp/app \
  && cd /go/src/github.com/tinyzimmer/kvdi/pkg/api \
  && /usr/local/bin/swagger generate spec -o /tmp/swagger.json --scan-models

##############
# UI Builder #
##############
FROM node:14-alpine as ui-builder

RUN apk add --update python2 build-base \
  && mkdir -p /build \
  && yarn global add @quasar/cli

COPY ui/app/package.json /build/package.json
RUN cd /build && yarn

COPY ui/app/ /build/
RUN cd /build && quasar build

###############
# Final Image #
###############
FROM scratch

COPY --from=builder /tmp/app /app
COPY --from=ui-builder /build/dist/spa /static
COPY --from=builder /tmp/swagger.json /static/swagger.json

EXPOSE 8443
ENTRYPOINT ["/app"]
