# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: BUSL-1.1

## Common Base Layer
FROM alpine:3.21 as base

RUN apk --no-cache upgrade && apk --no-cache add \
	bash \
	jq \
	nano

## No patch for vulnerabilities in alpine, downloading fixed versions from edge
## Pinning cURL to remediate CVE-2025-10966
RUN apk --no-cache add 'curl>=8.17.0'
  # Vim temporarily removed as there is no available patch on edge yet, although it is fixed in version > 9.1.0697
  # https://pkgs.alpinelinux.org/packages?name=vim&branch=edge&repo=&arch=&maintainer=
  # https://security.alpinelinux.org/vuln/CVE-2024-43802
  # https://security.alpinelinux.org/vuln/CVE-2024-43790

RUN touch ~/.bashrc

## Dev DOCKERFILE ##
FROM base as dev

COPY bin/hcp /bin/hcp
RUN hcp --autocomplete-install

CMD ["/bin/bash"]

## DOCKERHUB DOCKERFILE ##
FROM base as release

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
      description="The hcp CLI allows interaction with the HashiCorp Cloud Platform using the command-line." \
	  org.opencontainers.image.licenses="MPL-2.0"

COPY LICENSE /usr/share/doc/$NAME/LICENSE.txt
COPY dist/$TARGETOS/$TARGETARCH/$BIN_NAME /bin/
RUN hcp --autocomplete-install

CMD ["/bin/bash"]
