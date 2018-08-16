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
			},
			"issues_events": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"confidential_issues_events": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"merge_requests_events": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"tag_push_events": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"note_events": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"confidential_note_events": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"pipeline_events": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"wiki_page_events": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"job_events": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"properties": {
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
	}
}

func resourceGitlabServiceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	project := d.Get("project").(string)
	name := d.Get("name").(string)
	properties := d.Get("properties").(map[string]interface{})

	// gitlabService := gitlab.Service{
	// 	Active:                   d.Get("active").(bool),
	// 	PushEvents:               d.Get("pushEvents").(bool),
	// 	IssuesEvents:             d.Get("issuesEvents").(bool),
	// 	ConfidentialIssuesEvents: d.Get("confidentialIssuesEvents").(bool),
	// 	MergeRequestsEvents:      d.Get("mergeRequestsEvents").(bool),
	// 	TagPushEvents:            d.Get("tagPushEvents").(bool),
	// 	NoteEvents:               d.Get("noteEvents").(bool),
	// 	PipelineEvents:           d.Get("pipelineEvents").(bool),
	// 	JobEvents:                d.Get("jobEvents").(bool),
	// }

	log.Printf("[DEBUG] Create %s Gitlab service", name)

	var setServiceErr error

	switch name {
	case "slack":
		slackOptions, err := expandSlackOptions(properties)
		if err != nil {
			return err
		}
		_, setServiceErr = client.Services.SetSlackService(project, slackOptions)
	default:
		return serviceErrorMsg("name", name)
	}

	if setServiceErr != nil {
		return fmt.Errorf("[ERROR] Couldn't create Gitlab %s service", name)
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
	// d.Set("wikiPage_events", service.WikiPageEvents)
	// d.Set("confidentialNote_events", service.ConfidentialNoteEvents)

	return nil
}

func resourceGitlabServiceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	project := d.Get("project").(string)
	name := d.Get("name").(string)
	properties := d.Get("properties").(map[string]interface{})

	// if d.HasChange("title") {
	// 	title := d.Get("title").(string)
	// }
	// if d.HasChange("createdAt") {
	// 	createdAt := d.Get("createdAt").(string)
	// }
	// if d.HasChange("updatedAt") {
	// 	updatedAt := d.Get("updatedAt").(string)
	// }
	// if d.HasChange("active") {
	// 	active := d.Get("active").(string)
	// }
	// if d.HasChange("pushEvents") {
	// 	pushEvents := d.Get("pushEvents").(string)
	// }
	// if d.HasChange("issuesEvents") {
	// 	issuesEvents := d.Get("issuesEvents").(string)
	// }
	// if d.HasChange("confidentialIssuesEvents") {
	// 	confidentialIssuesEvents := d.Get("confidentialIssuesEvents").(string)
	// }
	// if d.HasChange("mergeRequestsEvents") {
	// 	mergeRequestsEvents := d.Get("mergeRequestsEvents").(string)
	// }
	// if d.HasChange("tagPushEvents") {
	// 	tagPushEvents := d.Get("tagPushEvents").(string)
	// }
	// if d.HasChange("noteEvents") {
	// 	noteEvents := d.Get("noteEvents").(string)
	// }
	// if d.HasChange("pipelineEvents") {
	// 	pipelineEvents := d.Get("pipelineEvents").(string)
	// }
	// if d.HasChange("jobEvents") {
	// 	jobEvents := d.Get("jobEvents").(string)
	// }
	// if d.HasChange("properties") {
	// 	properties := d.Get("properties").(map[string]string)
	// }

	log.Printf("[DEBUG] update gitlab label %s", d.Id())

	var setServiceErr error

	switch name {
	case "slack":
		slackOptions, err := expandSlackOptions(properties)
		if err != nil {
			return serviceErrorMsg("properties", name)
		}
		_, setServiceErr = client.Services.SetSlackService(project, slackOptions)
	default:
		return serviceErrorMsg("name", name)
	}

	if setServiceErr != nil {
		return fmt.Errorf("[ERROR] Couldn't create Gitlab service: %s", name)
	}

	return resourceGitlabServiceRead(d, meta)
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

func expandSlackOptions(properties map[string]interface{}) (*gitlab.SetSlackServiceOptions, error) {
	setSlackServiceOptions := &gitlab.SetSlackServiceOptions{}

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

	return setSlackServiceOptions, nil
}

func flattenSlackOptions(service *gitlab.SlackService) map[string]interface{} {
	values := make(map[string]interface{})

	values["notifyOnlyBrokenPipelines"] = strconv.FormatBool(service.Properties.NotifyOnlyBrokenPipelines)

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
