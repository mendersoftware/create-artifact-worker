ARG WORKFLOWS_VERSION=master
ARG MENDER_ARTIFACT_VERSION=3.9.0

FROM golang:1.19.3-alpine3.16 as builder
RUN apk add --no-cache \
    ca-certificates \
    musl-dev \
    gcc \
    git
WORKDIR /go/src/github.com/mendersoftware/create-artifact-worker
COPY ./ .
RUN env CGO_ENABLED=0 go build -o create-artifact

FROM mendersoftware/workflows:$WORKFLOWS_VERSION as workflows

FROM alpine:3.17.1
ARG MENDER_ARTIFACT_VERSION
RUN apk add --no-cache \ 
    xz \
    libc6-compat \
    binutils \
    file \
    rsync \
    parted \
    e2fsprogs \
    xfsprogs \
    pigz \
    dosfstools \
    wget \
    make \ 
    bash
    # bmap-tools not found

RUN sed -i 's/ash/bash/g' /etc/passwd
ADD https://downloads.mender.io/mender-artifact/$MENDER_ARTIFACT_VERSION/linux/mender-artifact /usr/bin/mender-artifact
ADD https://raw.githubusercontent.com/mendersoftware/mender/master/support/modules/single-file /usr/share/mender/modules/v3/single-file
ADD https://raw.githubusercontent.com/mendersoftware/mender/master/support/modules-artifact-gen/single-file-artifact-gen /usr/bin/single-file-artifact-gen
RUN chmod +x /usr/bin/mender-artifact /usr/bin/single-file-artifact-gen
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY ./workflows/generate_artifact.json /etc/workflows/definitions/generate_artifact.json
COPY ./config.yaml /etc/workflows/config.yaml
COPY --from=builder /go/src/github.com/mendersoftware/create-artifact-worker/create-artifact /usr/bin/
COPY --from=workflows /usr/bin/workflows /usr/bin/
ENTRYPOINT ["/usr/bin/workflows", "--config", "/etc/workflows/config.yaml", "worker"]
