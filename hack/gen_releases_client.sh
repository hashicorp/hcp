#!/bin/sh

# Clean the existing releasesapi client
rm -rf internal/pkg/api/releasesapi/*

spec_path="internal/pkg/api/releasesapi/public.swagger.json"

# Get the OpenAPI spec
spec_url="https://releases.hashicorp.com/docs/api/v1/public.swagger.json"
curl -o $spec_path $spec_url

# Strp the spec of the x-go-type field
cat $spec_path | jq 'walk(if type == "object" then del(.["x-go-package","x-go-name","x-go-type"]) else . end)' > "${spec_path}.tmp" && mv "${spec_path}.tmp" $spec_path

# Generate the client
swagger generate client -f internal/pkg/api/releasesapi/public.swagger.json \
	--target internal/pkg/api/releasesapi \
	-A releasesapi

# Run go mod tidy
go mod tidy

# Remove the spec
rm internal/pkg/api/releasesapi/public.swagger.json
