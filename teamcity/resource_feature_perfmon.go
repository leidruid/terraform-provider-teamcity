package teamcity

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/leidruid/go-teamcity/teamcity"
)

func resourceFeaturePerformanceMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceFeaturePerformanceMonitorCreate,
		Read:   resourceFeaturePerformanceMonitorRead,
		Delete: resourceFeaturePerformanceMonitorDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"build_config_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceFeaturePerformanceMonitorCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)
	var buildConfigID string

	if v, ok := d.GetOk("build_config_id"); ok {
		buildConfigID = v.(string)
	}

	// validates the Build Configuration exists
	if _, err := client.BuildTypes.GetByID(buildConfigID); err != nil {
		return fmt.Errorf("invalid build_config_id '%s' - Build configuration does not exist", buildConfigID)
	}

	srv := client.BuildFeatureService(buildConfigID)

	dt, _ := buildPerformanceMonitor(d)
	out, err := srv.Create(dt)

	if err != nil {
		return err
	}

	d.SetId(out.ID())
	return resourceFeaturePerformanceMonitorRead(d, meta)
}

func resourceFeaturePerformanceMonitorRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client).BuildFeatureService(d.Get("build_config_id").(string))

	dt, err := getBuildFeaturePerformanceMonitor(client, d.Id())
	if err != nil {
		return err
	}

	if err := d.Set("build_config_id", dt.BuildTypeID()); err != nil {
		return err
	}

	return err
}

func resourceFeaturePerformanceMonitorDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)
	svr := client.BuildFeatureService(d.Get("build_config_id").(string))

	return svr.Delete(d.Id())
}

func buildPerformanceMonitor(d *schema.ResourceData) (api.BuildFeature, error) {
	return api.NewPerformanceMonitor()
}

func getBuildFeaturePerformanceMonitor(c *api.BuildFeatureService, id string) (*api.FeaturePerformanceMonitor, error) {
	dt, err := c.GetByID(id)
	if err != nil {
		return nil, err
	}

	fcsp := dt.(*api.FeaturePerformanceMonitor)
	return fcsp, nil
}
