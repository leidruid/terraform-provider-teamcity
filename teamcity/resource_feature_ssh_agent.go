package teamcity

import (
	"fmt"
	api "github.com/cvbarros/go-teamcity/teamcity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceFeatureSshAgent() *schema.Resource {
	return &schema.Resource{
		Create: resourceFeatureSshAgentCreate,
		Read:   resourceFeatureSshAgentRead,
		Delete: resourceFeatureSshAgentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"build_config_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"uploaded_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceFeatureSshAgentCreate(d *schema.ResourceData, meta interface{}) error {
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

	dt, err := buildSshAgent(d)
	if err != nil {
		return err
	}
	out, err := srv.Create(dt)

	if err != nil {
		return err
	}

	d.SetId(out.ID())
	return resourceFeatureSshAgentRead(d, meta)
}

func resourceFeatureSshAgentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client).BuildFeatureService(d.Get("build_config_id").(string))

	dt, err := getBuildFeatureSshAgent(client, d.Id())
	if err != nil {
		return err
	}

	if err := d.Set("build_config_id", dt.BuildTypeID()); err != nil {
		return err
	}

	opt := dt.Options.(*api.SshAgentOptions)

	m := make(map[string]interface{})
	m["uploaded_key"] = opt.TeamcitySshKey

	return err
}

func resourceFeatureSshAgentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)
	svr := client.BuildFeatureService(d.Get("build_config_id").(string))

	return svr.Delete(d.Id())
}

func buildSshAgent(d *schema.ResourceData) (api.BuildFeature, error) {
	var opt api.SshAgentOptions
	sshKey := d.Get("uploaded_key").(string)
	opt = api.NewSshAgentOptions(sshKey)

	return api.NewFeatureSshAgent(opt)
}

func getBuildFeatureSshAgent(c *api.BuildFeatureService, id string) (*api.FeatureSshAgent, error) {
	dt, err := c.GetByID(id)
	if err != nil {
		return nil, err
	}

	fcsp := dt.(*api.FeatureSshAgent)
	return fcsp, nil
}
