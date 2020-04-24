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

# Fetch deps first as they don't change frequently
COPY go.mod /build/go.mod
COPY go.sum /build/go.sum
RUN go mod download

# Copy go code
COPY version/ /build/version
COPY pkg/     /build/pkg
COPY cmd/app  /build/cmd/app

# Build the binary
RUN go build \
  -o /tmp/app \
  ./cmd/app && upx /tmp/app

##############
# UI Builder #
##############
FROM node:14-alpine as ui-builder

RUN apk add --update python2 build-base && mkdir -p /build && yarn global add @quasar/cli

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

EXPOSE 8443
ENTRYPOINT ["/app"]
