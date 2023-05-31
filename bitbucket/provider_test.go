package bitbucket

import (
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

var testAccProvider *schema.Provider
var testAccProviders map[string]func() (*schema.Provider, error)

func init() {
	workspace := os.Getenv("BITBUCKET_WORKSPACE")
	if workspace == "" {
		os.Setenv("BITBUCKET_WORKSPACE", os.Getenv("BITBUCKET_USERNAME"))
	}

	testAccProvider = Provider()
	testAccProviders = map[string]func() (*schema.Provider, error){
		"bitbucket": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
	}
}

func TestProvider(t *testing.T) {
	err := testAccProvider.InternalValidate()
	assert.NoError(t, err)
}

func testAccPreCheck(t *testing.T) {
	username := os.Getenv("BITBUCKET_USERNAME")
	password := os.Getenv("BITBUCKET_PASSWORD")
	oauthClientId := os.Getenv("BITBUCKET_OAUTH_CLIENT_ID")
	oauthClientSecret := os.Getenv("BITBUCKET_OAUTH_CLIENT_SECRET")
	authMethod := os.Getenv("BITBUCKET_AUTH_METHOD")
	workspace := os.Getenv("BITBUCKET_WORKSPACE")

	if strings.EqualFold(authMethod, "oauth") {
		assert.NotEqual(t, "", oauthClientId, "BITBUCKET_OAUTH_CLIENT_ID must be set for acceptance tests")
		assert.NotEqual(t, "", oauthClientSecret, "BITBUCKET_OAUTH_CLIENT_SECRET must be set for acceptance tests")
	} else {
		assert.NotEqual(t, "", username, "BITBUCKET_USERNAME must be set for acceptance tests")
		assert.NotEqual(t, "", password, "BITBUCKET_PASSWORD must be set for acceptance tests")
	}

	if workspace == "" {
		if username == "" {
			assert.Fail(t, "Either BITBUCKET_WORKSPACE or BITBUCKET_USERNAME have to be set")
		} else {
			os.Setenv("BITBUCKET_WORKSPACE", username)
		}
	}
}
