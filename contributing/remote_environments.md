# Environments

In order to use the CLI against non-production remote environments the following
environment variables need to be set:

* `HCP_API_ADDRESS`: The address of the HCP API.
* `HCP_AUTH_URL`: The address to retrieve an HCP Access Token.
* `HCP_OAUTH_CLIENT_ID`: The OAuth Client ID used when retrieving an HCP Access Token.

[See here for the values per
environment](https://github.com/hashicorp/cloud-experiences-tooling-docs/blob/main/docs/go-sdk/environments.md).
