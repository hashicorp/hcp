# HCP CLI

The `hcp` CLI lets you administer [HashiCorp Cloud Platform (HCP)](https://cloud.hashicorp.com)
resources and services. You can interact with `hcp` directly or integrate it in
scripts to automate your workflows. The HCP CLI is available across most
platforms and lets you efficiently execute common platform tasks and manage HCP
resources at scale.

* Documentation: [https://developer.hashicorp.com/hcp/docs/cli](https://developer.hashicorp.com/hcp/docs/cli)
* GitHub: [github.com/hashicorp/hcp](https://github.com/hashicorp/hcp)


# How to use this image
## Interactively

The image can be used to run the `hcp` CLI interactively:

```
$ docker run -it hashicorp/hcp
# hcp auth login --client-id=<client-id> --client-secret=<client-secret>
# hcp ...
```

## Non-interactively

To run non-interactively, the `hcp` CLI must be passed credentials as part of
its invocation. As such a service principal must be used. There are two ways to
pass the service principal credentials:

1. Using environment variables

    The simplest way to pass the service principal credentials is to set them as environment variables:

    ```
    $ docker run -e HCP_CLIENT_ID=<client-id> -e HCP_CLIENT_SECRET=<client-secret> hashicorp/hcp hcp <command>
    ```

2. Using a credential file.

    The more secure option is to create a credential file containing the service
    principal credentials and mount it into the container. For details on creating a
    credential file, refer to the [documentation](https://developer.hashicorp.com/hcp/docs/cli/commands/iam/service-principals/keys/create)
    or if using a Workload Identity Provider, see the [create-cred-file](https://developer.hashicorp.com/hcp/docs/cli/commands/iam/workload-identity-providers/create-cred-file) command.

    As an example, the following command creates a service principal and a credential file:

    ```
    $ hcp iam service-principals create example-sp
    $ hcp iam service-principals keys create example-sp --output-cred-file=cred_file.json
    ```

    Once created, the credential file can be mounted into the container as follows:

    ```
    $ docker run -v /path/to/cred_file_dir/:/root/hcp/ \
      -e HCP_CRED_FILE=/root/hcp/cred_file.json hashicorp/hcp hcp <command>
    ```

    The `-v` flag mounts the credential file into the container at `/root/hcp/cred_file.json`.

    The `-e` flag sets the `HCP_CRED_FILE` environment variable to the path of the
    mounted credential file and instructs the `hcp` CLI to authenticate via the
    credential file.
