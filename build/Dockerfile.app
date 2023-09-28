#################
# Compile image #
#################
ARG BASE_IMAGE=ghcr.io/tinyzimmer/kvdi:build-base-latest
FROM ${BASE_IMAGE} as builder

# Go build options
ENV GO111MODULE=on
ENV CGO_ENABLED=0

ARG GO_SWAGGER_VERSION=v0.23.0
ARG TARGETARCH=amd64
RUN curl -JL -o /usr/local/bin/swagger https://github.com/go-swagger/go-swagger/releases/download/${GO_SWAGGER_VERSION}/swagger_linux_${TARGETARCH} \
  && chmod +x /usr/local/bin/swagger

# Copy go code
COPY apis/        /build/apis
COPY pkg/         /build/pkg
COPY cmd/app      /build/cmd/app

# Build the binary and swagger json
ARG LDFLAGS
RUN go build -o /tmp/app \
  -ldflags="${LDFLAGS}" \
  ./cmd/app \
  && cd pkg/api \
  && /usr/local/bin/swagger generate spec -o /tmp/swagger.json --scan-models

ARG TARGETARCH=amd64
RUN [[ "${TARGETARCH}" == "amd64" ]] && upx /tmp/app || true

##############
# UI Builder #
##############
FROM node:16-alpine as ui-builder

RUN apk add --update build-base \
  && mkdir -p /build \
  && yarn global add @quasar/cli

COPY ui/app/package.json /build/package.json
COPY ui/app/yarn.lock /build/yarn.lock
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
# Latest quasar does not currently copy statics into dist
COPY ui/app/src/statics /static/statics

ENTRYPOINT ["/app"]
