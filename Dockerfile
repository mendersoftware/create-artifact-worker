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
    apk add --no-cache \ 
    ca-certificates \
    xz \
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

RUN wget https://d1b0l86ne08fsf.cloudfront.net/mender-artifact/3.2.1/linux/mender-artifact -O /usr/bin/mender-artifact
RUN chmod +x /usr/bin/mender-artifact

RUN mkdir -p /usr/share/mender/modules/v3 && wget -N -P /usr/share/mender/modules/v3 https://raw.githubusercontent.com/mendersoftware/mender/master/support/modules/single-file

RUN wget https://raw.githubusercontent.com/mendersoftware/mender/master/support/modules-artifact-gen/single-file-artifact-gen -O /usr/bin/single-file-artifact-gen
RUN chmod +x /usr/bin/single-file-artifact-gen

COPY ./config.yaml /etc/workflows
COPY --from=builder /go/src/github.com/mendersoftware/create-artifact-worker/create-artifact /usr/bin
ENTRYPOINT ["/usr/bin/workflows", "--worker", "--config", "/etc/workflows/config.yaml"]
