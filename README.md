# pulumi-gcp-ai-batch

[![Develop](https://github.com/davidmontoyago/pulumi-gcp-ai-batch/actions/workflows/develop.yaml/badge.svg)](https://github.com/davidmontoyago/pulumi-gcp-ai-batch/actions/workflows/develop.yaml) [![Go Coverage](https://raw.githubusercontent.com/wiki/davidmontoyago/pulumi-gcp-ai-batch/coverage.svg)](https://raw.githack.com/wiki/davidmontoyago/pulumi-gcp-ai-batch/coverage.html) [![Go Reference](https://pkg.go.dev/badge/github.com/davidmontoyago/pulumi-gcp-ai-batch.svg)](https://pkg.go.dev/github.com/davidmontoyago/pulumi-gcp-ai-batch)

Deploy a model and a job for batched inference.

See [./example](example/README.md) for an end-to-end sentiment analysis pipeline with model https://huggingface.co/nlptown/bert-base-multilingual-uncased-sentiment.

```go
pulumi.Run(func(ctx *pulumi.Context) error {
    // Launch a new async inference job with a BERT-based model
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
        OutputDataPath: pulumi.String("gs://my-bucket/predictions/"),
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

## Features

- **Batch Job Lifecycle**: launch async jobs, replace on every run or ignore old runs
- **Model Upload and Deployment**: automatic model artifacts upload to GCS and deployment to the model registry
- **Model input and outputs storage**: model inputs and outputs automatically stored in GCS
- **Service Account**: dedicated service account with necessary IAM permissions
- **Bring your own docker image**: set `ModelImageURL` to serve the model with a custom image and Custom Prediction Routines

See:
- https://cloud.google.com/blog/topics/developers-practitioners/simplify-model-serving-custom-prediction-routines-vertex-ai
- https://cloud.google.com/go/docs/reference/cloud.google.com/go/aiplatform/latest/apiv1
- https://cloud.google.com/vertex-ai/docs/model-registry/import-model
- https://github.com/davidmontoyago/pulumi-gcp-vertex-model-deployment
- https://cloud.google.com/vertex-ai/docs/predictions/pre-built-containers
- https://cloud.google.com/compute/docs/gpus#gpu-models

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

    // Required: Model configuration
    ModelDir:                          "./models/my-model",
    ModelPredictionInputSchemaPath:    "input-schema.yaml",
    ModelPredictionOutputSchemaPath:   "output-schema.yaml",
    ModelPredictionBehaviorSchemaPath: "behavior-schema.yaml", // Optional

    // Optional: Model deployment settings
    ModelImageURL:    pulumi.String("us-docker.pkg.dev/vertex-ai/prediction/tf2-cpu.2-15:latest"),
    ModelBucketBasePath: "model", // Default: "model"
    JobDisplayName:   pulumi.String("my-batch-job"),
    ModelDisplayName: pulumi.String("my-model"),

    // Input data configuration
    InputDataPath: "./inputs",     // Default: "inputs"
    InputFormat:   "jsonl",        // Default: "jsonl"

    // Output data configuration
    OutputDataPath: pulumi.String("gs://my-bucket/predictions/"), // Default: "predictions/"
    OutputFormat:   pulumi.String("jsonl"),                       // Default: "jsonl"

    // Resource allocation
    MachineType:          pulumi.String("g2-standard-8"),  // Default: "n1-standard-4"
    StartingReplicaCount: pulumi.Int(1),                   // Default: 1
    MaxReplicaCount:      pulumi.Int(3),                   // Default: 3
    BatchSize:            pulumi.Int(64),                  // Default: 0 (auto-configure)

    // Accelerator configuration (optional)
    AcceleratorType:  pulumi.String("NVIDIA_L4"),          // Default: "ACCELERATOR_TYPE_UNSPECIFIED"
    AcceleratorCount: pulumi.Int(1),                       // Default: 1

    // Network configuration (optional)
    Network: pulumi.String("projects/my-project/global/networks/my-vpc"),
    Subnet:  pulumi.String("projects/my-project/regions/us-central1/subnetworks/my-subnet"),

    // Access control
    EnablePrivateRegistryAccess: true,  // Default: false
    RetainJobOnDelete:          false, // Default: false

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
