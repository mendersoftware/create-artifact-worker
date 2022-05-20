ARG WORKFLOWS_VERSION=master
ARG MENDER_ARTIFACT_VERSION=3.7.1

FROM golang:1.18.2-alpine3.15 as builder
RUN apk add --no-cache \
    ca-certificates \
    musl-dev \
    gcc \
    git
WORKDIR /go/src/github.com/mendersoftware/create-artifact-worker
COPY ./ .
RUN env CGO_ENABLED=0 go build -o create-artifact

FROM mendersoftware/workflows:$WORKFLOWS_VERSION as workflows

FROM alpine:3.15.4
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
USER 65534
ADD --chown=nobody https://downloads.mender.io/mender-artifact/$MENDER_ARTIFACT_VERSION/linux/mender-artifact /usr/bin/mender-artifact
ADD --chown=nobody https://raw.githubusercontent.com/mendersoftware/mender/master/support/modules/single-file /usr/share/mender/modules/v3/single-file
ADD --chown=nobody https://raw.githubusercontent.com/mendersoftware/mender/master/support/modules-artifact-gen/single-file-artifact-gen /usr/bin/single-file-artifact-gen
RUN chmod +x /usr/bin/mender-artifact /usr/bin/single-file-artifact-gen
COPY --from=builder --chown=nobody /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --chown=nobody ./workflows/generate_artifact.json /etc/workflows/definitions/generate_artifact.json
COPY --chown=nobody ./config.yaml /etc/workflows/config.yaml
COPY --from=builder --chown=nobody /go/src/github.com/mendersoftware/create-artifact-worker/create-artifact /usr/bin/
COPY --from=workflows --chown=nobody /usr/bin/workflows /usr/bin/
ENTRYPOINT ["/usr/bin/workflows", "--config", "/etc/workflows/config.yaml", "worker"]
