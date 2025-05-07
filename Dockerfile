ARG WORKFLOWS_VERSION=master
FROM --platform=$BUILDPLATFORM golang:1.24.3 AS builder-create-artifact-worker
ARG TARGETARCH
WORKDIR /go/src/github.com/mendersoftware/create-artifact-worker
COPY ./ .
RUN env CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o create-artifact

FROM mendersoftware/workflows:$WORKFLOWS_VERSION AS workflows

FROM --platform=$BUILDPLATFORM golang:1.24.3 AS builder-mender-artifact
ARG MENDER_ARTIFACT_VERSION=4.1.0
ARG TARGETARCH
RUN git clone \
    --depth 1 \
    --branch $MENDER_ARTIFACT_VERSION \
    https://github.com/mendersoftware/mender-artifact.git \
    /go/src/github.com/mendersoftware/mender-artifact
WORKDIR /go/src/github.com/mendersoftware/mender-artifact
RUN env CGO_ENABLED=0 GOARCH=${TARGETARCH} \
    go build \
    -tags nopkcs11 \
    -ldflags "-X github.com/mendersoftware/mender-artifact/cli.Version=${MENDER_ARTIFACT_VERSION}" \
    -o mender-artifact

FROM alpine:3.18.12
RUN apk add --no-cache \
    xz \
    libc6-compat \
    openssl1.1-compat \
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

RUN sed -i 's/ash/bash/g' /etc/passwd
COPY --from=builder-mender-artifact /go/src/github.com/mendersoftware/mender-artifact/mender-artifact /usr/bin/
ADD https://raw.githubusercontent.com/mendersoftware/mender/master/support/modules-artifact-gen/single-file-artifact-gen /usr/bin/single-file-artifact-gen
RUN chmod +x /usr/bin/mender-artifact /usr/bin/single-file-artifact-gen
COPY --from=builder-create-artifact-worker /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY ./workflows/generate_artifact.json /etc/workflows/definitions/generate_artifact.json
COPY ./config.yaml /etc/workflows/config.yaml
COPY --from=builder-create-artifact-worker /go/src/github.com/mendersoftware/create-artifact-worker/create-artifact /usr/bin/
COPY --from=workflows /usr/bin/workflows /usr/bin/
ENTRYPOINT ["/usr/bin/workflows", "--config", "/etc/workflows/config.yaml", "worker"]
