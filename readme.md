# Hood

Hood is an integration for [Helm for Kubernetes](https://github.com/helm/helm) and [Vault](https://github.com/hashicorp/vault).
The intention was to keep value files Helm charts clean from sensitive data such as credentials or private URI's, so
that they can be attached to the related project in their git repository.

In simple words Hood gets a bunch of Vault paths and keys which it looks reads from Vault and generates a fresh new
Kubernetes secret. It's used as a pre-install hook on some Helm charts, but it can be used as any other hook or even as
a standalone.

Hood authenticates against Vault with a service-account token from Kubernetes, you can find an example of a role binding
in *helm-serviceaccount.yaml* and an example Vault-policy in *helm-policy.hcl*. Remember to set the right paths for the
policy, otherwise Hood is able to read everything under **secrets/**

## Install

Docker users can just call `docker build -t aberfeldy/hood .`.

If you want to build it by source, clone it and run it `go run main.go`.

An example of how you activate the Vault-login via Kubernetes Token can be found in *install.sh*.

## Usage

Hood needs three Env-Variables set by Helm (or docker/bash if run in standalone):

`VAULT_SECRETS` A json representing the secrets Hood should be reading from Vault. See example below.

`VAULT_ADDRESS` The Vault address with trailing slash, e.g. http://vault:8200/

`NAMESPACE` The namespace in which Hood should create the secrets in, e.g. "apps"


## Example

Hood uses a json based configuration which looks like this
```json
{
  "secrets": [
    {
      "name": "<how the secret should be named",
      "path": "<the secret path in vault>",
      "props": [
        "<which keys from the secret should be read>"
      ]
    }
  ]
}

```

So let's say you want a secret called *db* containing **hostname**, **username** and **password** which reside in the
secret secret/production/db/mysql and another one called *mailer* with **inbox** and **outbox** from secret/mailer.
Your json config would look like this:

```json
{
  "secrets": [
    {
      "name": "db",
      "path": "secret/production/db/mysql",
      "props": [
        "hostname",
        "username",
        "password"
      ]
    },
    {
      "name": "mailer",
      "path": "secret/mailer",
      "props": [
        "inbox",
        "outbox"
      ]
    }
  ]
}
```

The resulting secret in the namespace app01 in Kubernetes would look something like that:
```json
{
  "kind": "Secret",
  "apiVersion": "v1",
  "metadata": {
    "name": "db",
    "namespace": "app01"
  },
  "data": {
    "db-hostname": "SomeBase64Value",
    "db-username": "SomeBase64Value",
    "db-password": "SomeBase64Value",
  },
  "type": "opaque"
}
```

## Why 'Hood'?

The german word 'Helm' means helmet in english and if you want to open a vault in secret
(as the pirate you are), you need a 'thief-helmet', a hood.