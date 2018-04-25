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
			{
				Config: testAccGitlabGroupMemberConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupMemberExists("gitlab_group_member.foo", &groupMember),
					testAccCheckGitlabGroupMemberAttributes(&groupMember, &testAccGitlabGroupMemberExpectedAttributes{
						AccessLevel: fmt.Sprintf("developer"),
					})),
			},
			{
				Config: testAccGitlabGroupMemberUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(testAccCheckGitlabGroupMemberExists("gitlab_group_member.foo", &groupMember),
					testAccCheckGitlabGroupMemberAttributes(&groupMember, &testAccGitlabGroupMemberExpectedAttributes{
						AccessLevel: fmt.Sprintf("guest"),
						ExpiresAt:   fmt.Sprintf("2099-01-01"),
					})),
			},
			{
				Config: testAccGitlabGroupMemberConfig(rInt),
				Check: resource.ComposeTestCheckFunc(testAccCheckGitlabGroupMemberExists("gitlab_group_member.foo", &groupMember),
					testAccCheckGitlabGroupMemberAttributes(&groupMember, &testAccGitlabGroupMemberExpectedAttributes{
						AccessLevel: fmt.Sprintf("developer"),
					})),
			},
		},
	})
}

func testAccCheckGitlabGroupMemberExists(n string, groupMember *gitlab.GroupMember) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		conn := testAccProvider.Meta().(*gitlab.Client)
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		groupID := rs.Primary.Attributes["group_id"]
		if groupID == "" {
			return fmt.Errorf("No group ID is set")
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

		*groupMember = *gotGroupMember
		return nil
	}
}

type testAccGitlabGroupMemberExpectedAttributes struct {
	AccessLevel string
	ExpiresAt   string
}

func testAccCheckGitlabGroupMemberAttributes(groupMember *gitlab.GroupMember, want *testAccGitlabGroupMemberExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		accessLevelID, ok := accessLevel[groupMember.AccessLevel]
		if !ok {
			return fmt.Errorf("Invalid access level '%s'", accessLevelID)
		}
		if accessLevelID != want.AccessLevel {
			return fmt.Errorf("got access level %s; want %s", accessLevelID, want.AccessLevel)
		}

		if (groupMember.ExpiresAt != nil) && (groupMember.ExpiresAt.String() != want.ExpiresAt) {
			return fmt.Errorf("got expires at %q; want %q", groupMember.ExpiresAt, want.ExpiresAt)
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

		userIDI, err := strconv.Atoi(userID)
		gotMembership, resp, err := conn.GroupMembers.GetGroupMember(groupID, userIDI)
		if err != nil {
			if gotMembership != nil && fmt.Sprintf("%d", gotMembership.AccessLevel) == rs.Primary.Attributes["access_level"] {
				return fmt.Errorf("group still has member")
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
	return fmt.Sprintf(`
resource "gitlab_group_member" "foo" {
  group_id     = "${gitlab_group.foo.id}"
  user_id      = "${gitlab_user.test.id}"
  access_level = "developer"
}

resource "gitlab_group" "foo" {
  name                   = "bar-name-%d"
  path                   = "bar-path-%d"
  description            = "Terraform acceptance tests - group member"
  lfs_enabled            = false
  request_access_enabled = true
  visibility_level       = "public"
}

resource "gitlab_user" "test" {
  name     = "foo%d"
  username = "listest%d"
  password = "test%dtt"
  email    = "listest%d@ssss.com"
}
`, rInt, rInt, rInt, rInt, rInt, rInt)
}

func testAccGitlabGroupMemberUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_group_member" "foo" {
  group_id     = "${gitlab_group.foo.id}"
  user_id      = "${gitlab_user.test.id}"
  access_level = "guest"
  expires_at   = "2099-01-01"
}

resource "gitlab_group" "foo" {
  name                   = "bar-name-%d"
  path                   = "bar-path-%d"
  description            = "Terraform acceptance tests - group member"
  lfs_enabled            = false
  request_access_enabled = true
  visibility_level       = "public"
}

resource "gitlab_user" "test" {
  name     = "foo%d"
  username = "listest%d"
  password = "test%dtt"
  email    = "listest%d@ssss.com"
}
`, rInt, rInt, rInt, rInt, rInt, rInt)
}
