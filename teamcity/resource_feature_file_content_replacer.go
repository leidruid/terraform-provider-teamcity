package teamcity

import (
	"fmt"
	api "github.com/cvbarros/go-teamcity/teamcity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceFeatureFileContentReplacer() *schema.Resource {
	return &schema.Resource{
		Create: resourceFeatureFileContentReplacerCreate,
		Read:   resourceFeatureFileContentReplacerRead,
		Delete: resourceFeatureFileContentReplacerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"build_config_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"fail_build": {
				Type:     schema.TypeBool,
				ForceNew: true,
				Optional: true,
				Default:  false,
			},
			"file_encoding": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"autodetect",
					"US-ASCII",
					"UTF-8",
					"UTF-16BE",
					"UTF-16LE",
					"custom",
				}, false),
			},
			"encoding_custom": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"find_what": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"match_case": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"regex_mode": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"FIXED_STRINGS",
					"REGEX",
					"REGEX_MIXED",
				}, false),
			},
			"replace_with": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
				ForceNew: true,
			},
			"process_files": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true,
				Optional: true,
			},
		},
	}
}

func resourceFeatureFileContentReplacerCreate(d *schema.ResourceData, meta interface{}) error {
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

	dt, err := buildFileContentReplacer(d)
	if err != nil {
		return err
	}
	out, err := srv.Create(dt)

	if err != nil {
		return err
	}

	d.SetId(out.ID())
	return resourceFeatureFileContentReplacerRead(d, meta)
}

func resourceFeatureFileContentReplacerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client).BuildFeatureService(d.Get("build_config_id").(string))

	dt, err := getBuildFeatureFileContentReplacer(client, d.Id())
	if err != nil {
		return err
	}

	if err := d.Set("build_config_id", dt.BuildTypeID()); err != nil {
		return err
	}

	opt := dt.Options.(*api.FileContentReplacerOptions)

	m := make(map[string]interface{})
	m["fail_build"] = opt.TeamcityFileContentReplacerFailBuild
	m["file_encoding"] = opt.TeamcityFileContentReplacerFileEncoding
	m["encoding_custom"] = opt.TeamcityFileContentReplacerFileEncodingCustom
	m["find_what"] = opt.TeamcityFileContentReplacerPattern
	m["match_case"] = opt.TeamcityFileContentReplacerPatternCaseSensitive
	m["regex_mode"] = opt.TeamcityFileContentReplacerRegexMode
	m["replace_with"] = opt.TeamcityFileContentReplacerReplacement
	m["process_files"] = opt.TeamcityFileContentReplacerWildcards

	return err
}

func resourceFeatureFileContentReplacerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)
	svr := client.BuildFeatureService(d.Get("build_config_id").(string))

	return svr.Delete(d.Id())
}

func buildFileContentReplacer(d *schema.ResourceData) (api.BuildFeature, error) {
	var opt api.FileContentReplacerOptions
	var encodingCustom string

	encoding := d.Get("file_encoding").(string)

	if encoding != "custom" {
		encodingCustom = d.Get("file_encoding").(string)
	} else {
		encodingCustom = d.Get("encoding_custom").(string)
	}

	pattern := d.Get("find_what").(string)
	regexMode := d.Get("regex_mode").(string)
	replacement := d.Get("replace_with").(string)
	wildcards := expandStringSlice(d.Get("process_files").([]interface{}))
	failBuild := d.Get("fail_build").(bool)
	caseSensitive := d.Get("match_case").(bool)

	opt = api.NewFileContentReplacerOptions(encoding, encodingCustom, pattern, regexMode, replacement, wildcards, failBuild, caseSensitive)

	return api.NewFeatureFileContentReplacer(opt)
}

func getBuildFeatureFileContentReplacer(c *api.BuildFeatureService, id string) (*api.FeatureFileContentReplacer, error) {
	dt, err := c.GetByID(id)
	if err != nil {
		return nil, err
	}

	fcsp := dt.(*api.FeatureFileContentReplacer)
	return fcsp, nil
}
