package teamcity

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	api "github.com/leidruid/go-teamcity/teamcity"
)

func resourceFeaturePullRequests() *schema.Resource {
	return &schema.Resource{
		Create: resourceFeaturePullRequestsCreate,
		Read:   resourceFeaturePullRequestsRead,
		Delete: resourceFeaturePullRequestsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"build_config_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hosting_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"bitbucket_server"}, true),
			},
			"bitbucket_server": {
				Type:     schema.TypeSet,
				ForceNew: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"vcsRoot", "password"}, true),
							ForceNew:     true,
						},
						"host": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"username": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"password": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
							Computed:  true,
							ForceNew:  true,
						},
						"filter_source_branch": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							ForceNew: true,
							Optional: true,
						},
						"filter_target_branch": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							ForceNew: true,
							Optional: true,
						},
					},
				},
				Set: bitbucketServerOptionsHash,
			},
		},
	}
}

func resourceFeaturePullRequestsCreate(d *schema.ResourceData, meta interface{}) error {
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

	//Only Bitbucket Server publisher for now - Add support for more publishers later

	if ht, ok := d.GetOk("hosting_type"); ok {
		switch ht {
		case "bitbucket_server":
			dt, err := buildBitbucketServerPullRequests(d)
			if err != nil {
				return err
			}
			out, err := srv.Create(dt)
			if err != nil {
				return err
			}
			d.SetId(out.ID())
		}
	}

	return resourceFeaturePullRequestsRead(d, meta)
}

func resourceFeaturePullRequestsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client).BuildFeatureService(d.Get("build_config_id").(string))

	dt, err := getBuildFeaturePullRequests(client, d.Id())
	if err != nil {
		return err
	}

	if err := d.Set("build_config_id", dt.BuildTypeID()); err != nil {
		return err
	}

	//TODO: Implement other publishers

	_, ok := d.GetOk("bitbucket_server")
	if ok {
		if err := d.Set("hosting_type", "bitbucket_server"); err != nil {
			return err
		}
		optsToSave := resourceReadPullRequests(dt)
		return d.Set("bitbucket_server", optsToSave)
	}

	return err
}

func resourceReadPullRequests(dt *api.FeaturePullRequests) (optsToSave []map[string]interface{}) {
	opt := dt.Options.(*api.PullRequestsOptions)

	m := make(map[string]interface{})
	m["host"] = opt.ServerUrl
	m["username"] = opt.Username
	m["password"] = opt.Password
	m["auth_type"] = opt.AuthenticationType
	m["filter_source_branch"] = opt.FilterSourceBranch
	m["filter_target_branch"] = opt.FilterTargetBranch

	optsToSave = append(optsToSave, m)
	return
}

func resourceFeaturePullRequestsDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)
	svr := client.BuildFeatureService(d.Get("build_config_id").(string))

	return svr.Delete(d.Id())
}

func buildBitbucketServerPullRequests(d *schema.ResourceData) (api.BuildFeature, error) {
	var opt api.PullRequestsOptions

	// MaxItems ensure at most 1 github element
	local := d.Get("bitbucket_server").(*schema.Set).List()[0].(map[string]interface{})
	url := local["host"].(string)
	username := local["username"].(string)
	password := local["password"].(string)
	source := expandStringSlice(local["filter_source_branch"].([]interface{}))
	target := expandStringSlice(local["filter_target_branch"].([]interface{}))
	authType := local["auth_type"].(string)

	switch authType {
	case "password":
		opt = api.NewPullRequestsOptionsPassword(username, password, url, source, target)
	case "vcsRoot":
		opt = api.NewPullRequestsOptionsVcs(url, source, target)
	}

	return api.NewFeaturePullRequests(opt, "")
}

func getBuildFeaturePullRequests(c *api.BuildFeatureService, id string) (*api.FeaturePullRequests, error) {
	dt, err := c.GetByID(id)
	if err != nil {
		return nil, err
	}

	fcsp := dt.(*api.FeaturePullRequests)
	return fcsp, nil
}

func bitbucketServerOptionsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["host"].(string)))

	if v, ok := m["username"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return hashcode.String(buf.String())
}
