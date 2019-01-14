package gitlab

import (
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	gitlab "github.com/xanzy/go-gitlab"
)

func resourceGitlabServiceJira() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabServiceJiraCreate,
		Read:   resourceGitlabServiceJiraRead,
		Update: resourceGitlabServiceJiraUpdate,
		Delete: resourceGitlabServiceJiraDelete,

		Schema: map[string]*schema.Schema{
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"title": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"url": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,6}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)`), "value must be a url"),
			},
			"project_key": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:     schema.TypeString,
				Required: true,
			},
			"jira_issue_transition_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"merge_requests_events": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceGitlabServiceJiraCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	project := d.Get("project").(string)

	jiraOptions, err := expandJiraOptions(d)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Create Gitlab Jira service")

	_, setServiceErr := client.Services.SetJiraService(project, jiraOptions)

	if err != nil {
		return fmt.Errorf("[ERROR] Couldn't create Gitlab Jira service: %s", setServiceErr)
	}

	d.SetId(project)

	return resourceGitlabServiceJiraRead(d, meta)
}

func resourceGitlabServiceJiraRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	project := d.Get("project").(string)

	log.Printf("[DEBUG] Read Gitlab Jira service %s", d.Id())

	var service gitlab.Service

	jiraService, response, err := client.Services.GetJiraService(project)
	if err != nil {
		if response.StatusCode == 404 {
			log.Printf("[WARN] removing Jira service from state because it no longer exists in Gitlab")
			d.SetId("")
			return nil
		}
		return err
	}
	flattenJiraOptions(d, jiraService)
	service = jiraService.Service

	d.Set("title", service.Title)
	d.Set("created_at", service.CreatedAt.String())
	d.Set("updated_at", service.UpdatedAt.String())
	d.Set("active", service.Active)
	d.Set("push_events", service.PushEvents)
	d.Set("issues_events", service.IssuesEvents)
	d.Set("confidential_issues_events", service.ConfidentialIssuesEvents)
	d.Set("merge_requests_events", service.MergeRequestsEvents)
	d.Set("tag_push_events", service.TagPushEvents)
	d.Set("note_events", service.NoteEvents)
	d.Set("pipeline_events", service.PipelineEvents)
	d.Set("job_events", service.JobEvents)
	d.Set("wikiPage_events", service.WikiPageEvents)
	d.Set("confidentialNote_events", service.ConfidentialNoteEvents)

	return nil
}

func resourceGitlabServiceJiraUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceGitlabServiceJiraCreate(d, meta)
}

func resourceGitlabServiceJiraDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	project := d.Get("project").(string)

	log.Printf("[DEBUG] Delete Gitlab Jira service %s", d.Id())

	_, err := client.Services.DeleteJiraService(project)

	return err
}

func expandJiraOptions(d *schema.ResourceData) (*gitlab.SetJiraServiceOptions, error) {
	setJiraServiceOptions := gitlab.SetJiraServiceOptions{}

	// Set required properties
	setJiraServiceOptions.URL = gitlab.String(d.Get("url").(string))
	setJiraServiceOptions.ProjectKey = gitlab.String(d.Get("project_key").(string))
	setJiraServiceOptions.Username = gitlab.String(d.Get("username").(string))
	setJiraServiceOptions.Password = gitlab.String(d.Get("password").(string))

	// Set optional properties
	if val := d.Get("jira_issue_transition_id"); val != nil {
		setJiraServiceOptions.JiraIssueTransitionID = gitlab.Int(val.(int))
	}
	if val := d.Get("merge_requests_events"); val != nil {
		setJiraServiceOptions.MergeRequestsEvents = gitlab.Bool(val.(bool))
	}

	return &setJiraServiceOptions, nil
}

func flattenJiraOptions(d *schema.ResourceData, service *gitlab.JiraService) []interface{} {
	values := []interface{}{}

	jiraOptions := map[string]interface{}{}
	jiraOptions["url"] = service.Properties.URL
	jiraOptions["project_key"] = service.Properties.ProjectKey
	jiraOptions["username"] = service.Properties.Username
	jiraOptions["password"] = service.Properties.Password
	jiraOptions["jira_issue_transition_id"] = service.Properties.JiraIssueTransitionID

	values = append(values, jiraOptions)

	return values
}
