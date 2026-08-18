.PHONY: run build image push clean image-multiarch push-multiarch

tag = v0.0.1
releaseName = sensecraft-voice-client
dockerhubUser ?= registry.example.com/respeaker
targetDir ?= dist
commitHash = $(shell git rev-parse --short HEAD)
# e.g. 1862ce5-20240203180617
version = $(commitHash)-$(shell date +%Y%m%d%H%M%S)

ALL: run

run: build
	./sensecraftVoice --configfile ./config.yaml

build:
	go build -o $(targetDir)/$(releaseName) -ldflags "-X 'main.version=$(version)'" ./cmd/

# 构建多架构镜像（支持AMD64和ARM64）
image-multiarch:
	docker buildx build --platform linux/amd64,linux/arm64 \
		-t $(dockerhubUser)/$(releaseName):$(tag) \
		--build-arg VERSION=$(version) \
		--push .

# 构建单架构镜像（当前架构）
image:
	docker build -t $(dockerhubUser)/$(releaseName):$(tag) --build-arg VERSION=$(version) .

# 推送多架构镜像
push-multiarch: image-multiarch

# 推送单架构镜像
push: image
	docker push $(dockerhubUser)/$(releaseName):$(tag)

webshell-image:
	docker build --build-arg K8S_VERSION=$(k8sVersion) \
		--build-arg HELM_VERSION=$(helmVersion) \
		-t $(dockerhubUser)/sensecraftVoice-webshell:$(tag) -f docker/Dockerfile .

push-webshell-image: webshell-image
	docker push $(dockerhubUser)/sensecraftVoice-webshell:$(tag)

clean:
	-rm -f ./$(releaseName)

.PHONY: api-docs
api-docs: ## generate the api docs
	swag init --generalInfo ./cmd/appserver.go --output ./api/docs
