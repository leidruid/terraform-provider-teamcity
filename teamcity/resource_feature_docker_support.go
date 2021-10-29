package teamcity

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/leidruid/go-teamcity/teamcity"
)

func resourceFeatureDockerSupport() *schema.Resource {
	return &schema.Resource{
		Create: resourceFeatureDockerSupportCreate,
		Read:   resourceFeatureDockerSupportRead,
		Delete: resourceFeatureDockerSupportDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"build_config_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"docker_registry": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Foreign key of Docker Registry Connection",
			},
			"cleanup": {
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    true,
				Description: "On server clean-up, delete pushed Docker images from registry",
			},
		},
	}
}

func resourceFeatureDockerSupportCreate(d *schema.ResourceData, meta interface{}) error {
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

	dt, err := buildDockerSupport(d)
	if err != nil {
		return err
	}
	out, err := srv.Create(dt)

	if err != nil {
		return err
	}

	d.SetId(out.ID())
	return resourceFeatureDockerSupportRead(d, meta)
}

func resourceFeatureDockerSupportRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client).BuildFeatureService(d.Get("build_config_id").(string))

	dt, err := getBuildFeatureDockerSupport(client, d.Id())
	if err != nil {
		return err
	}

	if err := d.Set("build_config_id", dt.BuildTypeID()); err != nil {
		return err
	}

	opt := dt.Options.(*api.DockerSupportOptions)

	m := make(map[string]interface{})
	m["docker_registry"] = opt.Login2registry
	m["cleanupPushed"] = opt.CleanupPushed

	return err
}

func resourceFeatureDockerSupportDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)
	svr := client.BuildFeatureService(d.Get("build_config_id").(string))

	return svr.Delete(d.Id())
}

func buildDockerSupport(d *schema.ResourceData) (api.BuildFeature, error) {
	var opt api.DockerSupportOptions
	registry := d.Get("docker_registry").(string)
	cleanup := d.Get("cleanup").(bool)

	opt = api.NewDockerSupportOptions(registry, cleanup)

	return api.NewFeatureDockerSupport(opt)
}

func getBuildFeatureDockerSupport(c *api.BuildFeatureService, id string) (*api.FeatureDockerSupport, error) {
	dt, err := c.GetByID(id)
	if err != nil {
		return nil, err
	}

	fcsp := dt.(*api.FeatureDockerSupport)
	return fcsp, nil
}
