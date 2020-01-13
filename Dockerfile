FROM golang:1.13.5-alpine3.10 as builder
RUN apk update && apk upgrade && \
    apk add \
    musl-dev \
    gcc
RUN mkdir -p /go/src/github.com/mendersoftware/create-artifact-worker
COPY . /go/src/github.com/mendersoftware/create-artifact-worker
RUN cd /go/src/github.com/mendersoftware/create-artifact-worker && env CGO_ENABLED=1 go build -o create-artifact

FROM  mendersoftware/workflows:master
RUN apk update && apk upgrade && \
    apk add --no-cache ca-certificates 

COPY ./config.yaml /etc/workflows
COPY --from=builder /go/src/github.com/mendersoftware/create-artifact-worker/create-artifact /usr/bin
ENTRYPOINT ["/usr/bin/workflows", "--worker", "--config", "/etc/workflows/config.yaml"]
