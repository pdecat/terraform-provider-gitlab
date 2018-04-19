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
				ConflictsWith: []string{"username", "email"},
			},
			"username": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"user_id", "email"},
			},
			"email": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"username", "user_id"},
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
		},
	}
}

func dataSourceGitlabUserRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	var user *gitlab.User
	var err error

	// Get user by id
	if data, ok := d.GetOk("user_id"); ok {
		userId := data.(int)
		user, _, err = client.Users.GetUser(userId)
	} else {
		// Get user by username or email using search parameter
		listUsersOptions := &gitlab.ListUsersOptions{}
		email := d.Get("email").(string)
		username := d.Get("username").(string)

		if username == "" && email == "" {
			return fmt.Errorf("one and only one of id, username or email must be set")
		}
		if username != "" {
			listUsersOptions.Search = &username
		} else {
			listUsersOptions.Search = &email
		}

		var users []*gitlab.User
		users, _, err = client.Users.ListUsers(listUsersOptions)
		user = users[0]
	}
	if err != nil {
		return err
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

	d.SetId(fmt.Sprintf("%d", user.ID))
	return nil
}
