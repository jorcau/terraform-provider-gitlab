package gitlab

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccGitlabGroupMembers_basic(t *testing.T) {
	resourceName := "gitlab_group_members.test-group-members"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{PreCheck: func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccGitlabGroupMembersConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "members.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "members.1.access_level", "developer"),
				),
			},
			{
				Config: testAccGitlabGroupMembersUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "members.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "members.1.access_level", "guest"),
					resource.TestCheckResourceAttr(resourceName, "members.1.expires_at", "2099-01-01"),
				),
			},
			{
				Config: testAccGitlabGroupMembersConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "members.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "members.1.access_level", "developer"),
					resource.TestCheckResourceAttr(resourceName, "members.1.expires_at", ""),
				),
			},
		},
	})
}

func testAccGitlabGroupMembersConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_group_members" "test-group-members" {
	group_id       = "${gitlab_group.test-group.id}"
	group_owner_id = 2
	access_level   = "developer"

  members  = [
		{
      id           = 2
      access_level = "owner"
		},
    {
      id           = "${gitlab_user.test-user.id}"
    }
  ]
}

resource "gitlab_group" "test-group" {
  name             = "bar-name-%d"
  path             = "bar-path-%d"
  description      = "Terraform acceptance tests - group members"
  visibility_level = "public"
}

resource "gitlab_user" "test-user" {
  name     = "foo%d"
  username = "listest%d"
  password = "test%dtt"
  email    = "listest%d@ssss.com"
}
`, rInt, rInt, rInt, rInt, rInt, rInt)
}

func testAccGitlabGroupMembersUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_group_members" "test-group-members" {
	group_id       = "${gitlab_group.test-group.id}"
	group_owner_id = 2
	access_level   = "guest"

  members  = [
		{
      id           = 2
      access_level = "owner"
		},
    {
      id = "${gitlab_user.test-user.id}"
      expires_at   = "2099-01-01"
    }
  ]
}

resource "gitlab_group" "test-group" {
  name             = "bar-name-%d"
  path             = "bar-path-%d"
  description      = "Terraform acceptance tests - group members"
  visibility_level = "public"
}

resource "gitlab_user" "test-user" {
  name     = "foo%d"
  username = "listest%d"
  password = "test%dtt"
  email    = "listest%d@ssss.com"
}
`, rInt, rInt, rInt, rInt, rInt, rInt)
}
