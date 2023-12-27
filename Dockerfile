FROM golang:rc-bullseye AS builder

LABEL maintainer="Jennings Liu <jenningsloy318@gmail.com>"

ARG ARCH=amd64

ENV GOROOT /usr/local/go
ENV GOPATH /go
ENV PATH "$GOROOT/bin:$GOPATH/bin:$PATH"
ENV GO_VERSION 1.15.2
ENV GO111MODULE=on 


# Build dependencies
RUN mkdir -p /go/src/github.com/ && \
    git clone https://github.com/GregWhiteyBialas/redfish_exporter.git /go/src/github.com/GregWhiteyBialas/redfish_exporter && \
    cd /go/src/github.com//GregWhiteyBialas/redfish_exporter && \
    make build

FROM golang:rc-bullseye

COPY --from=builder /go/src/github.com//GregWhiteyBialas/redfish_exporter/build/redfish_exporter /usr/local/bin/redfish_exporter
RUN mkdir /etc/prometheus
COPY config.yml.example /etc/prometheus/redfish_exporter.yml
CMD ["/usr/local/bin/redfish_exporter","--config.file","/etc/prometheus/redfish_exporter.yml"]
