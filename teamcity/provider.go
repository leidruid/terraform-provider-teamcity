package teamcity

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

//Provider is the plugin entry point
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"teamcity_project":                         resourceProject(),
			"teamcity_vcs_root_git":                    resourceVcsRootGit(),
			"teamcity_build_config":                    resourceBuildConfig(),
			"teamcity_snapshot_dependency":             resourceSnapshotDependency(),
			"teamcity_artifact_dependency":             resourceArtifactDependency(),
			"teamcity_build_trigger_vcs":               resourceBuildTriggerVcs(),
			"teamcity_build_trigger_build_finish":      resourceBuildTriggerBuildFinish(),
			"teamcity_build_trigger_schedule":          resourceBuildTriggerSchedule(),
			"teamcity_agent_requirement":               resourceAgentRequirement(),
			"teamcity_feature_commit_status_publisher": resourceFeatureCommitStatusPublisher(),
			"teamcity_group":                           resourceGroup(),
			"teamcity_feature_docker_support":          resourceFeatureDockerSupport(),
			"teamcity_feature_vcs_labeling":            resourceFeatureVcsLabeling(),
			"teamcity_feature_ssh_agent":               resourceFeatureSshAgent(),
			"teamcity_feature_perfmon":                 resourceFeaturePerformanceMonitor(),
			"teamcity_feature_pull_requests":           resourceFeaturePullRequests(),
			"teamcity_feature_file_content_replacer":   resourceFeatureFileContentReplacer(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"teamcity_project": dataSourceProject(),
		},
		Schema: map[string]*schema.Schema{
			"address": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("TEAMCITY_ADDR", nil),
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("TEAMCITY_USER", nil),
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("TEAMCITY_PASSWORD", nil),
			},
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Address:  d.Get("address").(string),
		Username: d.Get("username").(string),
		Password: d.Get("password").(string),
	}
	return config.Client()
}
