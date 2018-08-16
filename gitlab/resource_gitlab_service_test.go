package gitlab

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/xanzy/go-gitlab"
)

func TestAccGitlabService_basic(t *testing.T) {
	var service gitlab.Service
	rInt := acctest.RandInt()
	slackResourceName := "gitlab_service.slack"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabServiceDestroy,
		Steps: []resource.TestStep{
			// Create a project and some services
			{
				Config: testAccGitlabServiceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabServiceExists(slackResourceName, &service),
					resource.TestCheckResourceAttr(slackResourceName, "name", "slack"),
				),
			},
			// Update the services to change the parameters
			{
				Config: testAccGitlabServiceUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabServiceExists(slackResourceName, &service),
					resource.TestCheckResourceAttr(slackResourceName, "name", "slack"),
				),
			},
			// Update the services to get back to initial settings
			{
				Config: testAccGitlabServiceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabServiceExists(slackResourceName, &service),
					resource.TestCheckResourceAttr(slackResourceName, "name", "slack"),
				),
			},
		},
	})
}

func testAccCheckGitlabServiceExists(n string, service *gitlab.Service) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		serviceName := rs.Primary.Attributes["name"]
		project := rs.Primary.Attributes["project"]
		if project == "" {
			return fmt.Errorf("No project ID is set")
		}
		conn := testAccProvider.Meta().(*gitlab.Client)

		var err error
		switch serviceName {
		case "slack":
			_, _, err = conn.Services.GetSlackService(project)
		default:
			return fmt.Errorf("Service name '%s' is invalid", serviceName)
		}
		if err != nil {
			return fmt.Errorf("Service %s does not exist in project %s: %v", serviceName, project, err)
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

func testAccGitlabServiceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name        = "foo-%d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_service" "slack" {
  project = "${gitlab_project.foo.id}"
  name    = "slack"

  properties = {
    "webhook" = "https://test"
  }
}
`, rInt)
}

func testAccGitlabServiceUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name        = "foo-%d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_service" "slack" {
  project = "${gitlab_project.foo.id}"
  name    = "slack"

  properties = {
	"webhook"  = "https://test2"
	"username" = "Terraform test"
	"channel"  = "Terraform channel"
  }
}
`, rInt)
}
