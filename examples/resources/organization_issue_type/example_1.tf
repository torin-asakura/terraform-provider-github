resource "github_organization_issue_type" "example" {
  name        = "Epic"
  description = "A multi-week initiative composed of related tasks"
  color       = "purple"
  enabled     = true
}
