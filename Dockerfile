FROM golang:1.13
# Create required dirs and copy files
RUN mkdir -p /mittens
COPY ./ /mittens/
WORKDIR /mittens
# Run unit tests & build app
RUN make build

FROM alpine:3.7
COPY --from=0 /mittens/mittens /app/mittens
ENTRYPOINT ["/app/mittens"]
