package gitlab

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	gitlab "github.com/xanzy/go-gitlab"
)

func TestAccGitlabServiceSlack_basic(t *testing.T) {
	var service gitlab.SlackService
	rInt := acctest.RandInt()
	slackResourceName := "gitlab_service_slack.slack"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabServiceDestroy,
		Steps: []resource.TestStep{
			// Create a project and a slack service
			{
				Config: testAccGitlabServiceSlackConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabServiceExists(slackResourceName, &service),
					resource.TestCheckResourceAttr(slackResourceName, "webhook", "https://test.com"),
					resource.TestCheckResourceAttr(slackResourceName, "push_events", "true"),
					resource.TestCheckResourceAttr(slackResourceName, "push_channel", "test"),
				),
			},
			// Update the slack service
			{
				Config: testAccGitlabServiceSlackUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabServiceExists(slackResourceName, &service),
					resource.TestCheckResourceAttr(slackResourceName, "webhook", "https://testupdate.com"),
					resource.TestCheckResourceAttr(slackResourceName, "push_events", "false"),
					resource.TestCheckResourceAttr(slackResourceName, "push_channel", "testupdate"),
				),
			},
			// Update the slack service to get back to initial settings
			{
				Config: testAccGitlabServiceSlackConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabServiceExists(slackResourceName, &service),
					resource.TestCheckResourceAttr(slackResourceName, "webhook", "https://test.com"),
					resource.TestCheckResourceAttr(slackResourceName, "push_events", "true"),
					resource.TestCheckResourceAttr(slackResourceName, "push_channel", "test"),
				),
			},
		},
	})
}

func testAccCheckGitlabServiceExists(n string, service *gitlab.SlackService) resource.TestCheckFunc {
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

		_, _, err := conn.Services.GetSlackService(project)
		if err != nil {
			return fmt.Errorf("Slack service does not exist in project %s: %v", project, err)
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

func testAccGitlabServiceSlackConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name        = "foo-%d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_service_slack" "slack" {
  project                    = "${gitlab_project.foo.id}"
  webhook                    = "https://test.com"
  username                   = "test"
  channel                    = "test"
  push_events                = true
  push_channel               = "test"
  issues_events              = true
  issue_channel              = "test"
  confidential_issues_events = true
  confidential_issue_channel = "test"
  confidential_note_events   = true
  merge_requests_events      = true
  merge_request_channel      = "test"
  tag_push_events            = true
  tag_push_channel           = "test"
  note_events                = true
  note_channel               = "test"
  pipeline_events            = true
  pipeline_channel           = "test"
  wiki_page_events           = true
  wiki_page_channel          = "test"
  job_events                 = true
}
`, rInt)
}

func testAccGitlabServiceSlackUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name        = "foo-%d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_service_slack" "slack" {
    project                    = "${gitlab_project.foo.id}"
    webhook                    = "https://testupdate.com"
    username                   = "test"
    channel                    = "test"
    push_events                = false
    push_channel               = "testupdate"
    issues_events              = false
    issue_channel              = "test"
    confidential_issues_events = true
    confidential_issue_channel = "test"
    confidential_note_events   = true
    merge_requests_events      = false
    merge_request_channel      = "test"
    tag_push_events            = true
    tag_push_channel           = "testupdate"
    note_events                = false
    note_channel               = "test"
    pipeline_events            = true
    pipeline_channel           = "test"
    wiki_page_events           = true
    wiki_page_channel          = "testupdate"
    job_events                 = true
  }
`, rInt)
}
