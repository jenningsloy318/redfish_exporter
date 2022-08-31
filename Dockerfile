FROM golang:rc-bullseye AS builder

ARG https_proxy
ARG http_proxy
ARG ARCH=amd64

ENV GOROOT /usr/local/go
ENV GOPATH /go
ENV PATH "$GOROOT/bin:$GOPATH/bin:$PATH"
ENV GO_VERSION 1.15.2
ENV GO111MODULE=on 

# Build dependencies
RUN mkdir -p /go/src/github.com/ && \
    git clone https://github.com/vintury/redfish_exporter /go/src/github.com/vintury/redfish_exporter && \
    cd /go/src/github.com/vintury/redfish_exporter && \
    make build && \
    mkdir /etc/prometheus

FROM golang:rc-bullseye
COPY --from=builder /go/src/github.com/vintury/redfish_exporter/build/redfish_exporter /usr/local/bin/redfish_exporter

EXPOSE 9610
CMD ["/usr/local/bin/redfish_exporter","--config.file","/etc/prometheus/redfish_exporter.yml"]
