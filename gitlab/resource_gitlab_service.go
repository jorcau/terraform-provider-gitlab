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
	commonOptions := expandGitlabServiceCommonOptions(d)

	log.Printf("[DEBUG] Create %s Gitlab service", name)

	var setServiceErr error

	switch name {
	case "slack":
		slackOptions, err := expandSlackOptions(properties)
		slackOptions.SetGitlabServiceOptions = *commonOptions
		if err != nil {
			return err
		}
		_, setServiceErr = client.Services.SetSlackService(project, slackOptions)
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
		d.Set("properties", flattenSlackOptions(slackService))
		service = slackService.Service
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
	default:
		return serviceErrorMsg("name", name)
	}

	return err
}

func expandSlackOptions(d []interface{}) (*gitlab.SetSlackServiceOptions, error) {
	setSlackServiceOptions := gitlab.SetSlackServiceOptions{}
	properties := d[0].(map[string]interface{})

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
	if properties["username"] != nil {
		setSlackServiceOptions.Username = gitlab.String(properties["username"].(string))
	}
	if properties["channel"] != nil {
		setSlackServiceOptions.Channel = gitlab.String(properties["channel"].(string))
	}
	if properties["notify_only_broken_pipelines"] != nil {
		val, _ := strconv.ParseBool(properties["notify_only_broken_pipelines"].(string))
		setSlackServiceOptions.NotifyOnlyBrokenPipelines = gitlab.Bool(val)
	}
	if properties["notify_only_default_branch"] != nil {
		val, _ := strconv.ParseBool(properties["notify_only_default_branch"].(string))
		setSlackServiceOptions.NotifyOnlyDefaultBranch = gitlab.Bool(val)
	}
	if properties["push_channel"] != nil {
		setSlackServiceOptions.PushChannel = gitlab.String(properties["push_channel"].(string))
	}
	if properties["issue_channel"] != nil {
		setSlackServiceOptions.IssueChannel = gitlab.String(properties["issue_channel"].(string))
	}
	if properties["confidential_issue_channel"] != nil {
		setSlackServiceOptions.ConfidentialIssueChannel = gitlab.String(properties["confidential_issue_channel"].(string))
	}
	if properties["merge_request_channel"] != nil {
		setSlackServiceOptions.MergeRequestChannel = gitlab.String(properties["merge_request_channel"].(string))
	}
	if properties["note_channel"] != nil {
		setSlackServiceOptions.NoteChannel = gitlab.String(properties["note_channel"].(string))
	}
	// if properties["confidential_note_channel"] != nil {
	// 	setSlackServiceOptions.ConfidentialNoteChannel = gitlab.String(properties["confidential_note_channel"].(string))
	// }
	if properties["tag_push_channel"] != nil {
		setSlackServiceOptions.TagPushChannel = gitlab.String(properties["tag_push_channel"].(string))
	}
	if properties["pipeline_channel"] != nil {
		setSlackServiceOptions.PipelineChannel = gitlab.String(properties["pipeline_channel"].(string))
	}
	if properties["wiki_page_channel"] != nil {
		setSlackServiceOptions.WikiPageChannel = gitlab.String(properties["wiki_page_channel"].(string))
	}

	return &setSlackServiceOptions, nil
}

func flattenSlackOptions(service *gitlab.SlackService) []interface{} {
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

func expandGitlabServiceCommonOptions(d *schema.ResourceData) *gitlab.SetGitlabServiceOptions {
	commonOptions := gitlab.SetGitlabServiceOptions{}

	if val, ok := d.GetOk("push_events"); ok {
		commonOptions.PushEvents = gitlab.Bool(val.(bool))
	}
	if val, ok := d.GetOk("issues_events"); ok {
		commonOptions.IssuesEvents = gitlab.Bool(val.(bool))
	}
	if val, ok := d.GetOk("confidential_issues_events"); ok {
		commonOptions.ConfidentialIssuesEvents = gitlab.Bool(val.(bool))
	}
	if val, ok := d.GetOk("merge_requests_events"); ok {
		commonOptions.MergeRequestsEvents = gitlab.Bool(val.(bool))
	}
	if val, ok := d.GetOk("tag_push_events"); ok {
		commonOptions.TagPushEvents = gitlab.Bool(val.(bool))
	}
	if val, ok := d.GetOk("note_events"); ok {
		commonOptions.NoteEvents = gitlab.Bool(val.(bool))
	}
	if val, ok := d.GetOk("confidential_note_events"); ok {
		commonOptions.ConfidentialNoteEvents = gitlab.Bool(val.(bool))
	}
	if val, ok := d.GetOk("pipeline_events"); ok {
		commonOptions.PipelineEvents = gitlab.Bool(val.(bool))
	}
	if val, ok := d.GetOk("wiki_page_events"); ok {
		commonOptions.WikiPageEvents = gitlab.Bool(val.(bool))
	}

	return &commonOptions
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
