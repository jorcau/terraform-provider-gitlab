package gitlab

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccGitlabGroupMembers_basic(t *testing.T) {
	resourceName := "gitlab_group_members.foo"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{PreCheck: func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccGitlabGroupMembersConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "members.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "members.0.access_level", "developer"),
				),
			},
			{
				Config: testAccGitlabGroupMembersUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "members.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "members.0.access_level", "guest"),
					resource.TestCheckResourceAttr(resourceName, "members.0.expires_at", "2099-01-01"),
				),
			},
			{
				Config: testAccGitlabGroupMembersConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "members.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "members.0.access_level", "developer"),
					resource.TestCheckResourceAttr(resourceName, "members.0.expires_at", ""),
				),
			},
		},
	})
}

func testAccGitlabGroupMembersConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_group_members" "foo" {
  group_id     = "${gitlab_group.foo.id}"
  members      = [
	{
		id      	 = "${gitlab_user.test.id}"
		access_level = "developer"
		expires_at   = "2099-01-01"
	}
  ]
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

func testAccGitlabGroupMembersUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_group_members" "foo" {
  group_id     = "${gitlab_group.foo.id}"
  members      = [
	  {
		id      = "${gitlab_user.test.id}"
		access_level = "guest"
		expires_at   = "2099-01-01"
	  }
  ]
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
