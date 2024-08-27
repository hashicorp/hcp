# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: BUSL-1.1

## Dev DOCKERFILE ##
FROM alpine:3.20 as dev

COPY bin/hcp /bin/hcp

RUN apk --no-cache upgrade && apk --no-cache add \
	bash \
	jq \
	nano

## No patch for vulnerabilities in alpine, downloading fixed versions from edge
RUN apk --repository=https://dl-cdn.alpinelinux.org/alpine/edge/main --no-cache add \
  vim=9.1.0678-r0 \
  curl=8.9.1-r1

RUN touch ~/.bashrc && hcp --autocomplete-install

CMD ["/bin/bash"]

## DOCKERHUB DOCKERFILE ##
FROM dev as release

ARG BIN_NAME
ARG NAME=hcp
ARG PRODUCT_VERSION
ARG PRODUCT_REVISION
ARG TARGETOS TARGETARCH

# Additional metadata labels used by container registries, platforms
# and certification scanners.
LABEL name="HCP CLI" \
      vendor="HashiCorp" \
      version=${PRODUCT_VERSION} \
      release=${PRODUCT_REVISION} \
      revision=${PRODUCT_REVISION} \
      summary="The hcp CLI allows interaction with the HashiCorp Cloud Platform." \
      description="The hcp CLI allows interaction with the HashiCorp Cloud Platform using the command-line."

COPY dist/$TARGETOS/$TARGETARCH/$BIN_NAME /bin/

CMD ["/bin/bash"]
