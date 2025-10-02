# pulumi-gcp-ai-batch

[![Develop](https://github.com/davidmontoyago/pulumi-gcp-ai-batch/actions/workflows/develop.yaml/badge.svg)](https://github.com/davidmontoyago/pulumi-gcp-ai-batch/actions/workflows/develop.yaml) [![Go Coverage](https://raw.githubusercontent.com/wiki/davidmontoyago/pulumi-gcp-ai-batch/coverage.svg)](https://raw.githack.com/wiki/davidmontoyago/pulumi-gcp-ai-batch/coverage.html) [![Go Reference](https://pkg.go.dev/badge/github.com/davidmontoyago/pulumi-gcp-ai-batch.svg)](https://pkg.go.dev/github.com/davidmontoyago/pulumi-gcp-ai-batch)

Deploy a model and a job for batched inference in Vertex AI.

## Features

- **Batch Job Lifecycle**: launch async jobs, replace on every run or ignore old runs
- **Model Upload and Deployment**: automatic model artifacts upload to GCS and deployment to the model registry
- **Model input and outputs storage**: model inputs and outputs automatically stored in GCS
- **Service Account**: dedicated service account with necessary IAM permissions (not required for garden models)
- **Bring your own docker image**: set `ModelImageURL` to serve the model with a custom image and Custom Prediction Routines


## Deploy model from the model garden
```go
pulumi.Run(func(ctx *pulumi.Context) error {
    // Launch a new async inference job with a LLama-based model
    batchJob, err := gcp.NewAIBatch(ctx, "llama-sentiment-batch", &gcp.AIBatchArgs{
        Project: "my-gcp-project",
        Region:  "us-central1",

        // Model configuration
        ModelName:     "publishers/meta/models/llama3-2@llama-3.2-3b-instruct",

        // Input data configuration
        InputDataPath: "./inputs",
        InputFormat:   "jsonl",

        // Output configuration
        OutputDataPath: pulumi.String("my-predictions/"),
        OutputFormat:   pulumi.String("jsonl"),

        // Resource allocation
        MachineType:          pulumi.String("g2-standard-8"),
        AcceleratorType:      pulumi.String("NVIDIA_L4"),
        AcceleratorCount:     pulumi.Int(1),

        // Metadata
        Labels: map[string]string{
            "environment": "production",
            "model-type":  "llama",
            "use-case":    "sentiment-analysis",
        },
    })
    if err != nil {
        return err
    }

    // Export useful outputs
    ctx.Export("batchJobName", batchJob.GetBatchPredictionJob().Name)

    return nil
})
```

## Deploy custom model
```go
pulumi.Run(func(ctx *pulumi.Context) error {
    // Launch a new async inference job with a BERT-based model downloaded from Hugging Face
    batchJob, err := gcp.NewAIBatch(ctx, "bert-sentiment-batch", &gcp.AIBatchArgs{
        Project: "my-gcp-project",
        Region:  "us-central1",

        // Model configuration
        ModelDir:                          "./models/nlptown-bert-base-multilingual-uncased-sentiment",
        ModelPredictionInputSchemaPath:    "bert-instance-schema.yaml",
        ModelPredictionOutputSchemaPath:   "bert-prediction-schema.yaml",
        ModelPredictionBehaviorSchemaPath: "bert-parameters-schema.yaml",
        ModelImageURL:                     pulumi.String("us-docker.pkg.dev/vertex-ai/prediction/tf2-cpu.2-15:latest"),

        // Input data configuration
        InputDataPath: "./inputs",
        InputFormat:   "jsonl",

        // Output configuration
        OutputDataPath: pulumi.String("predictions/"),
        OutputFormat:   pulumi.String("jsonl"),

        // Resource allocation
        MachineType:          pulumi.String("g2-standard-8"),
        StartingReplicaCount: pulumi.Int(2),
        MaxReplicaCount:      pulumi.Int(5),
        BatchSize:            pulumi.Int(64),

        // Optional: GPU acceleration
        AcceleratorType:  pulumi.String("NVIDIA_L4"),
        AcceleratorCount: pulumi.Int(1),

        // Metadata
        Labels: map[string]string{
            "environment": "production",
            "model-type":  "bert",
            "use-case":    "sentiment-analysis",
        },
    })
    if err != nil {
        return err
    }

    // Export useful outputs
    ctx.Export("batchJobName", batchJob.GetBatchPredictionJob().Name)
    ctx.Export("modelServiceAccount", batchJob.GetModelServiceAccount().Email)

    return nil
})
```

## Model deployment examples
See working end-to-end model deployments:
| Example                                                                      | Model Source      | README                                                                                 |
| ---------------------------------------------------------------------------- | ----------------- | -------------------------------------------------------------------------------------- |
| Sentiment analysis with Fine tuned BERT model with custom prediction routine | HuggingFace Model | [examples/bert-sentiment-analysis-with-cpr](examples/bert-sentiment-analysis-with-cpr) |
| Sentiment analysis with out of the box Llama from the Model Garden           | GCP Model Garden  | [examples/llama-sentiment-analysis](examples/llama-sentiment-analysis)                 |
| Code change review with Mistral model from the Model Garden                  | GCP Model Garden  | [examples/mistral-code-change-review](examples/mistral-code-change-review)             |


See:
- Third party models with custom prediction routines: https://cloud.google.com/blog/topics/developers-practitioners/simplify-model-serving-custom-prediction-routines-vertex-ai
- Garden models supported: https://cloud.google.com/vertex-ai/docs/predictions/get-batch-model-garden#models
- https://cloud.google.com/compute/docs/gpus#gpu-models
- https://cloud.google.com/go/docs/reference/cloud.google.com/go/aiplatform/latest/apiv1
- https://cloud.google.com/vertex-ai/docs/model-registry/import-model
- https://github.com/davidmontoyago/pulumi-gcp-vertex-model-deployment
- https://cloud.google.com/vertex-ai/docs/predictions/pre-built-containers

## Install

```bash
go get github.com/davidmontoyago/pulumi-gcp-ai-batch
```

### Full Config

```go
args := &gcp.AIBatchArgs{
    // Required: GCP project and region
    Project: "my-gcp-project",
    Region:  "us-central1",

    // Model configuration - use either ModelDir OR ModelName
    // Option 1: Custom model with artifacts
    ModelDir:                            "./models/my-model",
    ModelImageURL:                       pulumi.String("us-docker.pkg.dev/vertex-ai/prediction/tf2-cpu.2-15:latest"),
    ModelPredictionInputSchemaPath:      "input-schema.yaml",
    ModelPredictionOutputSchemaPath:     "output-schema.yaml",
    ModelPredictionBehaviorSchemaPath:   "behavior-schema.yaml", // Optional
    ModelBucketBasePath:                 "model", // Default: "model"

    // Option 2: Model garden model (alternative to ModelDir)
    // ModelName: "publishers/google/models/gemma2@gemma-2-2b-it",

    // Display names
    JobDisplayName:   pulumi.String("my-batch-job"),
    ModelDisplayName: pulumi.String("my-model"),

    // Input data configuration
    InputDataPath: "inputs",     // Default: "inputs"
    InputFormat:   "jsonl",      // Default: "jsonl"
    InputFileName: "data.jsonl", // Default: "*.jsonl"

    // Output data configuration
    OutputDataPath: pulumi.String("predictions"), // Default: "predictions"
    OutputFormat:   pulumi.String("jsonl"),       // Default: "jsonl"

    // Resource allocation
    MachineType:          pulumi.String("n1-standard-4"), // Default: "n1-standard-4"
    StartingReplicaCount: pulumi.Int(1),                  // Default: 1
    MaxReplicaCount:      pulumi.Int(3),                  // Default: 3
    BatchSize:            pulumi.Int(64),                 // Optional, auto-configured if not set

    // Accelerator configuration (optional)
    AcceleratorType:  pulumi.String("NVIDIA_TESLA_T4"), // Default: "ACCELERATOR_TYPE_UNSPECIFIED"
    AcceleratorCount: pulumi.Int(1),                     // Default: 1

    // Access control
    EnablePrivateRegistryAccess: true,  // Default: false
    RetainJobOnDelete:           false, // Default: false

    // Metadata
    Labels: map[string]string{
        "environment": "production",
        "team":        "ml-ops",
        "cost-center": "research",
    },
}
```

## Development

- **Build**: `make build`
- **Test**: `make test`
- **Lint**: `make lint`
- **Clean**: `make clean`

## Requirements

- Go 1.24+
- GCP project with Vertex AI API enabled
- Pulumi CLI
