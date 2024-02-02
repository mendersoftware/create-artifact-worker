ARG WORKFLOWS_VERSION=master
FROM --platform=$BUILDPLATFORM golang:1.21.6-alpine3.18 as builder
ARG TARGETARCH
RUN apk add --no-cache \
    ca-certificates \
    musl-dev \
    gcc \
    git
WORKDIR /go/src/github.com/mendersoftware/create-artifact-worker
COPY ./ .
RUN env CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o create-artifact

FROM mendersoftware/workflows:$WORKFLOWS_VERSION as workflows

FROM --platform=$BUILDPLATFORM alpine:3.19.1 as mender-artifact-get
ARG TARGETARCH
ARG MENDER_ARTIFACT_VERSION=3.10.1
RUN apk --update --no-cache add dpkg
RUN deb_filename=mender-artifact_${MENDER_ARTIFACT_VERSION}-1%2Bdebian%2Bbullseye_${TARGETARCH}.deb && \
    wget "https://downloads.mender.io/repos/debian/pool/main/m/mender-artifact/${deb_filename}" \
    --output-document=/mender-artifact.deb && dpkg-deb --extract /mender-artifact.deb /

FROM alpine:3.19.1
ARG MENDER_ARTIFACT_VERSION
ARG TARGETARCH
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
    # bmap-tools not found

RUN sed -i 's/ash/bash/g' /etc/passwd
COPY --from=mender-artifact-get /usr/bin/mender-artifact /usr/bin/mender-artifact
ADD https://raw.githubusercontent.com/mendersoftware/mender/master/support/modules/single-file /usr/share/mender/modules/v3/single-file
ADD https://raw.githubusercontent.com/mendersoftware/mender/master/support/modules-artifact-gen/single-file-artifact-gen /usr/bin/single-file-artifact-gen
RUN chmod +x /usr/bin/mender-artifact /usr/bin/single-file-artifact-gen
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY ./workflows/generate_artifact.json /etc/workflows/definitions/generate_artifact.json
COPY ./config.yaml /etc/workflows/config.yaml
COPY --from=builder /go/src/github.com/mendersoftware/create-artifact-worker/create-artifact /usr/bin/
COPY --from=workflows /usr/bin/workflows /usr/bin/
ENTRYPOINT ["/usr/bin/workflows", "--config", "/etc/workflows/config.yaml", "worker"]
