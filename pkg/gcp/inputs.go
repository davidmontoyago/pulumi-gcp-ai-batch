// Package gcp provides Google Cloud Platform infrastructure components for Vertex AI Batch Prediction Jobs.
package gcp

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// AIBatchArgs contains configuration arguments for creating an AIBatch prediction job instance.
type AIBatchArgs struct {
	// GCP project ID where resources will be created
	Project string
	// GCP region where resources will be created (e.g., "us-central1")
	Region string
	// Container image URL for the model server. Defaults to Google's TensorFlow 2.15 CPU prediction container.
	// Example: "gcr.io/my-project/my-model:latest"
	ModelImageURL pulumi.StringInput
	// Path to the model artifacts for deployment, including the schemas. Required.
	ModelDir string
	// Path to the YAML file within ModelDir with the model prediction input schema. Required.
	ModelPredictionInputSchemaPath string
	// Path to the YAML file within ModelDir with the model prediction output schema. Required.
	ModelPredictionOutputSchemaPath string
	// Path to the YAML file within ModelDir with the model prediction behavior schema. Not required depending on the model.
	ModelPredictionBehaviorSchemaPath string
	// Base path to the model artifacts in the bucket. Defaults to "model/".
	ModelBucketBasePath string
	// Machine type for the batch prediction job (e.g., "n1-standard-2", "n1-standard-4").
	MachineType pulumi.StringInput
	// Display name for the batch prediction job (optional, defaults to component name)
	JobDisplayName pulumi.StringInput
	// Display name for the model (optional, defaults to component name + "-model")
	ModelDisplayName pulumi.StringInput

	// Batch prediction job specific fields
	// Input data configuration
	InputDataURI pulumi.StringInput // GCS URI where input data is stored (e.g., "gs://bucket/input-data.jsonl")
	InputFormat  pulumi.StringInput // Format of input data ("jsonl", "csv", "bigquery", etc.). Defaults to "jsonl"

	// Output data configuration
	OutputDataURIPrefix pulumi.StringInput // GCS URI prefix where prediction results will be stored. Defaults to same bucket under /predictions
	OutputFormat        pulumi.StringInput // Format of output data ("jsonl", "csv", "bigquery"). Defaults to "jsonl"

	// Resource allocation for batch job
	StartingReplicaCount pulumi.IntInput // Starting number of replica nodes. Defaults to 1
	MaxReplicaCount      pulumi.IntInput // Maximum number of replica nodes for scaling. Defaults to 3
	BatchSize            pulumi.IntInput // Number of instances processed per batch. Optional, auto-configured if not set

	// Compute resource specifications
	AcceleratorType  pulumi.StringInput // Type of accelerator (e.g., "NVIDIA_TESLA_T4"). Optional
	AcceleratorCount pulumi.IntInput    // Number of accelerators. Optional

	// Additional configuration
	// Additional labels to apply to resources
	Labels map[string]string
	// Network configuration for the job. Optional
	Network pulumi.StringInput
	Subnet  pulumi.StringInput
}
