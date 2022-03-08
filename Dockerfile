FROM golang:1.17.8-alpine3.15 as builder
RUN apk add --no-cache \
    musl-dev \
    gcc
WORKDIR /go/src/github.com/mendersoftware/create-artifact-worker
COPY ./ .
RUN env CGO_ENABLED=1 go build -o create-artifact

FROM mendersoftware/workflows:master
RUN apk add --no-cache \ 
    ca-certificates \
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

RUN wget https://downloads.mender.io/mender-artifact/3.5.0/linux/mender-artifact -O /usr/bin/mender-artifact
RUN chmod +x /usr/bin/mender-artifact

RUN mkdir -p /usr/share/mender/modules/v3 && wget -N -P /usr/share/mender/modules/v3 https://raw.githubusercontent.com/mendersoftware/mender/master/support/modules/single-file

RUN wget https://raw.githubusercontent.com/mendersoftware/mender/master/support/modules-artifact-gen/single-file-artifact-gen -O /usr/bin/single-file-artifact-gen
RUN chmod +x /usr/bin/single-file-artifact-gen

COPY ./workflows/generate_artifact.json /etc/workflows/definitions/generate_artifact.json

COPY ./config.yaml /etc/workflows
COPY --from=builder /go/src/github.com/mendersoftware/create-artifact-worker/create-artifact /usr/bin
ENTRYPOINT ["/usr/bin/workflows", "--config", "/etc/workflows/config.yaml", "worker"]
