package teamcity

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/leidruid/go-teamcity/teamcity"
)

func dataSourceProject() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceProjectRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Computed:      true,
				ConflictsWith: []string{"project_id"},
			},
			"parent_project_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceProjectRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)
	var id, name string
	var dt *api.Project

	if v, ok := d.GetOk("project_id"); ok {
		id = v.(string)
	}
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	}
	if id != "" {
		p, err := client.Projects.GetByID(id)
		if err != nil {
			return err
		}
		dt = p
	}
	if name != "" {
		p, err := client.Projects.GetByName(name)
		if err != nil {
			return err
		}
		dt = p
	}
	if dt == nil {
		return fmt.Errorf("error when retrieving project, either `project_id` or `name` are required to be set")
	}
	d.SetId(dt.ID)
	if err := d.Set("name", dt.Name); err != nil {
		return err
	}
	if err := d.Set("project_id", dt.ID); err != nil {
		return err
	}
	if err := d.Set("parent_project_id", dt.ParentProjectID); err != nil {
		return err
	}
	if err := d.Set("url", dt.WebURL); err != nil {
		return err
	}
	return nil
}
