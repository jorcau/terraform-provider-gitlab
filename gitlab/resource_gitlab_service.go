package gitlab

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	gitlab "github.com/xanzy/go-gitlab"
)

func resourceGitlabService() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabServiceCreate,
		Read:   resourceGitlabServiceRead,
		Update: resourceGitlabServiceUpdate,
		Delete: resourceGitlabServiceDelete,

		Schema: map[string]*schema.Schema{
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
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
			"push_events": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
			"issues_events": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
			"confidential_issues_events": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
			"merge_requests_events": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
			"tag_push_events": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
			"note_events": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
			"confidential_note_events": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
			"pipeline_events": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
			"wiki_page_events": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
			"job_events": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"properties": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Computed: true,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeMap},
			},
		},
	}
}

func resourceGitlabServiceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	project := d.Get("project").(string)
	name := d.Get("name").(string)
	properties := d.Get("properties").([]interface{})

	log.Printf("[DEBUG] Create %s Gitlab service", name)

	var setServiceErr error

	switch name {
	case "slack":
		slackOptions, err := expandSlackOptions(d, properties)
		if err != nil {
			return err
		}
		_, setServiceErr = client.Services.SetSlackService(project, slackOptions)
	case "jira":
		jiraOptions, err := expandJiraOptions(properties)
		if err != nil {
			return err
		}
		_, setServiceErr = client.Services.SetJiraService(project, jiraOptions)
	default:
		return serviceErrorMsg("name", name)
	}

	if setServiceErr != nil {
		return fmt.Errorf("[ERROR] Couldn't create Gitlab %s service: %s", name, setServiceErr)
	}

	d.SetId(fmt.Sprintf("%s/%s", project, name))

	return resourceGitlabServiceRead(d, meta)
}

func handleGetServiceError(d *schema.ResourceData, name string, statusCode int, err error) error {
	if statusCode == 404 {
		log.Printf("[WARN] removing service %s from state because it no longer exists in gitlab", name)
		d.SetId("")
		return nil
	}
	return err
}

func resourceGitlabServiceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	project := d.Get("project").(string)
	name := d.Get("name").(string)

	log.Printf("[DEBUG] Read Gitlab service %s", d.Id())

	var service gitlab.Service

	switch name {
	case "slack":
		slackService, response, err := client.Services.GetSlackService(project)
		if err != nil {
			return handleGetServiceError(d, name, response.StatusCode, err)
		}
		flattenSlackOptions(d, slackService)
		service = slackService.Service
	case "jira":
		jiraService, response, err := client.Services.GetJiraService(project)
		if err != nil {
			return handleGetServiceError(d, name, response.StatusCode, err)
		}
		d.Set("properties", flattenJiraOptions(jiraService))
		service = jiraService.Service
	default:
		return serviceErrorMsg("name", name)
	}

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

func resourceGitlabServiceUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceGitlabServiceCreate(d, meta)
}

func resourceGitlabServiceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	project := d.Get("project").(string)
	name := d.Get("name").(string)

	log.Printf("[DEBUG] Delete Gitlab service %s", d.Id())

	var err error
	switch name {
	case "slack":
		_, err = client.Services.DeleteSlackService(project)
	case "jira":
		_, err = client.Services.DeleteJiraService(project)
	default:
		return serviceErrorMsg("name", name)
	}

	return err
}

func expandSlackOptions(d *schema.ResourceData, data []interface{}) (*gitlab.SetSlackServiceOptions, error) {
	setSlackServiceOptions := gitlab.SetSlackServiceOptions{}
	properties := data[0].(map[string]interface{})

	// Check required properties are set and valid
	requiredProperties := []string{"webhook"}
	missingProperties := []string{}
	for i := 0; i < len(requiredProperties); i++ {
		if properties[requiredProperties[i]] == nil {
			missingProperties = append(missingProperties, requiredProperties[i])
		}
	}
	if len(missingProperties) > 0 {
		return nil, fmt.Errorf("[ERROR] Following properties are required: %s", strings.Join(missingProperties, ", "))
	}
	_, err := url.ParseRequestURI(properties["webhook"].(string))
	if err != nil {
		return nil, fmt.Errorf("[ERROR] 'webhook' propertie must be a valid URL")
	}

	// Set required properties
	setSlackServiceOptions.WebHook = gitlab.String(properties["webhook"].(string))

	// Set optional properties
	if val := d.Get("username"); val != nil {
		setSlackServiceOptions.Username = gitlab.String(val.(string))
	}
	if val := d.Get("channel"); val != nil {
		setSlackServiceOptions.Channel = gitlab.String(val.(string))
	}
	if val := d.Get("notify_only_broken_pipelines"); val != nil {
		value, _ := strconv.ParseBool(val.(string))
		setSlackServiceOptions.NotifyOnlyBrokenPipelines = gitlab.Bool(value)
	}
	if val := d.Get("notify_only_default_branch"); val != nil {
		value, _ := strconv.ParseBool(val.(string))
		setSlackServiceOptions.NotifyOnlyDefaultBranch = gitlab.Bool(value)
	}
	if val := d.Get("push_channel"); val != nil {
		setSlackServiceOptions.PushChannel = gitlab.String(val.(string))
	}
	if val := d.Get("issue_channel"); val != nil {
		setSlackServiceOptions.IssueChannel = gitlab.String(val.(string))
	}
	if val := d.Get("confidential_issue_channel"); val != nil {
		setSlackServiceOptions.ConfidentialIssueChannel = gitlab.String(val.(string))
	}
	if val := d.Get("merge_request_channel"); val != nil {
		setSlackServiceOptions.MergeRequestChannel = gitlab.String(val.(string))
	}
	if val := d.Get("note_channel"); val != nil {
		setSlackServiceOptions.NoteChannel = gitlab.String(val.(string))
	}
	// https://gitlab.com/gitlab-org/gitlab-ce/issues/49730
	// if properties["confidential_note_channel"] != nil {
	// 	setSlackServiceOptions.ConfidentialNoteChannel = gitlab.String(properties["confidential_note_channel"].(string))
	// }
	if val := d.Get("tag_push_channel"); val != nil {
		setSlackServiceOptions.TagPushChannel = gitlab.String(val.(string))
	}
	if val := d.Get("pipeline_channel"); val != nil {
		setSlackServiceOptions.PipelineChannel = gitlab.String(val.(string))
	}
	if val := d.Get("wiki_page_channel"); val != nil {
		setSlackServiceOptions.WikiPageChannel = gitlab.String(val.(string))
	}

	// Set other optional parameters
	if val := d.Get("push_events"); val != nil {
		setSlackServiceOptions.PushEvents = gitlab.Bool(val.(bool))
	}
	if val := d.Get("issues_events"); val != nil {
		setSlackServiceOptions.IssuesEvents = gitlab.Bool(val.(bool))
	}
	if val := d.Get("confidential_issues_events"); val != nil {
		setSlackServiceOptions.ConfidentialIssuesEvents = gitlab.Bool(val.(bool))
	}
	if val := d.Get("merge_requests_events"); val != nil {
		setSlackServiceOptions.MergeRequestsEvents = gitlab.Bool(val.(bool))
	}
	if val := d.Get("tag_push_events"); val != nil {
		setSlackServiceOptions.TagPushEvents = gitlab.Bool(val.(bool))
	}
	if val := d.Get("note_events"); val != nil {
		setSlackServiceOptions.NoteEvents = gitlab.Bool(val.(bool))
	}
	if val := d.Get("confidential_note_events"); val != nil {
		setSlackServiceOptions.ConfidentialNoteEvents = gitlab.Bool(val.(bool))
	}
	if val := d.Get("pipeline_events"); val != nil {
		setSlackServiceOptions.PipelineEvents = gitlab.Bool(val.(bool))
	}
	if val := d.Get("wiki_page_events"); val != nil {
		setSlackServiceOptions.WikiPageEvents = gitlab.Bool(val.(bool))
	}

	return &setSlackServiceOptions, nil
}

func flattenSlackOptions(d *schema.ResourceData, service *gitlab.SlackService) []interface{} {
	values := []interface{}{}

	slackOptions := map[string]interface{}{}
	slackOptions["webhook"] = service.Properties.WebHook
	slackOptions["username"] = service.Properties.Username
	slackOptions["notify_only_broken_pipelines"] = service.Properties.NotifyOnlyDefaultBranch.UnmarshalJSON
	slackOptions["notify_only_default_branch"] = service.Properties.NotifyOnlyBrokenPipelines.UnmarshalJSON
	slackOptions["push_channel"] = service.Properties.PushChannel
	slackOptions["issue_channel"] = service.Properties.IssueChannel
	slackOptions["confidential_issue_channel"] = service.Properties.ConfidentialIssueChannel
	slackOptions["merge_request_channel"] = service.Properties.MergeRequestChannel
	slackOptions["note_channel"] = service.Properties.NoteChannel
	slackOptions["confidential_note_channel"] = service.Properties.ConfidentialNoteChannel
	slackOptions["tag_push_channel"] = service.Properties.TagPushChannel
	slackOptions["pipeline_channel"] = service.Properties.PipelineChannel
	slackOptions["wiki_page_channel"] = service.Properties.WikiPageChannel

	values = append(values, slackOptions)

	return values
}

func expandJiraOptions(d []interface{}) (*gitlab.SetJiraServiceOptions, error) {
	setJiraServiceOptions := gitlab.SetJiraServiceOptions{}
	properties := d[0].(map[string]interface{})

	// Check required properties are set and valid
	requiredProperties := []string{"url", "project_key", "username", "password"}
	missingProperties := []string{}
	for i := 0; i < len(requiredProperties); i++ {
		if properties[requiredProperties[i]] == nil {
			missingProperties = append(missingProperties, requiredProperties[i])
		}
	}
	if len(missingProperties) > 0 {
		return nil, fmt.Errorf("[ERROR] Following properties are required: %s", strings.Join(missingProperties, ", "))
	}
	_, err := url.ParseRequestURI(properties["url"].(string))
	if err != nil {
		return nil, fmt.Errorf("[ERROR] 'url' propertie must be a valid URL")
	}

	// Set required properties
	setJiraServiceOptions.URL = gitlab.String(properties["url"].(string))
	setJiraServiceOptions.ProjectKey = gitlab.String(properties["project_key"].(string))
	setJiraServiceOptions.Username = gitlab.String(properties["username"].(string))
	setJiraServiceOptions.Password = gitlab.String(properties["password"].(string))

	// Set optional properties
	if properties["jira_issue_transition_id"] != nil {
		setJiraServiceOptions.JiraIssueTransitionID = gitlab.String(properties["jira_issue_transition_id"].(string))
	}

	return &setJiraServiceOptions, nil
}

func flattenJiraOptions(service *gitlab.JiraService) []interface{} {
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

func serviceErrorMsg(errType string, name string) error {
	switch errType {
	case "name":
		return fmt.Errorf("[ERROR] Invalid Gitlab service: %s", name)
	case "properties":
		return fmt.Errorf("[ERROR] Invalid options for Gitlab %s service", name)
	default:
		return fmt.Errorf("[ERROR] Unknown error")
	}
}
