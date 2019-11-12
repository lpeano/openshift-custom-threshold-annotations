FROM golang:latest

USER nobody

RUN mkdir -p /go/src/thresholdsMetrics
WORKDIR /go/src/thresholdsMetrics

COPY . /go/src/thresholdsMetrics
RUN go build

CMD ["./thresholdsMetrics"]
