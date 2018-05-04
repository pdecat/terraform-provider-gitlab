---
layout: "gitlab"
page_title: "GitLab: gitlab_user"
sidebar_current: "docs-gitlab-datasource-user"
description: |-
  Get GitLab users' data
---

# gitlab\_user

This resource allows you to get GitLab users' data.
Note your provider will need to be configured with admin-level access for this resource to work.

## Example Usage

```hcl
data "gitlab_user" "example" {
  username = "example"
}
```
```hcl
data "gitlab_user" "example2" {
  user_id = 1
}
```

## Argument Reference

The following arguments are supported. Note that at least one of them must be provided.

* `user_id` - (Optional) The id of the user.

* `username` - (Optional) The username of the user.

## Attributes Reference

The datasource exports the following attributes:

* `user_id` - The unique id assigned to the user by the GitLab server.

* `name` - The name of the user.

* `username` - The username of the user.

* `email` - The e-mail address of the user.

* `is_admin` - Whether to enable administrative priviledges for the user.

* `projects_limit` - Number of projects user can create.

* `can_create_group` - Whether to allow the user to create groups.

* `can_create_project` - Whether to allow the user to create projects.

* `created_at` - The user's creation date. 

* `state` - Whether the user is active or blocked.

* `external` - Whether the user is flagged as external.

* `extern_uid` - The external UID.

* `organization` - The organization name.

* `two_factor_enabled` - Whether two factor authentification is enabled.
