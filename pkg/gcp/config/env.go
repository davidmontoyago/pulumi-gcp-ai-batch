// Package config provides an environment config helper
package config

import (
	"fmt"
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/davidmontoyago/pulumi-gcp-ai-batch/pkg/gcp"
)

// Config allows setting the vertex batch prediction job configuration via environment variables
type Config struct {
	GCPProject                        string `envconfig:"GCP_PROJECT" required:"true"`
	GCPRegion                         string `envconfig:"GCP_REGION" required:"true"`
	ModelDir                          string `envconfig:"MODEL_DIR" required:"false"`
	ModelName                         string `envconfig:"MODEL_NAME" required:"false"`
	ModelPredictionInputSchemaPath    string `envconfig:"MODEL_PREDICTION_INPUT_SCHEMA_PATH" required:"false"`
	ModelPredictionOutputSchemaPath   string `envconfig:"MODEL_PREDICTION_OUTPUT_SCHEMA_PATH" required:"false"`
	ModelPredictionBehaviorSchemaPath string `envconfig:"MODEL_PREDICTION_BEHAVIOR_SCHEMA_PATH" default:""`
	ModelBucketBasePath               string `envconfig:"MODEL_BUCKET_BASE_PATH" default:"model/"`
	ModelImageURL                     string `envconfig:"MODEL_IMAGE_URL" default:"us-docker.pkg.dev/vertex-ai/prediction/tf2-cpu.2-15:latest"`
	EnablePrivateRegistryAccess       bool   `envconfig:"ENABLE_PRIVATE_REGISTRY_ACCESS" default:"false"`
	MachineType                       string `envconfig:"MACHINE_TYPE" default:"n1-standard-2"`
	JobDisplayName                    string `envconfig:"JOB_DISPLAY_NAME" default:""`
	ModelDisplayName                  string `envconfig:"MODEL_DISPLAY_NAME" default:""`

	// Batch prediction job specific configuration
	InputDataURI         string `envconfig:"INPUT_DATA_URI" default:"inputs/"`
	InputFileName        string `envconfig:"INPUT_FILE_NAME" default:"*.jsonl"`
	InputFormat          string `envconfig:"INPUT_FORMAT" default:"jsonl"`
	OutputDataURIPrefix  string `envconfig:"OUTPUT_DATA_URI_PREFIX" default:"predictions/"`
	OutputFormat         string `envconfig:"OUTPUT_FORMAT" default:"jsonl"`
	StartingReplicaCount int    `envconfig:"STARTING_REPLICA_COUNT" default:"1"`
	MaxReplicaCount      int    `envconfig:"MAX_REPLICA_COUNT" default:"3"`
	BatchSize            int    `envconfig:"BATCH_SIZE" default:"0"`
	AcceleratorType      string `envconfig:"ACCELERATOR_TYPE" default:"ACCELERATOR_TYPE_UNSPECIFIED"`
	AcceleratorCount     int    `envconfig:"ACCELERATOR_COUNT" default:"1"`
	RetainJobOnDelete    bool   `envconfig:"RETAIN_JOB_ON_DELETE" default:"false"`
}

// LoadConfig loads configuration from environment variables
// All required environment variables must be set or will cause an error
func LoadConfig() (*Config, error) {
	var config Config

	err := envconfig.Process("", &config)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration from environment variables: %w", err)
	}

	log.Printf("Configuration loaded successfully:")
	log.Printf("  GCP Project: %s", config.GCPProject)
	log.Printf("  GCP Region: %s", config.GCPRegion)
	log.Printf("  Model Dir: %s", config.ModelDir)
	log.Printf("  Model Name: %s", config.ModelName)
	log.Printf("  Model Prediction Input Schema Path: %s", config.ModelPredictionInputSchemaPath)
	log.Printf("  Model Prediction Output Schema Path: %s", config.ModelPredictionOutputSchemaPath)
	log.Printf("  Model Prediction Behavior Schema Path: %s", config.ModelPredictionBehaviorSchemaPath)
	log.Printf("  Model Bucket Base Path: %s", config.ModelBucketBasePath)
	log.Printf("  Model Image URL: %s", config.ModelImageURL)
	log.Printf("  Enable Private Registry Access: %t", config.EnablePrivateRegistryAccess)
	log.Printf("  Machine Type: %s", config.MachineType)
	log.Printf("  Job Display Name: %s", config.JobDisplayName)
	log.Printf("  Model Display Name: %s", config.ModelDisplayName)
	log.Printf("  Input Data URI: %s", config.InputDataURI)
	log.Printf("  Input File Name: %s", config.InputFileName)
	log.Printf("  Input Format: %s", config.InputFormat)
	log.Printf("  Output Data URI Prefix: %s", config.OutputDataURIPrefix)
	log.Printf("  Output Format: %s", config.OutputFormat)
	log.Printf("  Starting Replica Count: %d", config.StartingReplicaCount)
	log.Printf("  Max Replica Count: %d", config.MaxReplicaCount)
	log.Printf("  Batch Size: %d", config.BatchSize)
	log.Printf("  Accelerator Type: %s", config.AcceleratorType)
	log.Printf("  Accelerator Count: %d", config.AcceleratorCount)
	log.Printf("  Retain Job On Delete: %t", config.RetainJobOnDelete)

	return &config, nil
}

// ToAIBatchArgs converts the config to AIBatchArgs for use with the Pulumi component
func (c *Config) ToAIBatchArgs() *gcp.AIBatchArgs {
	args := &gcp.AIBatchArgs{
		Project:                         c.GCPProject,
		Region:                          c.GCPRegion,
		ModelDir:                        c.ModelDir,
		ModelName:                       c.ModelName,
		ModelPredictionInputSchemaPath:  c.ModelPredictionInputSchemaPath,
		ModelPredictionOutputSchemaPath: c.ModelPredictionOutputSchemaPath,
		ModelBucketBasePath:             c.ModelBucketBasePath,
		ModelImageURL:                   pulumi.String(c.ModelImageURL),
		MachineType:                     pulumi.String(c.MachineType),
		EnablePrivateRegistryAccess:     c.EnablePrivateRegistryAccess,

		// Batch prediction job specific fields
		InputDataPath:        c.InputDataURI,
		InputFormat:          c.InputFormat,
		InputFileName:        c.InputFileName,
		OutputDataPath:       pulumi.String(c.OutputDataURIPrefix),
		OutputFormat:         pulumi.String(c.OutputFormat),
		StartingReplicaCount: pulumi.Int(c.StartingReplicaCount),
		MaxReplicaCount:      pulumi.Int(c.MaxReplicaCount),
		BatchSize:            pulumi.Int(c.BatchSize),
		AcceleratorType:      pulumi.String(c.AcceleratorType),
		AcceleratorCount:     pulumi.Int(c.AcceleratorCount),
		RetainJobOnDelete:    c.RetainJobOnDelete,
	}

	// Set optional fields only if provided
	if c.JobDisplayName != "" {
		args.JobDisplayName = pulumi.String(c.JobDisplayName)
	}
	if c.ModelDisplayName != "" {
		args.ModelDisplayName = pulumi.String(c.ModelDisplayName)
	}
	if c.ModelPredictionBehaviorSchemaPath != "" {
		args.ModelPredictionBehaviorSchemaPath = c.ModelPredictionBehaviorSchemaPath
	}

	return args
}
