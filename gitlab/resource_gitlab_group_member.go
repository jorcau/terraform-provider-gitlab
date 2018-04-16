package gitlab

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	gitlab "github.com/xanzy/go-gitlab"
)

func resourceGitlabGroupMember() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabGroupMemberCreate,
		Read:   resourceGitlabGroupMemberRead,
		Update: resourceGitlabGroupMemberUpdate,
		Delete: resourceGitlabGroupMemberDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGitlabGroupMemberImport,
		},

		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"access_level": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice(
					[]string{"guest", "reporter", "developer", "master", "owner"}, true),
			},
			"expires_at": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

var accessLevel = map[gitlab.AccessLevelValue]string{
	gitlab.GuestPermissions:     "guest",
	gitlab.ReporterPermissions:  "reporter",
	gitlab.DeveloperPermissions: "developer",
	gitlab.MasterPermissions:    "master",
	gitlab.OwnerPermission:      "owner",
}

func resourceGitlabGroupMemberCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	group_id := d.Get("group_id").(string)
	user_id := d.Get("user_id").(int)
	access_level_id := accessLevelID[strings.ToLower(d.Get("access_level").(string))]

	options := &gitlab.AddGroupMemberOptions{
		UserID:      &user_id,
		AccessLevel: &access_level_id,
	}

	if exp, ok := d.GetOk("expires_at"); ok {
		expires_at := exp.(string)
		options.ExpiresAt = &expires_at
	}

	log.Printf("[DEBUG] create gitlab group member %d in %s", user_id, group_id)

	group_member, _, err := client.GroupMembers.AddGroupMember(group_id, options)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%d", group_member.ID))

	return resourceGitlabGroupMemberRead(d, meta)
}

func resourceGitlabGroupMemberRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	log.Printf("[DEBUG] read gitlab group member %s", d.Id())

	group_id := d.Get("group_id").(string)
	user_id := d.Get("user_id").(int)

	group_member, resp, err := client.GroupMembers.GetGroupMember(group_id, user_id)
	if err != nil {
		if resp.StatusCode == 404 {
			log.Printf("[WARN] removing group member %s in %s from state because it no longer exists in gitlab", d.Id(), group_id)
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("access_level", group_member.AccessLevel)
	if group_member.ExpiresAt != nil {
		d.Set("expires_at", group_member.ExpiresAt.String())
	}
	if group_member.CreatedAt != nil {
		d.Set("created_at", group_member.CreatedAt.String())
	}
	if group_member.Email != "" {
		d.Set("email", group_member.Email)
	}
	d.Set("username", group_member.Username)
	d.Set("name", group_member.Name)
	d.Set("state", group_member.State)

	return nil
}

func resourceGitlabGroupMemberUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	group_id := d.Get("group_id").(string)
	user_id := d.Get("user_id").(int)
	access_level_id := accessLevelID[strings.ToLower(d.Get("access_level").(string))]

	options := &gitlab.EditGroupMemberOptions{
		AccessLevel: &access_level_id,
	}

	if exp, ok := d.GetOk("expires_at"); ok {
		expires_at := exp.(string)
		options.ExpiresAt = &expires_at
	}

	log.Printf("[DEBUG] update gitlab group member %s in %s", d.Id(), group_id)

	_, _, err := client.GroupMembers.EditGroupMember(group_id, user_id, options)
	if err != nil {
		return err
	}

	return resourceGitlabGroupMemberRead(d, meta)
}

func resourceGitlabGroupMemberDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	group_id := d.Get("group_id").(string)
	user_id := d.Get("user_id").(int)

	log.Printf("[DEBUG] delete gitlab group member %s from %s", d.Id(), group_id)

	_, err := client.GroupMembers.RemoveGroupMember(group_id, user_id)
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func resourceGitlabGroupMemberImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Parse ID to get both user_id and group_id - Must be in user_id/group_id format
	ids := strings.Split(d.Id(), "/")
	user_id, err := strconv.Atoi(ids[0])
	if err != nil {
		fmt.Errorf("[ERROR] coulnd't parse user_id during import: %v", err)
	}
	group_id := ids[1]

	d.Set("group_id", group_id)
	d.Set("user_id", user_id)
	d.SetId(fmt.Sprintf("%d", user_id))

	log.Printf("[DEBUG] import gitlab group member %d from %s", user_id, group_id)

	return []*schema.ResourceData{d}, nil
}
