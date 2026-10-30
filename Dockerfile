FROM golang:1.27.0 AS builder

ARG ARCH=amd64

# Build binary
WORKDIR /
ADD collector collector
ADD common common
ADD go.mod go.mod
ADD go.sum go.sum
ADD main.go main.go
ADD Makefile Makefile
RUN make build

FROM gcr.io/distroless/base-debian13

ADD config.yml.example /etc/prometheus/redfish_exporter.yml
COPY --from=builder /build/redfish_exporter /usr/local/bin/redfish_exporter

EXPOSE 9610

CMD ["/usr/local/bin/redfish_exporter","--config.file","/etc/prometheus/redfish_exporter.yml"]
