package gitlab

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	gitlab "github.com/xanzy/go-gitlab"
)

func resourceGitlabServiceSlack() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabServiceSlackCreate,
		Read:   resourceGitlabServiceSlackRead,
		Update: resourceGitlabServiceSlackUpdate,
		Delete: resourceGitlabServiceSlackDelete,

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
			"webhook": {
				Type:     schema.TypeString,
				Required: true,
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"channel": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"push_events": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"issues_events": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"confidential_issues_events": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"merge_requests_events": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"tag_push_events": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"note_events": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"confidential_note_events": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"pipeline_events": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"wiki_page_events": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"job_events": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"push_channel": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"issue_channel": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"confidential_issue_channel": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"merge_request_channel": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tag_push_channel": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"note_channel": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"pipeline_channel": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"wiki_page_channel": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceGitlabServiceSlackCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	project := d.Get("project").(string)

	slackOptions, err := expandSlackOptions(d)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Create Gitlab Slack service")

	_, setServiceErr := client.Services.SetSlackService(project, slackOptions)

	if err != nil {
		return fmt.Errorf("[ERROR] Couldn't create Gitlab Slack service: %s", setServiceErr)
	}

	d.SetId(project)

	return resourceGitlabServiceSlackRead(d, meta)
}

func resourceGitlabServiceSlackRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	project := d.Get("project").(string)

	log.Printf("[DEBUG] Read Gitlab Slack service %s", d.Id())

	var service gitlab.Service

	slackService, response, err := client.Services.GetSlackService(project)
	if err != nil {
		if response.StatusCode == 404 {
			log.Printf("[WARN] removing Slack service from state because it no longer exists in Gitlab")
			d.SetId("")
			return nil
		}
		return err
	}
	flattenSlackOptions(d, slackService)
	service = slackService.Service

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

func resourceGitlabServiceSlackUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceGitlabServiceSlackCreate(d, meta)
}

func resourceGitlabServiceSlackDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	project := d.Get("project").(string)

	log.Printf("[DEBUG] Delete Gitlab Slack service %s", d.Id())

	_, err := client.Services.DeleteSlackService(project)

	return err
}

func expandSlackOptions(d *schema.ResourceData) (*gitlab.SetSlackServiceOptions, error) {
	setSlackServiceOptions := gitlab.SetSlackServiceOptions{}

	// Set required properties
	setSlackServiceOptions.WebHook = gitlab.String(d.Get("webhook").(string))

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
