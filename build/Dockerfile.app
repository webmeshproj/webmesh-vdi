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

ARG TARGETARCH TARGETOS
ADD dist/app_${TARGETOS}_${TARGETARCH}*/app /app
COPY --from=ui-builder /build/dist/spa /static
ADD swagger.json /static/swagger.json
COPY ui/app/src/statics /static/statics

ENTRYPOINT ["/app"]
