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

func TestAccGitlabGroupMember_basic(t *testing.T) {
	var groupMember gitlab.GroupMember
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{PreCheck: func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabGroupMemberDestroy,
		Steps: []resource.TestStep{

			// Assign member to the project as a developer
			{
				Config: testAccGitlabGroupMemberConfig(rInt),
				Check: resource.ComposeTestCheckFunc(testAccCheckGitlabGroupMemberExists("gitlab_group_member.foo", &groupMember), testAccCheckGitlabGroupMemberAttributes(&groupMember, &testAccGitlabGroupMemberExpectedAttributes{
					access_level: fmt.Sprintf("developer"),
				})),
			},

			// Update the group member to change the access level (use testAccGitlabGroupMemberUpdateConfig for Config)
			{
				Config: testAccGitlabGroupMemberUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(testAccCheckGitlabGroupMemberExists("gitlab_group_member.foo", &groupMember), testAccCheckGitlabGroupMemberAttributes(&groupMember, &testAccGitlabGroupMemberExpectedAttributes{
					access_level: fmt.Sprintf("guest"),
				})),
			},

			// Update the group member to change the access level back
			{
				Config: testAccGitlabGroupMemberConfig(rInt),
				Check: resource.ComposeTestCheckFunc(testAccCheckGitlabGroupMemberExists("gitlab_group_member.foo", &groupMember), testAccCheckGitlabGroupMemberAttributes(&groupMember, &testAccGitlabGroupMemberExpectedAttributes{
					access_level: fmt.Sprintf("developer"),
				})),
			},
		},
	})
}

func testAccCheckGitlabGroupMemberExists(n string, membership *gitlab.GroupMember) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		conn := testAccProvider.Meta().(*gitlab.Client)
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		groupID := rs.Primary.Attributes["group_id"]
		if groupID == "" {
			return fmt.Errorf("No project ID is set")
		}

		userID := rs.Primary.Attributes["user_id"]
		id, _ := strconv.Atoi(userID)
		if userID == "" {
			return fmt.Errorf("No user id is set")
		}

		gotGroupMember, _, err := conn.GroupMembers.GetGroupMember(groupID, id)
		if err != nil {
			return err
		}

		*membership = *gotGroupMember
		return nil
	}
}

type testAccGitlabGroupMemberExpectedAttributes struct {
	access_level string
}

// var accessLevel = map[gitlab.AccessLevelValue]string{
// 	gitlab.GuestPermissions:     "guest",
// 	gitlab.ReporterPermissions:  "reporter",
// 	gitlab.DeveloperPermissions: "developer",
// 	gitlab.MasterPermissions:    "master",
// 	gitlab.OwnerPermission:      "owner",
// }

func testAccCheckGitlabGroupMemberAttributes(membership *gitlab.GroupMember, want *testAccGitlabGroupMemberExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		access_level_id, ok := accessLevel[membership.AccessLevel]
		if !ok {
			return fmt.Errorf("Invalid access level '%s'", access_level_id)
		}
		if access_level_id != want.access_level {
			return fmt.Errorf("got access level %s; want %s", access_level_id, want.access_level)
		}
		return nil
	}
}

func testAccCheckGitlabGroupMemberDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*gitlab.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "gitlab_group_member" {
			continue
		}

		groupID := rs.Primary.Attributes["group_id"]
		userID := rs.Primary.Attributes["user_id"]

		// GetGroupMember needs int type for userID
		userIDI, err := strconv.Atoi(userID)
		gotMembership, resp, err := conn.GroupMembers.GetGroupMember(groupID, userIDI)
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

func testAccGitlabGroupMemberConfig(rInt int) string {
	return fmt.Sprintf(`resource "gitlab_group_member" "foo" {
group_id = "${gitlab_project.foo.id}"
user_id = "${gitlab_user.test.id}"
access_level = "developer"
}

resource "gitlab_project" "foo" {
name = "foo%d"
description = "Terraform acceptance tests"
visibility_level ="public"
}

resource "gitlab_user" "test" {
name = "foo%d"
username = "listest%d"
password = "test%dtt"
email = "listest%d@ssss.com"
}
`, rInt, rInt, rInt, rInt, rInt)
}

func testAccGitlabGroupMemberUpdateConfig(rInt int) string {
	return fmt.Sprintf(`resource "gitlab_group_member" "foo" {
group_id = "${gitlab_project.foo.id}"
user_id = "${gitlab_user.test.id}"
access_level = "guest"
}

resource "gitlab_project" "foo" {
name = "foo%d"
description = "Terraform acceptance tests"
visibility_level ="public"
}

resource "gitlab_user" "test" {
name = "foo%d"
username = "listest%d"
password = "test%dtt"
email = "listest%d@ssss.com"
}
`, rInt, rInt, rInt, rInt, rInt)
}
