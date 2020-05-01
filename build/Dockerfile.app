#################
# Compile image #
#################
FROM golang:1.14-alpine as builder

RUN apk --update-cache add upx

# Setup build directory
RUN mkdir -p /build
WORKDIR /build

# Go build options
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ARG VERSION
ENV VERSION=${VERSION}
ARG GIT_COMMIT
ENV GIT_COMMIT=${GIT_COMMIT}

# Fetch deps first as they don't change frequently
COPY go.mod /build/go.mod
COPY go.sum /build/go.sum
RUN go mod download

ARG GO_SWAGGER_VERSION=v0.23.0
RUN apk add --update curl \
  && curl -JL -o /usr/local/bin/swagger https://github.com/go-swagger/go-swagger/releases/download/${GO_SWAGGER_VERSION}/swagger_linux_amd64 \
  && chmod +x /usr/local/bin/swagger

# Copy go code
COPY version/     /build/version
COPY pkg/         /build/pkg
COPY cmd/app      /build/cmd/app

# Build the binary and swagger json
RUN go build -o /tmp/app \
    -ldflags="-X 'github.com/tinyzimmer/kvdi/version.Version=${VERSION}' -X 'github.com/tinyzimmer/kvdi/version.GitCommit=${GIT_COMMIT}'" \
    ./cmd/app \
  && upx /tmp/app \
  && cd pkg/api \
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
