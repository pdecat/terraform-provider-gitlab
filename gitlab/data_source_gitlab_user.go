package gitlab

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	gitlab "github.com/xanzy/go-gitlab"
)

func dataSourceGitlabUser() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGitlabUserRead,

		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:          schema.TypeInt,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"username"},
			},
			"username": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"user_id"},
			},
			"email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_admin": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"can_create_group": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"can_create_project": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"projects_limit": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"external": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"extern_uid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organization": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"two_factor_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceGitlabUserRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	var user *gitlab.User
	var err error

	userIdData, userIdOk := d.GetOk("user_id")
	usernameData, usernameOk := d.GetOk("username")

	if userIdOk {
		// Get user by id
		user, _, err = client.Users.GetUser(userIdData.(int))
		if err != nil {
			return err
		}
	} else if usernameOk {
		listUsersOptions := &gitlab.ListUsersOptions{}
		username := usernameData.(string)
		listUsersOptions.Username = &username

		// Get user by username
		var users []*gitlab.User
		users, _, err = client.Users.ListUsers(listUsersOptions)
		if err != nil {
			return err
		}

		if len(users) == 0 {
			return fmt.Errorf("couldn't find a user matching username: %s", username)
		} else if len(users) != 1 {
			return fmt.Errorf("more than one user found matching username: %s", username)
		}
		user = users[0]
	} else {
		return fmt.Errorf("one and only one of user_id or username must be set")
	}

	d.Set("user_id", user.ID)
	d.Set("username", user.Username)
	d.Set("email", user.Email)
	d.Set("name", user.Name)
	d.Set("is_admin", user.IsAdmin)
	d.Set("can_create_group", user.CanCreateGroup)
	d.Set("can_create_project", user.CanCreateProject)
	d.Set("projects_limit", user.ProjectsLimit)
	d.Set("state", user.State)
	d.Set("external", user.External)
	d.Set("extern_uid", user.ExternUID)
	d.Set("created_at", user.CreatedAt)
	d.Set("organization", user.Organization)
	d.Set("two_factor_enabled", user.TwoFactorEnabled)

	d.SetId(fmt.Sprintf("%d", user.ID))

	return nil
}
