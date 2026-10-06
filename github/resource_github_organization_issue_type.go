package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/go-github/v88/github"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type organizationIssueType struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
	Enabled     bool    `json:"is_enabled"`
}

func resourceGithubOrganizationIssueType() *schema.Resource {
	return &schema.Resource{
		Description: "Manage an issue type in a GitHub organization.",

		CreateContext: resourceGithubOrganizationIssueTypeCreate,
		ReadContext:   resourceGithubOrganizationIssueTypeRead,
		UpdateContext: resourceGithubOrganizationIssueTypeUpdate,
		DeleteContext: resourceGithubOrganizationIssueTypeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"issue_type_id": {
				Description: "The ID of the organization issue type.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"name": {
				Description: "The name of the organization issue type.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The description of the organization issue type.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"color": {
				Description:      "The color of the organization issue type.",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: validateValueFunc([]string{"gray", "blue", "green", "yellow", "orange", "red", "pink", "purple"}),
			},
			"enabled": {
				Description: "Whether the organization issue type is enabled.",
				Type:        schema.TypeBool,
				Required:    true,
			},
		},
	}
}

func resourceGithubOrganizationIssueTypeCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if err := checkOrganization(meta); err != nil {
		return diag.FromErr(err)
	}

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name

	issueType, _, err := client.Organizations.CreateIssueType(ctx, orgName, organizationIssueTypeOptions(d))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating GitHub organization issue type (%s/%s): %w", orgName, d.Get("name").(string), err))
	}

	if err = d.Set("issue_type_id", issueType.GetID()); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(issueType.GetID(), 10))
	return nil
}

func resourceGithubOrganizationIssueTypeRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if err := checkOrganization(meta); err != nil {
		return diag.FromErr(err)
	}

	issueTypeID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name

	issueTypes, err := listOrganizationIssueTypes(ctx, client, orgName)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error querying GitHub organization issue types (%s): %w", orgName, err))
	}

	var issueType *organizationIssueType
	for _, candidate := range issueTypes {
		if candidate.ID == issueTypeID {
			issueType = candidate
			break
		}
	}

	if issueType == nil {
		tflog.Warn(ctx, "GitHub organization issue type not found, removing from state", map[string]any{
			"orgName":     orgName,
			"issueTypeId": issueTypeID,
		})
		d.SetId("")
		return nil
	}

	if err = d.Set("issue_type_id", issueType.ID); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("name", issueType.Name); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("description", issueType.Description); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("color", issueType.Color); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("enabled", issueType.Enabled); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceGithubOrganizationIssueTypeUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if err := checkOrganization(meta); err != nil {
		return diag.FromErr(err)
	}

	issueTypeID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name

	issueType, _, err := client.Organizations.UpdateIssueType(ctx, orgName, issueTypeID, organizationIssueTypeOptions(d))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating GitHub organization issue type (%s/%s): %w", orgName, d.Get("name").(string), err))
	}

	if err = d.Set("issue_type_id", issueType.GetID()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceGithubOrganizationIssueTypeDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if err := checkOrganization(meta); err != nil {
		return diag.FromErr(err)
	}

	issueTypeID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name

	_, err = client.Organizations.DeleteIssueType(ctx, orgName, issueTypeID)
	if err != nil {
		if githubError, ok := errors.AsType[*github.ErrorResponse](err); ok && githubError.Response.StatusCode == http.StatusNotFound {
			return nil
		}

		return diag.FromErr(fmt.Errorf("error deleting GitHub organization issue type (%s/%d): %w", orgName, issueTypeID, err))
	}

	return nil
}

func organizationIssueTypeOptions(d *schema.ResourceData) *github.CreateOrUpdateIssueTypesOptions {
	options := &github.CreateOrUpdateIssueTypesOptions{
		Name:      d.Get("name").(string),
		IsEnabled: d.Get("enabled").(bool),
	}

	if value, ok := d.GetOk("description"); ok {
		options.Description = new(value.(string))
	}
	if value, ok := d.GetOk("color"); ok {
		options.Color = new(value.(string))
	}
	return options
}

func listOrganizationIssueTypes(ctx context.Context, client *github.Client, orgName string) ([]*organizationIssueType, error) {
	request, err := client.NewRequest(ctx, http.MethodGet, fmt.Sprintf("orgs/%s/issue-types", orgName), nil)
	if err != nil {
		return nil, err
	}

	var issueTypes []*organizationIssueType
	if _, err = client.Do(request, &issueTypes); err != nil {
		return nil, err
	}

	return issueTypes, nil
}
