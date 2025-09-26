MODEL_SERVER_IMAGE_NAME := pytorch-cpu.1-12-bert-with-cpr

.PHONY: build deploy clean

clean:
	go work sync
	go mod tidy

pulumi-clean:
	rm -rf ./sdk/*
	rm -rf ./build/*
	pulumi plugin rm --all -y

build: clean
	go build -o ./build/ ./...

# dockerize the model server with the custom prediction routine
image:
	# build the model server with the custom prediction routine
	docker buildx build \
		--platform=$${DOCKER_PLATFORM:-linux/amd64} \
		-t $${REGISTRY_URL}/${MODEL_SERVER_IMAGE_NAME} \
		--load \
		--file Dockerfile \
		.

pulumi-init:
	pulumi login file://.state
	pulumi stack init example

push-image:
	docker tag $${REGISTRY_URL}/${MODEL_SERVER_IMAGE_NAME}:latest $${REGISTRY_URL}/${MODEL_SERVER_IMAGE_NAME}:$${GITHUB_SHA:-latest}
	docker push $${REGISTRY_URL}/${MODEL_SERVER_IMAGE_NAME}:latest
	docker push $${REGISTRY_URL}/${MODEL_SERVER_IMAGE_NAME}:$${GITHUB_SHA:-latest}

deploy: build
	mkdir -p .state
	pulumi login file://.state
	pulumi stack select example
# 	pulumi refresh --verbose=10
	set -a && source ./config/env.example && set +a && pulumi up --color=always --verbose=1 --logtostderr --stack example

undeploy:	build
	pulumi login file://.state
	pulumi stack select example
	pulumi down --color=always --verbose=1 --logtostderr --stack example
