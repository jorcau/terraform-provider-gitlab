package gitlab

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	gitlab "github.com/xanzy/go-gitlab"
)

func resourceGitlabGroupMembers() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabGroupMembersCreate,
		Read:   resourceGitlabGroupMembersRead,
		Update: resourceGitlabGroupMembersUpdate,
		Delete: resourceGitlabGroupMembersDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"members": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
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
				},
				Required: true,
			},
		},
	}
}

type groupMemberAllOptions struct {
	addOption  *gitlab.AddGroupMemberOptions
	editOption *gitlab.EditGroupMemberOptions
}

func resourceGitlabGroupMembersCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	groupID := d.Get("group_id").(string)
	groupMembersOptions := expandGitlabAddGroupMembersOptions(d.Get("members").([]interface{}))

	for _, groupMemberOptions := range groupMembersOptions {
		log.Printf("[DEBUG] create gitlab group member %d in %s", groupMemberOptions.UserID, groupID)

		_, _, err := client.GroupMembers.AddGroupMember(groupID, groupMemberOptions)
		if err != nil {
			return err
		}
	}

	d.SetId(groupID)

	return resourceGitlabGroupMembersRead(d, meta)
}

func resourceGitlabGroupMembersRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	log.Printf("[DEBUG] read group members from group %s", d.Id())

	groupMembers, resp, err := client.Groups.ListGroupMembers(d.Id(), nil)
	if err != nil {
		if resp.StatusCode == 404 {
			d.SetId("")
			return fmt.Errorf("[WARN] removing all group members in %s from state because group no longer exists in gitlab", d.Id())
		}
		return err
	}

	d.Set("members", flattenGitlabGroupMembers(groupMembers))
	d.Set("group_id", d.Id())

	return nil
}

func resourceGitlabGroupMembersUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	groupID := d.Get("group_id")
	oldMembers, resp, err := client.Groups.ListGroupMembers(groupID, nil)
	if err != nil {
		if resp.StatusCode == 404 {
			d.SetId("")
			return fmt.Errorf("[WARN] removing all group members in %s from state because group no longer exists in gitlab", groupID)
		}
		return err
	}

	newMembers := expandGitlabAddGroupMembersOptions(d.Get("members").([]interface{}))

	groupMembersToAdd, groupMembersToUpdate, groupMemberToDelete := getGroupMembersUpdates(newMembers, oldMembers)

	// Create new group members
	for _, groupMember := range groupMembersToAdd {
		log.Printf("[DEBUG] create gitlab group member %d in %s", groupMember.addOption.UserID, groupID)

		_, _, err := client.GroupMembers.AddGroupMember(groupID, groupMember.addOption)
		if err != nil {
			return err
		}
	}

	// Update existing group members
	for _, groupMember := range groupMembersToUpdate {
		log.Printf("[DEBUG] update gitlab group member %d in %s", groupMember.addOption.UserID, groupID)

		_, _, err := client.GroupMembers.EditGroupMember(groupID, *groupMember.addOption.UserID, groupMember.editOption)
		if err != nil {
			return err
		}
	}

	// Remove group members not present in tf config
	for _, groupMember := range groupMemberToDelete {
		log.Printf("[DEBUG] delete group member %d from %s", groupMember.ID, groupID)

		_, err := client.GroupMembers.RemoveGroupMember(groupID, groupMember.ID)
		if err != nil {
			return err
		}
	}

	return resourceGitlabGroupMembersRead(d, meta)
}

func resourceGitlabGroupMembersDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	groupID := d.Get("group_id").(string)
	groupMembers := expandGitlabAddGroupMembersOptions(d.Get("members").([]interface{}))

	for _, groupMember := range groupMembers {
		log.Printf("[DEBUG] delete group member %d from %s", groupMember.UserID, groupID)

		_, err := client.GroupMembers.RemoveGroupMember(groupID, *groupMember.UserID)
		if err != nil {
			return err
		}
	}

	d.SetId("")

	return nil
}

func expandGitlabAddGroupMembersOptions(d []interface{}) []*gitlab.AddGroupMemberOptions {
	groupMemberOptions := []*gitlab.AddGroupMemberOptions{}

	for _, config := range d {
		data := config.(map[string]interface{})
		userID := data["id"].(int)
		accessLevel := accessLevelID[strings.ToLower(data["access_level"].(string))]
		expiresAt := data["expires_at"].(string)

		groupMemberOption := &gitlab.AddGroupMemberOptions{
			UserID:      &userID,
			AccessLevel: &accessLevel,
			ExpiresAt:   &expiresAt,
		}

		groupMemberOptions = append(groupMemberOptions, groupMemberOption)
	}

	return groupMemberOptions
}

func expandGitlabEditGroupMembersOptions(d []interface{}) *[]gitlab.EditGroupMemberOptions {
	groupMemberOptions := &[]gitlab.EditGroupMemberOptions{}

	for _, config := range d {
		data := config.(map[string]interface{})
		accessLevel := accessLevelID[strings.ToLower(data["access_level"].(string))]
		expiresAt := data["expires_at"].(string)

		groupMemberOption := gitlab.EditGroupMemberOptions{
			AccessLevel: &accessLevel,
			ExpiresAt:   &expiresAt,
		}

		*groupMemberOptions = append(*groupMemberOptions, groupMemberOption)
	}

	return groupMemberOptions
}

func findGroupMember(id int, groupMembers []*gitlab.GroupMember) (gitlab.GroupMember, int, error) {
	for i, groupMember := range groupMembers {
		if groupMember.ID == id {
			return *groupMember, i, nil
		}
	}

	return gitlab.GroupMember{}, 0, fmt.Errorf("Couldn't find group member: %d", id)
}

func getGroupMembersUpdates(newMembers []*gitlab.AddGroupMemberOptions,
	oldMembers []*gitlab.GroupMember) ([]*groupMemberAllOptions,
	[]*groupMemberAllOptions, []*gitlab.GroupMember) {
	groupMembersToUpdate := []*groupMemberAllOptions{}
	groupMembersToAdd := []*groupMemberAllOptions{}

	// Iterate through all members in tf config
	for _, newMember := range newMembers {
		newGroupMemberOptions := &groupMemberAllOptions{newMember,
			&gitlab.EditGroupMemberOptions{
				AccessLevel: newMember.AccessLevel,
				ExpiresAt:   newMember.ExpiresAt,
			}}

		// Check if member in tf config already exists on gitlab
		oldMember, index, err := findGroupMember(*newMember.UserID, oldMembers)

		// If it doesn't exist it must be added
		if err != nil {
			groupMembersToAdd = append(groupMembersToAdd, newGroupMemberOptions)
			continue
		}

		// If it exists but there's a change, it must be updated
		if (*newMember.AccessLevel != oldMember.AccessLevel) ||
			(oldMember.ExpiresAt != nil && (*newMember.ExpiresAt != oldMember.ExpiresAt.String()) ||
				(oldMember.ExpiresAt == nil && *newMember.ExpiresAt != "")) {
			groupMembersToUpdate = append(groupMembersToUpdate, newGroupMemberOptions)
		}

		// Remove existing member from existing members list
		oldMembers = append(oldMembers[:index], oldMembers[index+1:]...)
	}

	// Members still present in existing members list must be removed

	return groupMembersToAdd, groupMembersToUpdate, oldMembers
}

func flattenGitlabGroupMembers(groupMembers []*gitlab.GroupMember) []interface{} {
	groupMembersList := []interface{}{}

	for _, groupMember := range groupMembers {
		values := map[string]interface{}{
			"id":           groupMember.ID,
			"access_level": accessLevel[groupMember.AccessLevel],
			"username":     groupMember.Username,
			"name":         groupMember.Name,
			"state":        groupMember.State,
		}

		if groupMember.ExpiresAt != nil {
			values["expires_at"] = groupMember.ExpiresAt.String()
		}
		if groupMember.CreatedAt != nil {
			values["created_at"] = groupMember.CreatedAt.String()
		}
		if groupMember.Email != "" {
			values["email"] = groupMember.Email
		}

		groupMembersList = append(groupMembersList, values)
	}

	// Sort by id in ascendant order
	sort.Slice(groupMembersList, func(i, j int) bool {
		return groupMembersList[i].(map[string]interface{})["id"].(int) <
			groupMembersList[j].(map[string]interface{})["id"].(int)
	})

	return groupMembersList
}
