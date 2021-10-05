module github.com/cvbarros/terraform-provider-teamcity

go 1.13

replace github.com/cvbarros/go-teamcity => ./go-teamcity

require (
	github.com/cvbarros/go-teamcity v1.1.0
	github.com/dghubble/sling v1.3.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/terraform-plugin-sdk v1.17.2
	github.com/hashicorp/terraform-plugin-test v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.8 // indirect
	github.com/motemen/go-nuts v0.0.0-20190725124253-1d2432db96b0 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	google.golang.org/grpc v1.32.0 // indirect
)
