# Vault Environment Exporter

This tool will take a set of secrets under a Vault path prefix and built out a set of environment variables.

This is based on vaultenvporter which is written in Python. I decided to rewrite this in Go to remove the requirement of a Python runtime and the required tools inside of containers that aren't Python based (i.e. the default nodejs containers based on Alpine Linux).

## Download and Installation

You should be able to download a version of this tool from the releases section of this site. It should be ready to go for Linux based system.

## Usage

An example usage of this tool in the Kubernetes environment would be added to whatever runs your app in its container:

``` bash
eval $(vaultenvporter-go --auth-method kubernetes --vault-prefix <prefix>)
```