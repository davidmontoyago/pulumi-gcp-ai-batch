package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidmontoyago/pulumi-gcp-ai-batch/pkg/gcp/config"
)

//nolint:paralleltest,tparallel // Cannot run in parallel due to global environment variable modifications
func TestLoadConfig_RequiredFields(t *testing.T) {

	// Save original environment
	origProject := os.Getenv("GCP_PROJECT")
	origRegion := os.Getenv("GCP_REGION")
	origModelDir := os.Getenv("MODEL_DIR")
	origInputSchema := os.Getenv("MODEL_PREDICTION_INPUT_SCHEMA_PATH")
	origOutputSchema := os.Getenv("MODEL_PREDICTION_OUTPUT_SCHEMA_PATH")
	origImageURL := os.Getenv("MODEL_IMAGE_URL")

	// Clean up after test
	t.Cleanup(func() {
		_ = os.Setenv("GCP_PROJECT", origProject)
		_ = os.Setenv("GCP_REGION", origRegion)
		_ = os.Setenv("MODEL_DIR", origModelDir)
		_ = os.Setenv("MODEL_PREDICTION_INPUT_SCHEMA_PATH", origInputSchema)
		_ = os.Setenv("MODEL_PREDICTION_OUTPUT_SCHEMA_PATH", origOutputSchema)
		_ = os.Setenv("MODEL_IMAGE_URL", origImageURL)
	})

	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
	}{
		{
			name: "all required fields set",
			envVars: map[string]string{
				"GCP_PROJECT":                         "test-project",
				"GCP_REGION":                          "us-central1",
				"MODEL_DIR":                           "./models/test-model",
				"MODEL_PREDICTION_INPUT_SCHEMA_PATH":  "input_schema.yaml",
				"MODEL_PREDICTION_OUTPUT_SCHEMA_PATH": "output_schema.yaml",
			},
			expectError: false,
		},
		{
			name: "missing GCP_PROJECT",
			envVars: map[string]string{
				"GCP_REGION":                          "us-central1",
				"MODEL_DIR":                           "./models/test-model",
				"MODEL_PREDICTION_INPUT_SCHEMA_PATH":  "input_schema.yaml",
				"MODEL_PREDICTION_OUTPUT_SCHEMA_PATH": "output_schema.yaml",
			},
			expectError: true,
		},
		{
			name: "missing GCP_REGION",
			envVars: map[string]string{
				"GCP_PROJECT":                         "test-project",
				"MODEL_DIR":                           "./models/test-model",
				"MODEL_PREDICTION_INPUT_SCHEMA_PATH":  "input_schema.yaml",
				"MODEL_PREDICTION_OUTPUT_SCHEMA_PATH": "output_schema.yaml",
			},
			expectError: true,
		},
		{
			name: "missing MODEL_DIR",
			envVars: map[string]string{
				"GCP_PROJECT":                         "test-project",
				"GCP_REGION":                          "us-central1",
				"MODEL_PREDICTION_INPUT_SCHEMA_PATH":  "input_schema.yaml",
				"MODEL_PREDICTION_OUTPUT_SCHEMA_PATH": "output_schema.yaml",
			},
			expectError: true,
		},
		{
			name: "missing MODEL_PREDICTION_INPUT_SCHEMA_PATH",
			envVars: map[string]string{
				"GCP_PROJECT":                         "test-project",
				"GCP_REGION":                          "us-central1",
				"MODEL_DIR":                           "./models/test-model",
				"MODEL_PREDICTION_OUTPUT_SCHEMA_PATH": "output_schema.yaml",
			},
			expectError: true,
		},
		{
			name: "missing MODEL_PREDICTION_OUTPUT_SCHEMA_PATH",
			envVars: map[string]string{
				"GCP_PROJECT":                        "test-project",
				"GCP_REGION":                         "us-central1",
				"MODEL_DIR":                          "./models/test-model",
				"MODEL_PREDICTION_INPUT_SCHEMA_PATH": "input_schema.yaml",
			},
			expectError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Note: We cannot run these tests in parallel because they modify global environment variables
			// Clear environment
			_ = os.Unsetenv("GCP_PROJECT")
			_ = os.Unsetenv("GCP_REGION")
			_ = os.Unsetenv("MODEL_DIR")
			_ = os.Unsetenv("MODEL_PREDICTION_INPUT_SCHEMA_PATH")
			_ = os.Unsetenv("MODEL_PREDICTION_OUTPUT_SCHEMA_PATH")
			_ = os.Unsetenv("MODEL_IMAGE_URL")

			// Set test environment variables
			for key, value := range testCase.envVars {
				_ = os.Setenv(key, value)
			}

			cfg, err := config.LoadConfig()

			if testCase.expectError {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, cfg)
				assert.Equal(t, "test-project", cfg.GCPProject)
				assert.Equal(t, "us-central1", cfg.GCPRegion)
				assert.Equal(t, "./models/test-model", cfg.ModelDir)
				assert.Equal(t, "input_schema.yaml", cfg.ModelPredictionInputSchemaPath)
				assert.Equal(t, "output_schema.yaml", cfg.ModelPredictionOutputSchemaPath)
				assert.Equal(t, "us-docker.pkg.dev/vertex-ai/prediction/tf2-cpu.2-15:latest", cfg.ModelImageURL)
			}
		})
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	t.Parallel()

	// Save original environment
	origProject := os.Getenv("GCP_PROJECT")
	origRegion := os.Getenv("GCP_REGION")
	origModelDir := os.Getenv("MODEL_DIR")
	origInputSchema := os.Getenv("MODEL_PREDICTION_INPUT_SCHEMA_PATH")
	origOutputSchema := os.Getenv("MODEL_PREDICTION_OUTPUT_SCHEMA_PATH")
	origBehaviorSchema := os.Getenv("MODEL_PREDICTION_BEHAVIOR_SCHEMA_PATH")
	origBucketBasePath := os.Getenv("MODEL_BUCKET_BASE_PATH")
	origImageURL := os.Getenv("MODEL_IMAGE_URL")

	// Clean up after test
	defer func() {
		_ = os.Setenv("GCP_PROJECT", origProject)
		_ = os.Setenv("GCP_REGION", origRegion)
		_ = os.Setenv("MODEL_DIR", origModelDir)
		_ = os.Setenv("MODEL_PREDICTION_INPUT_SCHEMA_PATH", origInputSchema)
		_ = os.Setenv("MODEL_PREDICTION_OUTPUT_SCHEMA_PATH", origOutputSchema)
		_ = os.Setenv("MODEL_PREDICTION_BEHAVIOR_SCHEMA_PATH", origBehaviorSchema)
		_ = os.Setenv("MODEL_BUCKET_BASE_PATH", origBucketBasePath)
		_ = os.Setenv("MODEL_IMAGE_URL", origImageURL)
	}()

	// Set only required fields
	_ = os.Setenv("GCP_PROJECT", "test-project")
	_ = os.Setenv("GCP_REGION", "us-central1")
	_ = os.Setenv("MODEL_DIR", "./models/test-model")
	_ = os.Setenv("MODEL_PREDICTION_INPUT_SCHEMA_PATH", "input_schema.yaml")
	_ = os.Setenv("MODEL_PREDICTION_OUTPUT_SCHEMA_PATH", "output_schema.yaml")

	// Clear optional fields to test defaults
	_ = os.Unsetenv("MODEL_PREDICTION_BEHAVIOR_SCHEMA_PATH")
	_ = os.Unsetenv("MODEL_BUCKET_BASE_PATH")
	_ = os.Unsetenv("MODEL_IMAGE_URL")
	_ = os.Unsetenv("MACHINE_TYPE")
	_ = os.Unsetenv("JOB_DISPLAY_NAME")
	_ = os.Unsetenv("INPUT_DATA_URI")
	_ = os.Unsetenv("INPUT_FORMAT")
	_ = os.Unsetenv("OUTPUT_DATA_URI_PREFIX")
	_ = os.Unsetenv("OUTPUT_FORMAT")
	_ = os.Unsetenv("STARTING_REPLICA_COUNT")
	_ = os.Unsetenv("MAX_REPLICA_COUNT")
	_ = os.Unsetenv("BATCH_SIZE")
	_ = os.Unsetenv("ACCELERATOR_TYPE")
	_ = os.Unsetenv("ACCELERATOR_COUNT")
	_ = os.Unsetenv("NETWORK")
	_ = os.Unsetenv("SUBNET")

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify required fields are set
	assert.Equal(t, "test-project", cfg.GCPProject)
	assert.Equal(t, "us-central1", cfg.GCPRegion)
	assert.Equal(t, "./models/test-model", cfg.ModelDir)
	assert.Equal(t, "input_schema.yaml", cfg.ModelPredictionInputSchemaPath)
	assert.Equal(t, "output_schema.yaml", cfg.ModelPredictionOutputSchemaPath)

	// Verify defaults
	assert.Equal(t, "", cfg.ModelPredictionBehaviorSchemaPath)
	assert.Equal(t, "model/", cfg.ModelBucketBasePath)
	assert.Equal(t, "us-docker.pkg.dev/vertex-ai/prediction/tf2-cpu.2-15:latest", cfg.ModelImageURL)
	assert.Equal(t, "n1-standard-2", cfg.MachineType)
	assert.Equal(t, "", cfg.JobDisplayName)
	assert.Equal(t, "", cfg.InputDataURI)
	assert.Equal(t, "jsonl", cfg.InputFormat)
	assert.Equal(t, "", cfg.OutputDataURIPrefix)
	assert.Equal(t, "jsonl", cfg.OutputFormat)
	assert.Equal(t, 1, cfg.StartingReplicaCount)
	assert.Equal(t, 3, cfg.MaxReplicaCount)
	assert.Equal(t, 0, cfg.BatchSize)
	assert.Equal(t, "", cfg.AcceleratorType)
	assert.Equal(t, 0, cfg.AcceleratorCount)
	assert.Equal(t, "", cfg.Network)
	assert.Equal(t, "", cfg.Subnet)
}

func TestToAIBatchArgs(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		GCPProject:                        "test-project",
		GCPRegion:                         "us-central1",
		ModelDir:                          "./models/test-model",
		ModelPredictionInputSchemaPath:    "input_schema.yaml",
		ModelPredictionOutputSchemaPath:   "output_schema.yaml",
		ModelPredictionBehaviorSchemaPath: "behavior_schema.yaml",
		ModelBucketBasePath:               "models/v1/",
		ModelImageURL:                     "gcr.io/test/model:latest",
		MachineType:                       "n1-standard-4",
		JobDisplayName:                    "test-batch-job",
		ModelDisplayName:                  "test-model",
		InputDataURI:                      "gs://test-bucket/input-data.jsonl",
		InputFormat:                       "jsonl",
		OutputDataURIPrefix:               "gs://test-bucket/predictions",
		OutputFormat:                      "jsonl",
		StartingReplicaCount:              2,
		MaxReplicaCount:                   5,
		BatchSize:                         100,
		AcceleratorType:                   "NVIDIA_TESLA_T4",
		AcceleratorCount:                  1,
		Network:                           "projects/test-project/global/networks/test-network",
		Subnet:                            "projects/test-project/regions/us-central1/subnetworks/test-subnet",
	}

	args := cfg.ToAIBatchArgs()
	require.NotNil(t, args)

	// Verify required fields
	assert.Equal(t, "test-project", args.Project)
	assert.Equal(t, "us-central1", args.Region)
	assert.Equal(t, "./models/test-model", args.ModelDir)
	assert.Equal(t, "input_schema.yaml", args.ModelPredictionInputSchemaPath)
	assert.Equal(t, "output_schema.yaml", args.ModelPredictionOutputSchemaPath)
	assert.Equal(t, "behavior_schema.yaml", args.ModelPredictionBehaviorSchemaPath)
	assert.Equal(t, "models/v1/", args.ModelBucketBasePath)

	// For this simple test, we just verify the args are set correctly
	// The actual Pulumi resource creation is tested in the main component tests
	require.NotNil(t, args.ModelImageURL)
	require.NotNil(t, args.MachineType)
	require.NotNil(t, args.JobDisplayName)
	require.NotNil(t, args.ModelDisplayName)
	require.NotNil(t, args.InputDataPath)
	require.NotNil(t, args.InputFormat)
	require.NotNil(t, args.OutputDataPath)
	require.NotNil(t, args.OutputFormat)
	require.NotNil(t, args.StartingReplicaCount)
	require.NotNil(t, args.MaxReplicaCount)
	require.NotNil(t, args.BatchSize)
	require.NotNil(t, args.AcceleratorType)
	require.NotNil(t, args.AcceleratorCount)
	require.NotNil(t, args.Network)
	require.NotNil(t, args.Subnet)
}
