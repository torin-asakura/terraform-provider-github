package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccGithubOrganizationIssueType(t *testing.T) {
	t.Run("creates updates and imports an organization issue type", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		name := fmt.Sprintf("TFAcc%s", randomID)
		updatedName := fmt.Sprintf("TFAccUpdated%s", randomID)

		config := `
			resource "github_organization_issue_type" "test" {
				name        = "%s"
				description = "%s"
				color       = "%s"
				enabled     = %t
			}
		`

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(config, name, "Initial description", "blue", true),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("github_organization_issue_type.test", tfjsonpath.New("issue_type_id"), knownvalue.NotNull()),
						statecheck.ExpectKnownValue("github_organization_issue_type.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
						statecheck.ExpectKnownValue("github_organization_issue_type.test", tfjsonpath.New("description"), knownvalue.StringExact("Initial description")),
						statecheck.ExpectKnownValue("github_organization_issue_type.test", tfjsonpath.New("color"), knownvalue.StringExact("blue")),
						statecheck.ExpectKnownValue("github_organization_issue_type.test", tfjsonpath.New("enabled"), knownvalue.Bool(true)),
					},
				},
				{
					Config: fmt.Sprintf(config, updatedName, "Updated description", "yellow", false),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("github_organization_issue_type.test", tfjsonpath.New("name"), knownvalue.StringExact(updatedName)),
						statecheck.ExpectKnownValue("github_organization_issue_type.test", tfjsonpath.New("description"), knownvalue.StringExact("Updated description")),
						statecheck.ExpectKnownValue("github_organization_issue_type.test", tfjsonpath.New("color"), knownvalue.StringExact("yellow")),
						statecheck.ExpectKnownValue("github_organization_issue_type.test", tfjsonpath.New("enabled"), knownvalue.Bool(false)),
					},
				},
				{
					ResourceName:      "github_organization_issue_type.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}
