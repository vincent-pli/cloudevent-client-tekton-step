FROM gcr.io/distroless/base:latest

WORKDIR /

COPY _output/bin/cloudeventclient /usr/local/bin/cloudeventclient

ENTRYPOINT ["/usr/local/bin/cloudeventclient"]
