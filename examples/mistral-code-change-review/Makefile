EXAMPLE_STACK_NAME := organization/mistral-code-change-review/example

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

pulumi-init:
	mkdir -p .state
	pulumi login file://.state
	pulumi stack init "${EXAMPLE_STACK_NAME}" 2>/dev/null; true

deploy: build
	mkdir -p .state
	pulumi login file://.state
	pulumi stack select "${EXAMPLE_STACK_NAME}"
	pulumi refresh
	set -a && source ./config/env.example && set +a && pulumi up --color=always --verbose=1 --logtostderr --stack ${EXAMPLE_STACK_NAME}

undeploy:	build
	pulumi login file://.state
	pulumi stack select "${EXAMPLE_STACK_NAME}"
	pulumi down --color=always --verbose=1 --logtostderr --stack "${EXAMPLE_STACK_NAME}"
