package gitlab

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	gitlab "github.com/xanzy/go-gitlab"
)

func TestAccGitlabServiceJira_basic(t *testing.T) {
	var service gitlab.JiraService
	rInt := acctest.RandInt()
	jiraResourceName := "gitlab_service_jira.jira"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabServiceDestroy,
		Steps: []resource.TestStep{
			// Create a project and a jira service
			{
				Config: testAccGitlabServiceJiraConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabServiceExists(jiraResourceName, &service),
					resource.TestCheckResourceAttr(jiraResourceName, "url", "https://test.io"),
					resource.TestCheckResourceAttr(jiraResourceName, "username", "test"),
					resource.TestCheckResourceAttr(jiraResourceName, "password", "test"),
					resource.TestCheckResourceAttr(jiraResourceName, "project_key", "2"),
					resource.TestCheckResourceAttr(jiraResourceName, "jira_issue_transition_id", "4"),
				),
			},
			// Update the jira service
			{
				Config: testAccGitlabServiceJiraUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabServiceExists(jiraResourceName, &service),
					resource.TestCheckResourceAttr(jiraResourceName, "url", "https://test-url.io"),
					resource.TestCheckResourceAttr(jiraResourceName, "username", "test_username"),
					resource.TestCheckResourceAttr(jiraResourceName, "password", "test_password"),
					resource.TestCheckResourceAttr(jiraResourceName, "project_key", "3"),
					resource.TestCheckResourceAttr(jiraResourceName, "jira_issue_transition_id", "5"),
				),
			},
			// Update the jira service to get back to initial settings
			{
				Config: testAccGitlabServiceJiraConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabServiceExists(jiraResourceName, &service),
					resource.TestCheckResourceAttr(jiraResourceName, "url", "https://test.io"),
					resource.TestCheckResourceAttr(jiraResourceName, "username", "test"),
					resource.TestCheckResourceAttr(jiraResourceName, "password", "test"),
					resource.TestCheckResourceAttr(jiraResourceName, "project_key", "2"),
					resource.TestCheckResourceAttr(jiraResourceName, "jira_issue_transition_id", "4"),
				),
			},
		},
	})
}

func testAccCheckGitlabServiceExists(n string, service *gitlab.JiraService) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		project := rs.Primary.Attributes["project"]
		if project == "" {
			return fmt.Errorf("No project ID is set")
		}
		conn := testAccProvider.Meta().(*gitlab.Client)

		_, _, err := conn.Services.GetJiraService(project)
		if err != nil {
			return fmt.Errorf("Jira service does not exist in project %s: %v", project, err)
		}

		return nil
	}
}

func testAccCheckGitlabServiceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*gitlab.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "gitlab_project" {
			continue
		}

		gotRepo, resp, err := conn.Projects.GetProject(rs.Primary.ID)
		if err == nil {
			if gotRepo != nil && fmt.Sprintf("%d", gotRepo.ID) == rs.Primary.ID {
				return fmt.Errorf("Repository still exists")
			}
		}
		if resp.StatusCode != 404 {
			return err
		}
		return nil
	}
	return nil
}

func testAccGitlabServiceJiraConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name        = "foo-%d"
  description = "Terraform acceptance tests"
  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_service_jira" "jira" {
	project                  = "${gitlab_project.foo.id}"
	url                      = "https://test.io"
	username                 = "test"
	password                 = "test"
	project_key              = "2"
	jira_issue_transition_id = 4
}
`, rInt)
}

func testAccGitlabServiceJiraUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name        = "foo-%d"
  description = "Terraform acceptance tests"
  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_service_jira" "jira" {
  project                  = "${gitlab_project.foo.id}"
  url                      = "https://test-url.io"
  username                 = "test_username"
  password                 = "test_password"
  project_key              = "3"
  jira_issue_transition_id = 5
}
`, rInt)
}
