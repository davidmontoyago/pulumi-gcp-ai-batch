.PHONY: build clean test lint

build: clean
	go mod download
	go build ./...

test: build
	go test -v -race -count=1 -timeout=30s -coverprofile=coverage.out ./...

clean:
	go work sync
	go mod tidy
	go mod verify

lint:
	docker run --rm -v $$(pwd):/app \
		-v $$(go env GOCACHE):/.cache/go-build -e GOCACHE=/.cache/go-build \
		-v $$(go env GOMODCACHE):/.cache/mod -e GOMODCACHE=/.cache/mod \
		-w /app golangci/golangci-lint:v2.4.0 \
		golangci-lint run --fix --verbose --output.text.colors

upgrade:
	go get -u ./...

get-model-signature:
	saved_model_cli show --dir ./models/bert-tensorflow2-bert-en-uncased-l-10-h-128-a-2-v2/1/ --tag_set serve

check-model:
	saved_model_cli show --dir ./models/bert-tensorflow2-bert-en-uncased-l-10-h-128-a-2-v2/1/ --all

run-local:
	docker run -p 8501:8501 -p 8500:8500 \
		--platform=linux/amd64 \
		--mount type=bind,source=/Users/David/projects/go/src/github.com/davidmontoyago/pulumi-gcp-vertex-endpoint/models/bert-tensorflow2-bert-en-uncased-l-10-h-128-a-2-v2,target=/models/bert_model \
		-e MODEL_NAME=bert_model \
		-e AIP_STORAGE_URI=/models \
		-e AIP_HEALTH_ROUTE=/health \
		-e AIP_PREDICT_ROUTE=/predict \
		-e AIP_HTTP_PORT=8501 \
		-e TF_CPP_MIN_LOG_LEVEL=0 \
    -e TENSORFLOW_SERVING_VERBOSITY=1 \
		us-docker.pkg.dev/vertex-ai/prediction/tf2-cpu.2-15:latest
