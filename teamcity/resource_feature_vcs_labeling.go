package teamcity

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/leidruid/go-teamcity/teamcity"
	"log"
)

func resourceFeatureVcsLabeling() *schema.Resource {
	return &schema.Resource{
		Create: resourceFeatureVcsLabelingCreate,
		Read:   resourceFeatureVcsLabelingRead,
		Delete: resourceFeatureVcsLabelingDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"build_config_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_vcs_config_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"branch_filter": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true,
				Optional: true,
			},
			"successful_only": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"labeling_pattern": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceFeatureVcsLabelingCreate(d *schema.ResourceData, meta interface{}) error {
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

	dt, err := buildVcsLabeling(d)
	if err != nil {
		return err
	}
	out, err := srv.Create(dt)

	if err != nil {
		return err
	}

	d.SetId(out.ID())
	return resourceFeatureVcsLabelingRead(d, meta)
}

func resourceFeatureVcsLabelingRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client).BuildFeatureService(d.Get("build_config_id").(string))

	dt, err := getBuildFeatureVcsLabeling(client, d.Id())
	if err != nil {
		return err
	}

	if err := d.Set("build_config_id", dt.BuildTypeID()); err != nil {
		return err
	}

	opt := dt.Options.(*api.VcsLabelingOptions)

	m := make(map[string]interface{})
	m["branch_filter"] = opt.BranchFilter
	m["labeling_pattern"] = opt.LabelingPattern
	m["source_vcs_config_id"] = opt.VcsRootId
	m["successful_only"] = opt.SuccessfulOnly

	return err
}

func resourceFeatureVcsLabelingDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)
	svr := client.BuildFeatureService(d.Get("build_config_id").(string))

	return svr.Delete(d.Id())
}

func buildVcsLabeling(d *schema.ResourceData) (api.BuildFeature, error) {
	var opt api.VcsLabelingOptions
	var filter []string

	if v, ok := d.GetOk("branch_filter"); ok {
		filter = expandStringSlice(v.([]interface{}))
		log.Printf("[INFO] BranchFilter: %s, State: %s", filter, v)
	}

	label := d.Get("labeling_pattern").(string)
	vcs := d.Get("source_vcs_config_id").(string)
	success := d.Get("successful_only").(bool)

	opt = api.NewVcsLabelingOptions(filter, label, vcs, success)

	return api.NewFeatureVcsLabeling(opt)
}

func getBuildFeatureVcsLabeling(c *api.BuildFeatureService, id string) (*api.FeatureVcsLabeling, error) {
	dt, err := c.GetByID(id)
	if err != nil {
		return nil, err
	}

	fcsp := dt.(*api.FeatureVcsLabeling)
	return fcsp, nil
}
