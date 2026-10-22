FROM golang:1.27.0 AS builder

ARG ARCH=amd64

# Build binary
WORKDIR /
ADD . .
RUN make build

FROM golang:1.26.0

COPY --from=builder /build/redfish_exporter /usr/local/bin/redfish_exporter
RUN mkdir /etc/prometheus
COPY --from=builder /config.yml.example /etc/prometheus/redfish_exporter.yml
CMD ["/usr/local/bin/redfish_exporter","--config.file","/etc/prometheus/redfish_exporter.yml"]
