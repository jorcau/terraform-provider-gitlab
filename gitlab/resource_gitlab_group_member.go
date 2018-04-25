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

func resourceGitlabGroupMemberCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	groupID := d.Get("group_id").(string)
	userID := d.Get("user_id").(int)
	accessLevelID := accessLevelID[strings.ToLower(d.Get("access_level").(string))]

	options := &gitlab.AddGroupMemberOptions{
		UserID:      &userID,
		AccessLevel: &accessLevelID,
	}

	if exp, ok := d.GetOk("expires_at"); ok {
		expiresAt := exp.(string)
		options.ExpiresAt = &expiresAt
	}

	log.Printf("[DEBUG] create gitlab group member %d in %s", userID, groupID)

	groupMember, _, err := client.GroupMembers.AddGroupMember(groupID, options)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%d", groupMember.ID))

	return resourceGitlabGroupMemberRead(d, meta)
}

func resourceGitlabGroupMemberRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	log.Printf("[DEBUG] read gitlab group member %s", d.Id())

	groupID := d.Get("group_id").(string)
	userID := d.Get("user_id").(int)

	groupMember, resp, err := client.GroupMembers.GetGroupMember(groupID, userID)
	if err != nil {
		if resp.StatusCode == 404 {
			log.Printf("[WARN] removing group member %s in %s from state because it no longer exists in gitlab", d.Id(), groupID)
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("access_level", accessLevel[groupMember.AccessLevel])
	if groupMember.ExpiresAt != nil {
		d.Set("expires_at", groupMember.ExpiresAt.String())
	}
	if groupMember.CreatedAt != nil {
		d.Set("created_at", groupMember.CreatedAt.String())
	}
	if groupMember.Email != "" {
		d.Set("email", groupMember.Email)
	}
	d.Set("username", groupMember.Username)
	d.Set("name", groupMember.Name)
	d.Set("state", groupMember.State)

	return nil
}

func resourceGitlabGroupMemberUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	groupID := d.Get("group_id").(string)
	userID := d.Get("user_id").(int)
	accessLevelID := accessLevelID[strings.ToLower(d.Get("access_level").(string))]

	options := &gitlab.EditGroupMemberOptions{
		AccessLevel: &accessLevelID,
	}

	if exp, ok := d.GetOk("expires_at"); ok {
		expiresAt := exp.(string)
		options.ExpiresAt = &expiresAt
	}

	log.Printf("[DEBUG] update gitlab group member %s in %s", d.Id(), groupID)

	_, _, err := client.GroupMembers.EditGroupMember(groupID, userID, options)
	if err != nil {
		return err
	}

	return resourceGitlabGroupMemberRead(d, meta)
}

func resourceGitlabGroupMemberDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	groupID := d.Get("group_id").(string)
	userID := d.Get("user_id").(int)

	log.Printf("[DEBUG] delete gitlab group member %s from %s", d.Id(), groupID)

	_, err := client.GroupMembers.RemoveGroupMember(groupID, userID)
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func resourceGitlabGroupMemberImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Parse ID to get both user_id and group_id - Must be in user_id/group_id format
	ids := strings.Split(d.Id(), "/")
	userID, err := strconv.Atoi(ids[0])
	if err != nil {
		return nil, fmt.Errorf("[ERROR] coulnd't parse user_id during import: %v", err)
	}
	groupID := ids[1]

	d.Set("group_id", groupID)
	d.Set("user_id", userID)
	d.SetId(fmt.Sprintf("%d", userID))

	log.Printf("[DEBUG] import gitlab group member %d from %s", userID, groupID)

	return []*schema.ResourceData{d}, nil
}
