package bitbucket

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	gobb "github.com/ktrysmt/go-bitbucket"
)

func resourceBitbucketDeploymentVariable() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBitbucketDeploymentVariableCreate,
		ReadContext:   resourceBitbucketDeploymentVariableRead,
		UpdateContext: resourceBitbucketDeploymentVariableUpdate,
		DeleteContext: resourceBitbucketDeploymentVariableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceBitbucketDeploymentVariableImport,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the deployment variable.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"workspace": {
				Description: "The slug or UUID (including the enclosing `{}`) of the workspace.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"repository": {
				Description:      "The name of the repository (must consist of only lowercase ASCII letters, numbers, underscores and hyphens).",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateRepositoryName,
			},
			"deployment": {
				Description: "The UUID (including the enclosing `{}`) of the deployment environment.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"key": {
				Description:      "The name of the variable (must consist of only ASCII letters, numbers, underscores & not begin with a number).",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateRepositoryVariableName,
			},
			"value": {
				Description: "The value of the variable.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"secured": {
				Description: "Whether this variable is considered secure/sensitive. If true, then it's value will not be exposed in any logs or API requests.",
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
			},
		},
	}
}

func resourceBitbucketDeploymentVariableCreate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Clients).V2

	deploymentVariable, err := client.Repositories.Repository.AddDeploymentVariable(
		&gobb.RepositoryDeploymentVariableOptions{
			Owner:       resourceData.Get("workspace").(string),
			RepoSlug:    resourceData.Get("repository").(string),
			Environment: &gobb.Environment{Uuid: resourceData.Get("deployment").(string)},
			Key:         resourceData.Get("key").(string),
			Value:       resourceData.Get("value").(string),
			Secured:     resourceData.Get("secured").(bool),
		},
	)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to create deployment variable with error: %s", err))
	}

	resourceData.SetId(deploymentVariable.Uuid)

	return resourceBitbucketDeploymentVariableRead(ctx, resourceData, meta)
}

func resourceBitbucketDeploymentVariableRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Clients).V2

	// Artificial sleep due to Bitbucket's API taking time to return newly created variables :(
	time.Sleep(7 * time.Second)

	var deploymentVariables []gobb.DeploymentVariable
	var newDeploymentVariables *gobb.DeploymentVariables
	var page int = 1
	var err error

	for {
		newDeploymentVariables, err = client.Repositories.Repository.ListDeploymentVariables(
			&gobb.RepositoryDeploymentVariablesOptions{
				Owner:       resourceData.Get("workspace").(string),
				RepoSlug:    resourceData.Get("repository").(string),
				Environment: &gobb.Environment{Uuid: resourceData.Get("deployment").(string)},
				Pagelen:     100, // Bitbucket's API doesn't support querying, so we have to get as many variables as possible in one go and loop over :(
				PageNum:     page,
			},
		)
		if err != nil {
			return diag.FromErr(fmt.Errorf("unable to get deployment variable with error: %s", err))
		}

		if len(newDeploymentVariables.Variables) == 0 {
			break
		}

		deploymentVariables = append(deploymentVariables, newDeploymentVariables.Variables...)
		page++
	}

	var deploymentVariable *gobb.DeploymentVariable
	for _, deploymentVar := range deploymentVariables {
		if deploymentVar.Key == resourceData.Get("key").(string) {
			deploymentVariable = &deploymentVar
			break
		}
	}

	if deploymentVariable == nil {
		return diag.FromErr(errors.New("unable to get deployment variable, Bitbucket API did not return it"))
	}

	_ = resourceData.Set("key", deploymentVariable.Key)

	if !deploymentVariable.Secured {
		_ = resourceData.Set("value", deploymentVariable.Value)
	} else {
		_ = resourceData.Set("value", resourceData.Get("value").(string))
	}

	_ = resourceData.Set("secured", deploymentVariable.Secured)

	resourceData.SetId(deploymentVariable.Uuid)

	return nil
}

func resourceBitbucketDeploymentVariableUpdate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Clients).V2

	_, err := client.Repositories.Repository.UpdateDeploymentVariable(
		&gobb.RepositoryDeploymentVariableOptions{
			Owner:       resourceData.Get("workspace").(string),
			RepoSlug:    resourceData.Get("repository").(string),
			Environment: &gobb.Environment{Uuid: resourceData.Get("deployment").(string)},
			Uuid:        resourceData.Id(),
			Key:         resourceData.Get("key").(string),
			Value:       resourceData.Get("value").(string),
			Secured:     resourceData.Get("secured").(bool),
		},
	)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to update deployment variable with error: %s", err))
	}

	return resourceBitbucketDeploymentVariableRead(ctx, resourceData, meta)
}

func resourceBitbucketDeploymentVariableDelete(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Clients).V2

	_, err := client.Repositories.Repository.DeleteDeploymentVariable(
		&gobb.RepositoryDeploymentVariableDeleteOptions{
			Owner:       resourceData.Get("workspace").(string),
			RepoSlug:    resourceData.Get("repository").(string),
			Environment: &gobb.Environment{Uuid: resourceData.Get("deployment").(string)},
			Uuid:        resourceData.Id(),
		},
	)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to delete deployment variable with error: %s", err))
	}

	resourceData.SetId("")

	return nil
}

func resourceBitbucketDeploymentVariableImport(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	ret := []*schema.ResourceData{resourceData}

	splitID := strings.Split(resourceData.Id(), "/")
	if len(splitID) < 3 {
		return ret, fmt.Errorf("invalid import ID. It must to be in this format \"<workspace-slug|workspace-uuid>/<repository-slug|repository-uuid>/<deployment-variable-uuid>\"")
	}

	_ = resourceData.Set("workspace", splitID[0])
	_ = resourceData.Set("repository", splitID[1])
	_ = resourceData.Set("id", splitID[2])

	_ = resourceBitbucketDeploymentVariableRead(ctx, resourceData, meta)

	return ret, nil
}
