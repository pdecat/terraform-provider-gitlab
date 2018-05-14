package gitlab

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceGitlabUser_basic(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataSourceGitlabUserConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.gitlab_user.test",
						"user.id", regexp.MustCompile("^[1-9]*$"))),
			},
		},
	})
}

func testAccDataSourceGitlabUserConfig(rInt int) string {
	return fmt.Sprintf(`
data "gitlab_user" "test" {
  user_id = "${gitlab_user.foo.id}"
}

resource "gitlab_user" "foo" {
  name             = "testdatasource %d"
  username         = "test%d"
  password         = "test%d"
  email            = "test%d@test.com"
  is_admin         = false
  projects_limit   = 0
  can_create_group = false
}
`, rInt, rInt, rInt, rInt)
}
