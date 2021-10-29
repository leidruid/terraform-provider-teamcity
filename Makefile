export GO111MODULE=on

GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
BUILD_DATE=$(shell date '+%Y-%m-%d-%H:%M:%S')
BUILDER_IMAGE=leidruid/terraform-provider-teamcity-builder

default: test

build:
#	GO111MODULE=on go build -o ./bin/terraform-provider-teamcity_${VERSION}
	GO111MODULE=on go build -ldflags="-w -s" -o ~/.terraform.d/plugins/github.com/leidruid/teamcity/${VERSION}/linux_amd64/terraform-provider-teamcity_v${VERSION}
	mkdir -p ~/.terraform.d/plugins/github.com/leidruid/teamcity/${VERSION}/linux_amd64/
	chmod +x ~/.terraform.d/plugins/github.com/leidruid/teamcity/${VERSION}/linux_amd64/terraform-provider-teamcity_v${VERSION}
	cd /mnt/c/Users/user/work/terraform/teamcity/ && rm .terraform.lock.hcl && terraform init

install: build
	cp ./bin/terraform-provider-teamcity_${VERSION} ~/.terraform.d/plugins/

clean:
	rm -rf ./bin

builder-action:
	docker run -e GITHUB_WORKSPACE='/github/workspace' -e GITHUB_REPOSITORY='terraform-provider-teamcity' -e GITHUB_REF='v0.0.1-alpha' --name terraform-provider-teamcity-builder $(BUILDER_IMAGE):latest

builder-image:
	docker build .github/builder --tag $(BUILDER_IMAGE)

clean_samples:
	find ./examples -name '*.tfstate' -delete
	find ./examples -name ".terraform" -type d -exec rm -rf "{}" \;

fmt_samples:
	terraform fmt -recursive examples/
