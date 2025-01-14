package teamcity

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	api "github.com/leidruid/go-teamcity/teamcity"
)

func resourceFeatureCommitStatusPublisher() *schema.Resource {
	return &schema.Resource{
		Create: resourceFeatureCommitStatusPublisherCreate,
		Read:   resourceFeatureCommitStatusPublisherRead,
		Delete: resourceFeatureCommitStatusPublisherDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"build_config_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"publisher": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"github", "bitbucket_server"}, true),
			},
			"github": {
				Type:     schema.TypeSet,
				ForceNew: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"token", "password"}, true),
							ForceNew:     true,
						},
						"host": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "https://api.github.com",
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
						"access_token": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
							Computed:  true,
							ForceNew:  true,
						},
					},
				},
				Set: githubPublisherOptionsHash,
			},
			"bitbucket_server": {
				Type:     schema.TypeSet,
				ForceNew: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
					},
				},
				Set: bitbucketPublisherOptionsHash,
			},
		},
	}
}

func resourceFeatureCommitStatusPublisherCreate(d *schema.ResourceData, meta interface{}) error {
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

	//Only Github publisher for now - Add support for more publishers later

	if pub, ok := d.GetOk("publisher"); ok {
		switch pub {
		case "github":
			dt, err := buildGithubCommitStatusPublisher(d)
			if err != nil {
				return err
			}
			out, err := srv.Create(dt)
			if err != nil {
				return err
			}
			d.SetId(out.ID())
		case "bitbucket_server":
			dt, err := buildBitbucketServerCommitStatusPublisher(d)
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

	return resourceFeatureCommitStatusPublisherRead(d, meta)
}

func resourceFeatureCommitStatusPublisherRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client).BuildFeatureService(d.Get("build_config_id").(string))

	dt, err := getBuildFeatureCommitPublisher(client, d.Id())
	if err != nil {
		return err
	}

	if err := d.Set("build_config_id", dt.BuildTypeID()); err != nil {
		return err
	}

	//TODO: Implement other publishers

	_, ok := d.GetOk("bitbucket_server")
	if ok {
		if err := d.Set("publisher", "bitbucket_server"); err != nil {
			return err
		}
		optsToSave := resourceReadBitBucketStatusPublisher(dt)
		return d.Set("bitbucket_server", optsToSave)
	}
	_, ok = d.GetOk("github")
	if ok {
		if err := d.Set("publisher", "github"); err != nil {
			return err
		}
		optsToSave := resourceReadGithubStatusPublisher(dt)
		return d.Set("bitbucket_server", optsToSave)
	}
	return err
}

func resourceReadBitBucketStatusPublisher(dt *api.FeatureCommitStatusPublisher) (optsToSave []map[string]interface{}) {
	opt := dt.Options.(*api.StatusPublisherBitbucketServerOptions)

	m := make(map[string]interface{})
	m["host"] = opt.Host
	m["password"] = opt.Password
	m["username"] = opt.Username

	optsToSave = append(optsToSave, m)
	return
}

func resourceReadGithubStatusPublisher(dt *api.FeatureCommitStatusPublisher) (optsToSave []map[string]interface{}) {
	opt := dt.Options.(*api.StatusPublisherGithubOptions)

	m := make(map[string]interface{})
	m["auth_type"] = opt.AuthenticationType
	m["host"] = opt.Host

	if opt.AuthenticationType == "password" {
		m["username"] = opt.Username
	}

	optsToSave = append(optsToSave, m)
	return
}

func resourceFeatureCommitStatusPublisherDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)
	svr := client.BuildFeatureService(d.Get("build_config_id").(string))

	return svr.Delete(d.Id())
}

func buildGithubCommitStatusPublisher(d *schema.ResourceData) (api.BuildFeature, error) {
	var opt api.StatusPublisherGithubOptions
	// MaxItems ensure at most 1 github element
	local := d.Get("github").(*schema.Set).List()[0].(map[string]interface{})
	host := local["host"].(string)
	authType := local["auth_type"].(string)
	switch strings.ToLower(authType) {
	case "token":
		opt = api.NewCommitStatusPublisherGithubOptionsToken(host, local["access_token"].(string))
	case "password":
		opt = api.NewCommitStatusPublisherGithubOptionsPassword(host, local["username"].(string), local["password"].(string))
	}

	return api.NewFeatureCommitStatusPublisherGithub(opt, "")
}

func buildBitbucketServerCommitStatusPublisher(d *schema.ResourceData) (api.BuildFeature, error) {
	var opt api.StatusPublisherBitbucketServerOptions
	// MaxItems ensure at most 1 github element
	local := d.Get("bitbucket_server").(*schema.Set).List()[0].(map[string]interface{})
	host := local["host"].(string)
	username := local["username"].(string)
	password := local["password"].(string)
	opt = api.NewCommitStatusPublisherBitbucketServerOptionsPassword(host, username, password)

	return api.NewFeatureCommitStatusPublisherBitbucketServer(opt, "")
}

func getBuildFeatureCommitPublisher(c *api.BuildFeatureService, id string) (*api.FeatureCommitStatusPublisher, error) {
	dt, err := c.GetByID(id)
	if err != nil {
		return nil, err
	}

	fcsp := dt.(*api.FeatureCommitStatusPublisher)
	return fcsp, nil
}

func githubPublisherOptionsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["auth_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["host"].(string)))

	if v, ok := m["username"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return hashcode.String(buf.String())
}
func bitbucketPublisherOptionsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["host"].(string)))

	if v, ok := m["username"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return hashcode.String(buf.String())
}
