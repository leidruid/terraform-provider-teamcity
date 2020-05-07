package teamcity

import (
	"fmt"
	"log"
	"strings"

	api "github.com/cvbarros/go-teamcity/teamcity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceProjectFeatureOauthProviderSettings() *schema.Resource {
	return &schema.Resource{
		Create: resourceProjectFeatureOauthProviderSettingsCreate,
		Read:   resourceProjectFeatureOauthProviderSettingsRead,
		Update: resourceProjectFeatureOauthProviderSettingsUpdate,
		Delete: resourceProjectFeatureOauthProviderSettingsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"feature_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"display_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"endpoint": {
				Type:     schema.TypeString,
				Required: true,
			},

			"fail_on_error": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"namespace": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"provider_type": {
				Type:     schema.TypeString,
				Required: true,
			},

			"role_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"secret_id": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},

			"url": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceProjectFeatureOauthProviderSettingsCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)

	projectId := d.Get("project_id").(string)
	service := client.ProjectFeatureService(projectId)

	feature := api.NewProjectFeatureOauthProviderSettings(projectId, api.ProjectFeatureOauthProviderSettingsOptions{
		DisplayName:  d.Get("display_name").(string),
		Endpoint:     d.Get("endpoint").(string),
		FailOnError:  d.Get("fail_on_error").(bool),
		Namespace:    d.Get("namespace").(string),
		ProviderType: d.Get("provider_type").(string),
		RoleId:       d.Get("role_id").(string),
		SecretId:     d.Get("secret_id").(string),
		Url:          d.Get("url").(string),
	})

	if createdProjectFeature, err := service.Create(feature); err != nil {
		return err
	} else {
		d.Set("feature_id", createdProjectFeature.ID())
		d.SetId(fmt.Sprintf("%s/%s", projectId, createdProjectFeature.ID()))
	}
	return resourceProjectFeatureOauthProviderSettingsRead(d, meta)
}

func resourceProjectFeatureOauthProviderSettingsUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)

	projectId := d.Get("project_id").(string)
	service := client.ProjectFeatureService(projectId)
	feature, err := service.GetByID(d.Get("feature_id").(string))
	if err != nil {
		return err
	}
	oauthProviderFeature, ok := feature.(*api.ProjectFeatureOauthProviderSettings)
	if !ok {
		return fmt.Errorf("Expected a OAuth Provider Feature but wasn't!")
	}

	oauthProviderFeature.Options.SecretId = d.Get("secret_id").(string)
	if d.HasChange("display_name") {
		oauthProviderFeature.Options.DisplayName = d.Get("display_name").(string)
	}
	if d.HasChange("endpoint") {
		oauthProviderFeature.Options.Endpoint = d.Get("endpoint").(string)
	}
	if d.HasChange("fail_on_error") {
		oauthProviderFeature.Options.FailOnError = d.Get("fail_on_error").(bool)
	}
	if d.HasChange("namespace") {
		oauthProviderFeature.Options.Namespace = d.Get("namespace").(string)
	}
	if d.HasChange("provider_type") {
		oauthProviderFeature.Options.ProviderType = d.Get("provider_type").(string)
	}
	if d.HasChange("role_id") {
		oauthProviderFeature.Options.RoleId = d.Get("role_id").(string)
	}
	if d.HasChange("url") {
		oauthProviderFeature.Options.Url = d.Get("url").(string)
	}

	if _, err := service.Update(oauthProviderFeature); err != nil {
		return err
	}

	return resourceProjectFeatureOauthProviderSettingsRead(d, meta)
}

func resourceProjectFeatureOauthProviderSettingsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)

	projectId := d.Get("project_id").(string)
	service := client.ProjectFeatureService(projectId)
	feature, err := service.GetByID(d.Get("feature_id").(string))
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			log.Printf("[DEBUG] Project Feature OAuthProvider was not found - removing from state!")
			d.SetId("")
			return nil
		}

		return err
	}

	oauthProviderFeature, ok := feature.(*api.ProjectFeatureOauthProviderSettings)
	if !ok {
		return fmt.Errorf("Expected a OAuthProvider Feature but wasn't!")
	}

	d.Set("project_id", projectId)
	d.Set("display_name", string(oauthProviderFeature.Options.DisplayName))
	d.Set("endpoint", oauthProviderFeature.Options.Endpoint)
	d.Set("fail_on_error", bool(oauthProviderFeature.Options.FailOnError))
	d.Set("namespace", oauthProviderFeature.Options.Namespace)
	d.Set("provider_type", oauthProviderFeature.Options.ProviderType)
	d.Set("role_id", oauthProviderFeature.Options.RoleId)
	d.Set("url", oauthProviderFeature.Options.Url)

	return nil
}

func resourceProjectFeatureOauthProviderSettingsDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)

	projectId := d.Get("project_id").(string)
	service := client.ProjectFeatureService(projectId)
	feature, err := service.GetByID(d.Get("feature_id").(string))
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			// already gone
			return nil
		}

		return err
	}

	return service.Delete(feature.ID())
}
