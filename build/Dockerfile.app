# Copy release asssets to scratch image
FROM scratch

ARG TARGETARCH TARGETOS
ADD dist/app_${TARGETOS}_${TARGETARCH}*/app /app
ADD ui/app/dist/spa /static
ADD ui/swagger.json /static/swagger.json
ADD ui/app/src/statics /static/statics

ENTRYPOINT ["/app"]
