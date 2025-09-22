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
	// Base path to the model artifacts in the bucket. Defaults to "model".
	ModelBucketBasePath string
	// Machine type for the batch prediction job (e.g., "n1-standard-2", "n1-standard-4").
	MachineType pulumi.StringInput
	// Display name for the batch prediction job (optional, defaults to component name)
	JobDisplayName pulumi.StringInput
	// Display name for the model (optional, defaults to component name + "-model")
	ModelDisplayName pulumi.StringInput
	// If true, the model Service Account is granted access to the Artifact Registry repository in ModelImageURL.
	EnablePrivateRegistryAccess bool
	// Every pulumi up operation is a new job launch with a unique name.
	// Set this to true to retain jobs in between runs, and ensure old jobs are
	// eventually cleaned up.
	// If not set, the job will be replaced regardless of the state.
	RetainJobOnDelete bool

	// Batch prediction job specific fields

	// --- Input data configuration ---
	// Path to the local directory containing input data files (e.g., "data/inputs/")
	// This directory is SEPARATE from the model directory and contains the actual input data
	// that will be processed by the batch prediction job. Files will be uploaded to the
	// bucket separately from model artifacts. Defaults to "inputs".
	// Input data files will be uploaded to the bucket under the "inputs" directory.
	InputDataPath string
	// Format of input data ("jsonl", "csv", "bigquery", etc.). Defaults to "jsonl"
	InputFormat string

	// --- Output data configuration ---
	// Path to the directory within the bucket where the output data will be stored.
	// Defaults to "/predictions"
	OutputDataPath pulumi.StringInput
	// Format of output data ("jsonl", "csv", "bigquery"). Defaults to "jsonl"
	OutputFormat pulumi.StringInput

	// Resource allocation for batch job
	// Starting number of replica nodes. Defaults to 1
	StartingReplicaCount pulumi.IntInput
	// Maximum number of replica nodes for scaling. Defaults to 3
	MaxReplicaCount pulumi.IntInput
	// Number of instances processed per batch. Optional, auto-configured if not set
	BatchSize pulumi.IntInput

	// Compute resource specifications
	// Type of accelerator (e.g., "NVIDIA_TESLA_T4"). Optional.
	// Defaults to "ACCELERATOR_TYPE_UNSPECIFIED"
	AcceleratorType pulumi.StringInput
	// Number of accelerators. Defaults to 1
	AcceleratorCount pulumi.IntInput

	// Additional configuration
	// Additional labels to apply to resources
	Labels map[string]string
	// Network configuration for the job. Optional
	Network pulumi.StringInput
	Subnet  pulumi.StringInput
}
