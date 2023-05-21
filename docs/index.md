# Terraform Provider: Bitbucket Cloud
This is a Terraform provider for managing resources within a Bitbucket Cloud account.

In terms of authentication, you have two options:
## Username and Password
You must use your Bitbucket username (not your email address) and an app password.
Visit here for more information on [app passwords](https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/).

### Example Usage
#### With Embedded Credentials
```hcl
provider "bitbucket" {
  username = "my-username" 
  password = "my-password"
}

resource "bitbucket_xxx" "example" {
  ...
}
```

#### With Environment Variables
Please set the following environment variables:
```shell
BITBUCKET_USERNAME=my-username
BITBUCKET_PASSWORD=my-password
```

```hcl
provider "bitbucket" {}

resource "bitbucket_xxx" "example" {
  ...
}
```

## OAuth 2.0
You must use a [OAuth consumer](https://support.atlassian.com/bitbucket-cloud/docs/use-oauth-on-bitbucket-cloud/) that is marked as private.

### Example Usage
#### With Embedded Credentials
```hcl
provider "bitbucket" {
  oauth_client_id = "Key" 
  oauth_client_secret = "Secret"
}

resource "bitbucket_xxx" "example" {
  ...
}
```

#### With Environment Variables
Please set the following environment variables:
```shell
BITBUCKET_OAUTH_CLIENT_ID=Key
BITBUCKET_OAUTH_CLIENT_SECRET=Secret
```

```hcl
provider "bitbucket" {}

resource "bitbucket_xxx" "example" {
  ...
}
```