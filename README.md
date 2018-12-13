# Vault Environment Exporter

This tool will take a set of secrets under a Vault path prefix and built out a set of environment variables.

This is based on vaultenvporter which is written in Python. I decided to rewrite this in Go to remove the requirement of a Python runtime and the required tools inside of containers that aren't Python based (i.e. the default nodejs containers based on Alpine Linux).

## How it works

Let's say that we have an example prefix of `secret/services/my-service/production/0` with the following secrets:

```
db/username
db/password
auth_db/username
auth_db/password
...
```

Vaultenvporter will create the following environment variables with the value set to the `value` property of each secret:

```bash
DB_USERNAME=...
DB_PASSWORD=...
AUTH_DB_USERNAME=...
AUTH_DB_PASSWORD=...
```

The tool will convert '/' characters to '_' characters when looking up values. The entire secret path after the prefix will be used as the name of the environment variables.

## Download and Installation

You should be able to download a version of this tool from the releases section of this site. It should be ready to go for Linux based system.

TODO(@ahamilton55: add an example of downloading and installing inside of a Docker container)

## Usage

An example usage of this tool in the Kubernetes environment would be added to whatever runs your app in its container:

```bash
eval $(vaultenvporter-go --auth-method kubernetes --vault-prefix <prefix>)
```

This should be placed as part of the command that the Docker container will run when executing.

You will also want to set the address of your Vault server using the `VAULT_ADDR` environment variable inside of your container.

## Notes

### Alpine Linux

If you're using Alpine Linux for your system you will need to install the `ca-certificates` package inside of your container. If you do not then Vault will be unable to verify the certificate and the program will fail.

```dockerfile
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
```
