package gitlab

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	gitlab "github.com/xanzy/go-gitlab"
)

func resourceGitlabProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabProjectCreate,
		Read:   resourceGitlabProjectRead,
		Update: resourceGitlabProjectUpdate,
		Delete: resourceGitlabProjectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"path": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "" {
						return true
					}
					return old == new
				},
			},
			"namespace_id": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"default_branch": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"issues_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"merge_requests_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"wiki_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"snippets_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"visibility_level": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"private", "internal", "public"}, true),
				Default:      "private",
			},

			"ssh_url_to_repo": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"http_url_to_repo": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"web_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"shared_with_groups": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"group_access": {
							Type:     schema.TypeString,
							Required: true,
						},
						"expires_at": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"group_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceGitlabProjectSetToState(d *schema.ResourceData, project *gitlab.Project) {
	d.SetId(fmt.Sprintf("%d", project.ID))
	d.Set("name", project.Name)
	d.Set("path", project.Path)
	d.Set("description", project.Description)
	d.Set("default_branch", project.DefaultBranch)
	d.Set("issues_enabled", project.IssuesEnabled)
	d.Set("merge_requests_enabled", project.MergeRequestsEnabled)
	d.Set("wiki_enabled", project.WikiEnabled)
	d.Set("snippets_enabled", project.SnippetsEnabled)
	d.Set("visibility_level", string(project.Visibility))
	d.Set("namespace_id", project.Namespace.ID)

	d.Set("ssh_url_to_repo", project.SSHURLToRepo)
	d.Set("http_url_to_repo", project.HTTPURLToRepo)
	d.Set("web_url", project.WebURL)
	d.Set("shared_with_groups", flattenSharedWithGroupsOptions(project))
}

func resourceGitlabProjectCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	options := &gitlab.CreateProjectOptions{
		Name:                 gitlab.String(d.Get("name").(string)),
		IssuesEnabled:        gitlab.Bool(d.Get("issues_enabled").(bool)),
		MergeRequestsEnabled: gitlab.Bool(d.Get("merge_requests_enabled").(bool)),
		WikiEnabled:          gitlab.Bool(d.Get("wiki_enabled").(bool)),
		SnippetsEnabled:      gitlab.Bool(d.Get("snippets_enabled").(bool)),
		Visibility:           stringToVisibilityLevel(d.Get("visibility_level").(string)),
	}

	if v, ok := d.GetOk("path"); ok {
		options.Path = gitlab.String(v.(string))
	}

	if v, ok := d.GetOk("namespace_id"); ok {
		options.NamespaceID = gitlab.Int(v.(int))
	}

	if v, ok := d.GetOk("description"); ok {
		options.Description = gitlab.String(v.(string))
	}

	log.Printf("[DEBUG] create gitlab project %q", options.Name)

	project, _, err := client.Projects.CreateProject(options)
	if err != nil {
		return err
	}

	if v, ok := d.GetOk("shared_with_groups"); ok {
		options := expandSharedWithGroupsOptions(v.([]interface{}))

		for _, option := range options {
			log.Printf("[DEBUG] project shared with group %v", option)
			_, err := client.Projects.ShareProjectWithGroup(project.ID, option)
			if err != nil {
				return err
			}
		}
	}

	d.SetId(fmt.Sprintf("%d", project.ID))

	return resourceGitlabProjectRead(d, meta)
}

func resourceGitlabProjectRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	log.Printf("[DEBUG] read gitlab project %s", d.Id())

	project, response, err := client.Projects.GetProject(d.Id())
	if err != nil {
		if response.StatusCode == 404 {
			log.Printf("[WARN] removing project %s from state because it no longer exists in gitlab", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	resourceGitlabProjectSetToState(d, project)
	return nil
}

func resourceGitlabProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	options := &gitlab.EditProjectOptions{}

	if d.HasChange("name") {
		options.Name = gitlab.String(d.Get("name").(string))
	}

	if d.HasChange("path") && (d.Get("path").(string) != "") {
		options.Path = gitlab.String(d.Get("path").(string))
	}

	if d.HasChange("description") {
		options.Description = gitlab.String(d.Get("description").(string))
	}

	if d.HasChange("default_branch") {
		options.DefaultBranch = gitlab.String(d.Get("default_branch").(string))
	}

	if d.HasChange("visibility_level") {
		options.Visibility = stringToVisibilityLevel(d.Get("visibility_level").(string))
	}

	if d.HasChange("issues_enabled") {
		options.IssuesEnabled = gitlab.Bool(d.Get("issues_enabled").(bool))
	}

	if d.HasChange("merge_requests_enabled") {
		options.MergeRequestsEnabled = gitlab.Bool(d.Get("merge_requests_enabled").(bool))
	}

	if d.HasChange("wiki_enabled") {
		options.WikiEnabled = gitlab.Bool(d.Get("wiki_enabled").(bool))
	}

	if d.HasChange("snippets_enabled") {
		options.SnippetsEnabled = gitlab.Bool(d.Get("snippets_enabled").(bool))
	}

	if *options != (gitlab.EditProjectOptions{}) {
		log.Printf("[DEBUG] update gitlab project %s", d.Id())
		_, _, err := client.Projects.EditProject(d.Id(), options)
		if err != nil {
			return err
		}
	}

	if d.HasChange("shared_with_groups") {
		updateSharedWithGroups(d, meta)
	}

	return resourceGitlabProjectRead(d, meta)
}

func resourceGitlabProjectDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	log.Printf("[DEBUG] Delete gitlab project %s", d.Id())

	_, err := client.Projects.DeleteProject(d.Id())
	if err != nil {
		return err
	}

	// Wait for the project to be deleted.
	// Deleting a project in gitlab is async.
	stateConf := &resource.StateChangeConf{
		Pending: []string{"Deleting"},
		Target:  []string{"Deleted"},
		Refresh: func() (interface{}, string, error) {
			out, response, err := client.Projects.GetProject(d.Id())
			if err != nil {
				if response.StatusCode == 404 {
					return out, "Deleted", nil
				}
				log.Printf("[ERROR] Received error: %#v", err)
				return out, "Error", err
			}
			return out, "Deleting", nil
		},

		Timeout:    10 * time.Minute,
		MinTimeout: 3 * time.Second,
		Delay:      5 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for project (%s) to become deleted: %s", d.Id(), err)
	}
	return nil
}

func expandSharedWithGroupsOptions(d []interface{}) []*gitlab.ShareWithGroupOptions {
	shareWithGroupOptionsList := []*gitlab.ShareWithGroupOptions{}

	for _, config := range d {
		data := config.(map[string]interface{})

		groupAccess := accessLevelID[data["group_access"].(string)]

		shareWithGroupOptions := &gitlab.ShareWithGroupOptions{
			GroupID:     gitlab.Int(data["group_id"].(int)),
			GroupAccess: &groupAccess,
			ExpiresAt:   gitlab.String(data["expires_at"].(string)),
		}

		shareWithGroupOptionsList = append(shareWithGroupOptionsList,
			shareWithGroupOptions)
	}

	return shareWithGroupOptionsList
}

func flattenSharedWithGroupsOptions(project *gitlab.Project) []interface{} {
	sharedWithGroups := project.SharedWithGroups
	sharedWithGroupsList := []interface{}{}

	for _, option := range sharedWithGroups {
		values := map[string]interface{}{
			"group_id":     option.GroupID,
			"group_access": accessLevel[gitlab.AccessLevelValue(option.GroupAccessLevel)],
			"group_name":   option.GroupName,
		}

		sharedWithGroupsList = append(sharedWithGroupsList, values)
	}

	return sharedWithGroupsList
}

func updateSharedWithGroups(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	// Get existing groups and unshare all of them
	project, _, err := client.Projects.GetProject(d.Id())
	if err != nil {
		return err
	}
	for _, group := range project.SharedWithGroups {
		log.Printf("[DEBUG] update shared_with_group: unshare %v", group)
		_, err := client.Projects.DeleteSharedProjectFromGroup(d.Id(), group.GroupID)
		if err != nil {
			return err
		}
	}

	// Share project with groups from the new config
	shareWithGroupOptions := expandSharedWithGroupsOptions(d.Get("shared_with_groups").([]interface{}))
	for _, group := range shareWithGroupOptions {
		log.Printf("[DEBUG] update shared_with_group: share %v", group)
		_, err := client.Projects.ShareProjectWithGroup(d.Id(), group)
		if err != nil {
			return err
		}
	}

	return nil
}
