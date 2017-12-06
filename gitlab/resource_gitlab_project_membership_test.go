package gitlab

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/xanzy/go-gitlab"
)

func TestAccGitlabProjectMembership_basic(t *testing.T) {
	var membership gitlab.ProjectMember
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{PreCheck: func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabProjectMembershipDestroy,
		Steps: []resource.TestStep{

			// Assign member to the project as a developer
			{
				Config: testAccGitlabProjectMembershipConfig(rInt),
				Check: resource.ComposeTestCheckFunc(testAccCheckGitlabProjectMembershipExists("gitlab_project_membership.foo", &membership), testAccCheckGitlabProjectMembershipAttributes(&membership, &testAccGitlabProjectMembershipExpectedAttributes{
					access_level: fmt.Sprintf("developer"),
				})),
			},

			// Update the project member to change the access level (use testAccGitlabProjectMembershipUpdateConfig for Config)
			{
				Config: testAccGitlabProjectMembershipUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(testAccCheckGitlabProjectMembershipExists("gitlab_project_membership.foo", &membership), testAccCheckGitlabProjectMembershipAttributes(&membership, &testAccGitlabProjectMembershipExpectedAttributes{
					access_level: fmt.Sprintf("guest"),
				})),
			},

			// Update the project member to change the access level back
			{
				Config: testAccGitlabProjectMembershipConfig(rInt),
				Check: resource.ComposeTestCheckFunc(testAccCheckGitlabProjectMembershipExists("gitlab_project_membership.foo", &membership), testAccCheckGitlabProjectMembershipAttributes(&membership, &testAccGitlabProjectMembershipExpectedAttributes{
					access_level: fmt.Sprintf("developer"),
				})),
			},
		},
	})
}

func testAccCheckGitlabProjectMembershipExists(n string, membership *gitlab.ProjectMember) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		projectID := rs.Primary.Attributes["project_id"]
		if projectID == "" {
			return fmt.Errorf("No project ID is set")
		}

		userID := rs.Primary.Attributes["userID"]
		if userID == "" {
			return fmt.Errorf("No user id is set")
		}

		conn := testAccProvider.Meta().(*gitlab.Client)

		user_ID, err := strconv.Atoi(userID)
		gotProjectMembership, _, err := conn.ProjectMembers.GetProjectMember(projectID, user_ID)
		if err != nil {
			return err
		}

		*membership = *gotProjectMembership
		return nil
	}
}

type testAccGitlabProjectMembershipExpectedAttributes struct {
	access_level string
}

var AccessLevel = map[gitlab.AccessLevelValue]string{
	gitlab.GuestPermissions:     "guest",
	gitlab.ReporterPermissions:  "reporter",
	gitlab.DeveloperPermissions: "developer",
	gitlab.MasterPermissions:    "master",
	gitlab.OwnerPermission:      "owner",
}

func testAccCheckGitlabProjectMembershipAttributes(membership *gitlab.ProjectMember, want *testAccGitlabProjectMembershipExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		access_level_id, ok := AccessLevel[membership.AccessLevel]
		if !ok {
			return fmt.Errorf("Invalid access level '%s'", access_level_id)
		}
		if access_level_id != want.access_level {
			return fmt.Errorf("got access level %s; want %s", access_level_id, want.access_level)
		}
		return nil
	}
}

func testAccCheckGitlabProjectMembershipDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*gitlab.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "gitlab_project_membership" {
			continue
		}

		projectID := rs.Primary.Attributes["project_id"]
		userID := rs.Primary.Attributes["user_id"]

		// GetProjectMember needs int type for userID
		userIDI, err := strconv.Atoi(userID)
		gotMembership, resp, err := conn.ProjectMembers.GetProjectMember(projectID, userIDI)
		if err != nil {
			if gotMembership != nil && fmt.Sprintf("%d", gotMembership.AccessLevel) == rs.Primary.Attributes["access_level"] {
				return fmt.Errorf("Project still has member.")
			}
			return nil
		}

		if resp.StatusCode != 404 {
			return err
		}
		return nil
	}
	return nil
}

func testAccGitlabProjectMembershipConfig(rInt int) string {
	return fmt.Sprintf(`resource "gitlab_project" "foo" {
		name = "foo-%d"
		description = "Terraform acceptance tests"

		# So that acceptance tests can be run in a gitlab organziation with no billing
		visibility_level ="public"
	}

	resource "gitlab_user" "foo" {
		name = "foo %d"
		username = "listest%d"
		password = "test%dtt"
		email = "listest%d@ssss.com"
		is_admin = false
		projects_limit = 20
		can_create_group = false
	}

	resource "gitlab_project_membership" "foo" {
		project_id = "${gitlab_project.foo.id}"
		user_id = "${gitlab_user.foo.id}"
		access_level = "developer"
	}
	`, rInt, rInt, rInt, rInt, rInt)
}

func testAccGitlabProjectMembershipUpdateConfig(rInt int) string {
	return fmt.Sprintf(`resource "gitlab_project" "foo" {
		name = "foo-%d"
		description = Terraform acceptance tests"

		# So that acceptance tests can be run in  a gitlab organization with no billing
		visibility_level ="public"
	}

	resource "gitlab_user" "foo" {
		name = "foo %d"
		username = "listest%d"
		password = "test%dtt"
		email = "listest%d@ssss.com"
		is_admin = false
		projects_limit = 20
		can_creat_group = false
	}

	resource "gitlab_project_membership" "foo" {
		project_id = "${gitlab_project.foo.id}"
		user_id = "${gitlab_user.foo.id}"
		access_level = "guest"
	}
	`, rInt, rInt, rInt, rInt, rInt)
}
