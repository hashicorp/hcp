# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: BUSL-1.1

## Dev DOCKERFILE ##
FROM alpine:3.19 as dev
COPY bin/hcp /bin/hcp
RUN apk --no-cache upgrade && apk --no-cache add \
	bash \
	curl \
	jq \
	nano \
	vim
RUN touch ~/.bashrc && hcp --autocomplete-install
CMD ["/bin/bash"]

## DOCKERHUB DOCKERFILE ##
FROM alpine:3.19 as default

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
RUN apk --no-cache upgrade && apk --no-cache add \
	bash \
	curl \
	jq \
	nano \
	vim
RUN touch ~/.bashrc && hcp --autocomplete-install
CMD ["/bin/bash"]
