package gcp_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidmontoyago/pulumi-gcp-ai-batch/pkg/gcp"
)

const (
	testProjectName      = "test-project"
	testRegion           = "us-central1"
	testModelInputSchema = `
	---
type: object
properties:
  input_word_ids:
    type: array
    items:
      type: integer
      format: int32
    description: "BERT token IDs from vocabulary"
    minItems: 1
    maxItems: 512
  input_mask:
    type: array
    items:
      type: integer
      format: int32
      minimum: 0
      maximum: 1
    description: "Attention mask (1 for tokens, 0 for padding)"
    minItems: 1
    maxItems: 512
  input_type_ids:
    type: array
    items:
      type: integer
      format: int32
      minimum: 0
      maximum: 1
    description: "Segment IDs (0 for first segment, 1 for second)"
    minItems: 1
    maxItems: 512
required:
  - input_word_ids
  - input_mask
  - input_type_ids
additionalProperties: false
	`
	testModelOutputSchema = `
---
type: object
properties:
  pooled_output:
    type: array
    items:
      type: number
    description: "The pooled [CLS] token embedding representing the input sequence"
  sequence_output:
    type: array
    items:
      type: array
      items:
        type: number
    description: "Token-level embeddings for each token in the input sequence"
required:
  - pooled_output
  - sequence_output
additionalProperties: false
	`
)

type AIBatchMocks struct {
	mockFailedJob bool
	t             *testing.T
}

func (m *AIBatchMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	outputs := map[string]interface{}{}
	for k, v := range args.Inputs {
		outputs[string(k)] = v
	}

	// Mock resource outputs for each resource type:
	switch args.TypeToken {
	case "gcp:serviceaccount/account:Account":
		outputs["name"] = args.Name
		outputs["accountId"] = args.Name + "123" // Mock accountId
		outputs["project"] = testProjectName
		outputs["displayName"] = args.Name
		outputs["email"] = args.Name + "@test-project.iam.gserviceaccount.com"
		// Expected outputs: name, accountId, project, displayName, email
	case "gcp:projects/iAMMember:IAMMember":
		// Mock one of the expected roles - storage.objectViewer, logging.logWriter, or monitoring.metricWriter
		outputs["member"] = "serviceAccount:test-user@example.com"
		outputs["project"] = testProjectName
		// Expected outputs: role, member, project
	case "google-native:aiplatform/v1:BatchPredictionJob":
		outputs["name"] = args.Name
		outputs["project"] = testProjectName
		outputs["location"] = testRegion
		if m.mockFailedJob {
			outputs["state"] = "JOB_STATE_FAILED"
		} else {
			outputs["state"] = "JOB_STATE_SUCCEEDED"
		}
		outputs["createTime"] = "2023-01-01T00:00:00Z"
		// Expected outputs: name, project, location, displayName, state, createTime
	case "gcp:storage/bucket:Bucket":
		outputs["name"] = args.Name
		outputs["project"] = testProjectName
		outputs["location"] = testRegion
		outputs["forceDestroy"] = true
		outputs["uniformBucketLevelAccess"] = true
		// Expected outputs: name, project, location, forceDestroy, uniformBucketLevelAccess
	case "gcp:storage/bucketObject:BucketObject":
		outputs["selfLink"] = "https://storage.googleapis.com/storage/v1/b/test-bucket/o/test-object-" + args.Name
		// Expected outputs: name, bucket, source, contentType, selfLink
	case "gcp:projects/service:Service":
		outputs["project"] = testProjectName
		outputs["service"] = args.Inputs["service"]
		// Expected outputs: project, service
	case "gcp-vertex-model-deployment:resources:VertexModelDeployment":
		outputs["projectId"] = testProjectName
		outputs["deployedModelId"] = "test-deployed-model-id"
		outputs["modelArtifactsBucketUri"] = "gs://test-bucket"
	}

	return args.Name + "_id", resource.NewPropertyMapFromMap(outputs), nil
}

func (m *AIBatchMocks) Call(_ pulumi.MockCallArgs) (resource.PropertyMap, error) {
	// No function calls needed for basic vertex endpoint test
	return resource.PropertyMap{}, nil
}

// createTempModelDir creates a temporary directory with a dummy model file for testing
func createTempModelDir(t *testing.T) string {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "test-model-*")
	require.NoError(t, err)

	// Create a dummy model file
	modelFile := filepath.Join(tempDir, "saved_model.pb")
	err = os.WriteFile(modelFile, []byte("dummy model content"), 0600)
	require.NoError(t, err)

	// Create a variables directory with a dummy file
	varsDir := filepath.Join(tempDir, "variables")
	err = os.MkdirAll(varsDir, 0750)
	require.NoError(t, err)

	varsFile := filepath.Join(varsDir, "variables.data-00000-of-00001")
	err = os.WriteFile(varsFile, []byte("dummy variables content"), 0600)
	require.NoError(t, err)

	// Create schema files
	inputSchemaFile := filepath.Join(tempDir, "input_schema.yaml")
	err = os.WriteFile(inputSchemaFile, []byte(testModelInputSchema), 0600)
	require.NoError(t, err)

	outputSchemaFile := filepath.Join(tempDir, "output_schema.yaml")
	err = os.WriteFile(outputSchemaFile, []byte(testModelOutputSchema), 0600)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir) // Ignore cleanup errors in tests
	})

	return tempDir
}

// createTempInputDataDir creates a separate temporary directory with input data files for testing
func createTempInputDataDir(t *testing.T) string {
	t.Helper()
	inputDataDir, err := os.MkdirTemp("", "test-input-data-*")
	require.NoError(t, err)

	// Create test input data files (separate from model directory)
	inputDataFile1 := filepath.Join(inputDataDir, "data1.jsonl")
	err = os.WriteFile(inputDataFile1, []byte(`{"text": "This movie was absolutely fantastic! The acting was superb."}`), 0600)
	require.NoError(t, err)

	inputDataFile2 := filepath.Join(inputDataDir, "data2.jsonl")
	err = os.WriteFile(inputDataFile2, []byte(`{"text": "I didn't enjoy this film at all. The storyline was confusing."}`), 0600)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.RemoveAll(inputDataDir) // Ignore cleanup errors in tests
	})

	return inputDataDir
}

func TestNewAIBatch_HappyPath(t *testing.T) {
	t.Parallel()

	// Create temporary model directory for testing
	tempModelDir := createTempModelDir(t)
	// Create separate temporary input data directory
	tempInputDataDir := createTempInputDataDir(t)

	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		args := &gcp.AIBatchArgs{
			Project:                         testProjectName,
			Region:                          testRegion,
			ModelImageURL:                   pulumi.String("gcr.io/test-project/my-model:latest"),
			ModelDir:                        tempModelDir,
			ModelPredictionInputSchemaPath:  "input_schema.yaml",
			ModelPredictionOutputSchemaPath: "output_schema.yaml",
			MachineType:                     pulumi.String("n1-standard-2"),
			JobDisplayName:                  pulumi.String("test-batch-job"),
			ModelDisplayName:                pulumi.String("test-model"),
			InputDataPath:                   tempInputDataDir, // Use the separate input data directory
			InputFormat:                     "jsonl",
			OutputDataPath:                  pulumi.String("test-model/predictions/"),
			OutputFormat:                    pulumi.String("jsonl"),
			StartingReplicaCount:            pulumi.Int(1),
			MaxReplicaCount:                 pulumi.Int(3),
			BatchSize:                       pulumi.Int(100),
			Labels: map[string]string{
				"environment": "test",
				"team":        "ai-platform",
			},
		}

		AIBatch, err := gcp.NewAIBatch(ctx, "test-vertex-batch", args)
		require.NoError(t, err)

		// Verify basic properties
		assert.Equal(t, testProjectName, AIBatch.Project)
		assert.Equal(t, testRegion, AIBatch.Region)

		// Verify schema paths
		assert.Equal(t, "input_schema.yaml", AIBatch.ModelPredictionInputSchemaPath)
		assert.Equal(t, "output_schema.yaml", AIBatch.ModelPredictionOutputSchemaPath)
		assert.Empty(t, AIBatch.ModelPredictionBehaviorSchemaPath, "Behavior schema path should be empty when not provided")

		// Verify model image URL using async pattern
		modelImageCh := make(chan string, 1)
		defer close(modelImageCh)
		AIBatch.ModelImageURL.ApplyT(func(image string) error {
			modelImageCh <- image

			return nil
		})
		assert.Equal(t, "gcr.io/test-project/my-model:latest", <-modelImageCh, "Model image URL should match")

		// Verify machine type
		machineTypeCh := make(chan string, 1)
		defer close(machineTypeCh)
		AIBatch.MachineType.ApplyT(func(machineType string) error {
			machineTypeCh <- machineType

			return nil
		})
		assert.Equal(t, "n1-standard-2", <-machineTypeCh, "Machine type should match")

		// Verify batch job specific fields
		jobDisplayNameCh := make(chan string, 1)
		defer close(jobDisplayNameCh)
		AIBatch.JobDisplayName.ApplyT(func(displayName string) error {
			jobDisplayNameCh <- displayName

			return nil
		})
		assert.Equal(t, "test-batch-job", <-jobDisplayNameCh, "Job display name should match")

		inputDataURICh := make(chan string, 1)
		defer close(inputDataURICh)
		AIBatch.InputDataPath.ApplyT(func(uri string) error {
			inputDataURICh <- uri

			return nil
		})
		assert.Equal(t, tempInputDataDir, <-inputDataURICh, "Input data path should match the separate input data directory")

		outputDataURIPrefixCh := make(chan string, 1)
		defer close(outputDataURIPrefixCh)
		AIBatch.OutputDataPath.ApplyT(func(uri string) error {
			outputDataURIPrefixCh <- uri

			return nil
		})
		assert.Equal(t, "test-model/predictions/", <-outputDataURIPrefixCh, "Output data URI prefix should match")

		// Verify model service account email
		modelServiceAccountEmail := AIBatch.GetModelServiceAccountEmail()

		// Assert service account email is set correctly
		serviceAccountEmailCh := make(chan string, 1)
		defer close(serviceAccountEmailCh)
		modelServiceAccountEmail.ApplyT(func(email string) error {
			serviceAccountEmailCh <- email

			return nil
		})
		expectedEmail := "test-vertex-batch-model-account@test-project.iam.gserviceaccount.com"
		assert.Equal(t, expectedEmail, <-serviceAccountEmailCh, "Model service account email should match expected pattern")

		// Verify batch prediction job
		batchPredictionJob := AIBatch.GetBatchPredictionJob()
		require.NotNil(t, batchPredictionJob, "Batch prediction job should not be nil")

		// Verify model deployment has non-empty deployedModelId
		modelDeployment := AIBatch.GetModelDeployment()
		require.NotNil(t, modelDeployment, "Model deployment should not be nil")

		deployedModelIDCh := make(chan string, 1)
		defer close(deployedModelIDCh)
		modelDeployment.DeployedModelId.ApplyT(func(deployedModelID string) error {
			deployedModelIDCh <- deployedModelID

			return nil
		})
		deployedModelID := <-deployedModelIDCh
		assert.NotEmpty(t, deployedModelID, "Deployed model ID should not be empty")

		// Verify uploaded files (model artifacts and input data files uploaded separately)
		uploadedFiles := AIBatch.GetUploadedModelArtifacts()
		filesCh := make(chan []string, 1)
		defer close(filesCh)
		uploadedFiles.ApplyT(func(files []string) error {
			filesCh <- files

			return nil
		})
		files := <-filesCh
		require.Len(t, files, 6, "Should have uploaded exactly 6 files (4 model artifacts + 2 input data files)")

		// Verify model artifacts are in the model/ path
		expectedModelArtifacts := []string{
			"model/saved_model.pb",
			"model/variables/variables.data-00000-of-00001",
			"model/input_schema.yaml",
			"model/output_schema.yaml",
		}
		// Verify input data files are in the inputs/ path (separate from model)
		expectedInputDataFiles := []string{
			"inputs/data1.jsonl",
			"inputs/data2.jsonl",
		}

		// Verify model artifacts
		for _, expectedArtifact := range expectedModelArtifacts {
			assert.Contains(t, files, expectedArtifact, "Should contain model artifact with correct path: %s", expectedArtifact)
		}
		// Verify input data files are uploaded separately
		for _, expectedInputFile := range expectedInputDataFiles {
			assert.Contains(t, files, expectedInputFile, "Should contain input data file uploaded separately: %s", expectedInputFile)
		}

		// Verify IAM members for batch prediction job
		iamMembers := AIBatch.GetIAMMembers()
		require.Len(t, iamMembers, 5, "Should have exactly 5 IAM members (storage.bucketViewer, storage.objectCreator, logging.logWriter, monitoring.metricWriter, aiplatform.user)")

		// Check that IAM members have the expected roles
		for _, member := range iamMembers {
			roleCh := make(chan string, 1)
			member.Role.ApplyT(func(role string) error {
				roleCh <- role

				return nil
			})
			role := <-roleCh
			assert.Contains(t, []string{
				"roles/storage.bucketViewer",
				"roles/storage.objectCreator",
				"roles/logging.logWriter",
				"roles/monitoring.metricWriter",
				"roles/aiplatform.user",
			}, role, "IAM member should have expected role")
		}

		// verify that the model directory path is correctly set
		assert.Equal(t, tempModelDir, AIBatch.ModelDir, "Model directory should match the temp directory")
		assert.Equal(t, "model", AIBatch.ModelBucketBasePath, "Model bucket base path should use default value")

		// Verify input and output config URIs are properly constructed with bucket URI and paths

		// Verify input config URI construction
		batchJob := AIBatch.GetBatchPredictionJob()
		require.NotNil(t, batchJob, "Batch prediction job should not be nil")

		// Extract input config GCS URIs
		inputConfigCh := make(chan []string, 1)
		defer close(inputConfigCh)
		batchJob.InputConfig.GcsSource().Uris().ApplyT(func(uris []string) error {
			inputConfigCh <- uris

			return nil
		})
		inputConfigURIs := <-inputConfigCh
		require.Len(t, inputConfigURIs, 1, "Should have exactly one input URI")

		expectedInputURI := "gs://test-vertex-batch-vertex-model-bucket/inputs/*.jsonl"
		assert.Equal(t, expectedInputURI, inputConfigURIs[0], "Input config URI should point to inputs directory in bucket")

		// Extract output config GCS URI prefix
		outputConfigCh := make(chan string, 1)
		defer close(outputConfigCh)
		batchJob.OutputConfig.GcsDestination().OutputUriPrefix().ApplyT(func(uriPrefix string) error {
			outputConfigCh <- uriPrefix

			return nil
		})
		outputConfigURI := <-outputConfigCh

		expectedOutputURI := "gs://test-vertex-batch-vertex-model-bucket/test-model/predictions/"
		assert.Equal(t, expectedOutputURI, outputConfigURI, "Output config URI should be bucket URI + output data path")

		return nil
	}, pulumi.WithMocks("project", "stack", &AIBatchMocks{t: t}))

	if err != nil {
		t.Fatalf("Pulumi WithMocks failed: %v", err)
	}
}

func TestNewAIBatch_WithDefaults(t *testing.T) {
	t.Parallel()

	// Create temporary model directory for testing
	tempModelDir := createTempModelDir(t)
	// Create separate temporary input data directory
	tempInputDataDir := createTempInputDataDir(t)

	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		args := &gcp.AIBatchArgs{
			Project:                         testProjectName,
			Region:                          testRegion,
			ModelImageURL:                   pulumi.String("gcr.io/test-project/my-model:latest"),
			ModelDir:                        tempModelDir,
			ModelPredictionInputSchemaPath:  "input_schema.yaml",
			ModelPredictionOutputSchemaPath: "output_schema.yaml",
			MachineType:                     pulumi.String("n1-standard-2"),
			InputDataPath:                   tempInputDataDir, // Use the separate input data directory
			// Using defaults for other fields
		}

		AIBatch, err := gcp.NewAIBatch(ctx, "test-vertex-batch", args)
		require.NoError(t, err)

		// Verify defaults are applied correctly
		jobDisplayNameCh := make(chan string, 1)
		defer close(jobDisplayNameCh)
		AIBatch.JobDisplayName.ApplyT(func(displayName string) error {
			jobDisplayNameCh <- displayName

			return nil
		})
		assert.Equal(t, "test-vertex-batch", <-jobDisplayNameCh, "Job display name should default to component name")

		modelDisplayNameCh := make(chan string, 1)
		defer close(modelDisplayNameCh)
		AIBatch.ModelDisplayName.ApplyT(func(displayName string) error {
			modelDisplayNameCh <- displayName

			return nil
		})
		assert.Equal(t, "test-vertex-batch-model", <-modelDisplayNameCh, "Model display name should default to component name + '-model'")

		// Input / output data paths
		inputDataURICh := make(chan string, 1)
		defer close(inputDataURICh)
		AIBatch.InputDataPath.ApplyT(func(uri string) error {
			inputDataURICh <- uri

			return nil
		})
		assert.Equal(t, tempInputDataDir, <-inputDataURICh, "Input data path should match the separate input data directory")

		outputDataURIPrefixCh := make(chan string, 1)
		defer close(outputDataURIPrefixCh)
		AIBatch.OutputDataPath.ApplyT(func(uri string) error {
			outputDataURIPrefixCh <- uri

			return nil
		})
		assert.Equal(t, "predictions/", <-outputDataURIPrefixCh, "Output data URI prefix should match")

		inputFormatCh := make(chan string, 1)
		defer close(inputFormatCh)
		AIBatch.InputFormat.ApplyT(func(format string) error {
			inputFormatCh <- format

			return nil
		})
		assert.Equal(t, "jsonl", <-inputFormatCh, "Input format should default to 'jsonl'")

		outputFormatCh := make(chan string, 1)
		defer close(outputFormatCh)
		AIBatch.OutputFormat.ApplyT(func(format string) error {
			outputFormatCh <- format

			return nil
		})
		assert.Equal(t, "jsonl", <-outputFormatCh, "Output format should default to 'jsonl'")

		startingReplicaCountCh := make(chan int, 1)
		defer close(startingReplicaCountCh)
		AIBatch.StartingReplicaCount.ApplyT(func(count int) error {
			startingReplicaCountCh <- count

			return nil
		})
		assert.Equal(t, 1, <-startingReplicaCountCh, "Starting replica count should default to 1")

		maxReplicaCountCh := make(chan int, 1)
		defer close(maxReplicaCountCh)
		AIBatch.MaxReplicaCount.ApplyT(func(count int) error {
			maxReplicaCountCh <- count

			return nil
		})
		assert.Equal(t, 3, <-maxReplicaCountCh, "Max replica count should default to 3")

		batchSizeCh := make(chan int, 1)
		defer close(batchSizeCh)
		AIBatch.BatchSize.ApplyT(func(size int) error {
			batchSizeCh <- size

			return nil
		})
		assert.Equal(t, 0, <-batchSizeCh, "Batch size should default to 0 (auto-configure)")

		acceleratorTypeCh := make(chan string, 1)
		defer close(acceleratorTypeCh)
		AIBatch.AcceleratorType.ApplyT(func(accelType string) error {
			acceleratorTypeCh <- accelType

			return nil
		})
		assert.Equal(t, "ACCELERATOR_TYPE_UNSPECIFIED", <-acceleratorTypeCh, "Accelerator type should default to 'ACCELERATOR_TYPE_UNSPECIFIED'")

		// Verify bucket operations with defaults
		assert.Equal(t, tempModelDir, AIBatch.ModelDir, "Model directory should match the temp directory")
		assert.Equal(t, "model", AIBatch.ModelBucketBasePath, "Model bucket base path should use default value")

		// Verify uploaded files (model artifacts and input data files uploaded separately) with defaults
		uploadedFiles := AIBatch.GetUploadedModelArtifacts()
		filesCh := make(chan []string, 1)
		defer close(filesCh)
		uploadedFiles.ApplyT(func(files []string) error {
			filesCh <- files

			return nil
		})
		files := <-filesCh
		require.Len(t, files, 6, "Should have uploaded exactly 6 files (4 model artifacts + 2 input data files)")

		// Verify model artifacts are in the model/ path
		expectedModelArtifacts := []string{
			"model/saved_model.pb",
			"model/variables/variables.data-00000-of-00001",
			"model/input_schema.yaml",
			"model/output_schema.yaml",
		}
		// Verify input data files are in the inputs/ path (separate from model)
		expectedInputDataFiles := []string{
			"inputs/data1.jsonl",
			"inputs/data2.jsonl",
		}

		// Verify model artifacts
		for _, expectedArtifact := range expectedModelArtifacts {
			assert.Contains(t, files, expectedArtifact, "Should contain model artifact with correct path: %s", expectedArtifact)
		}
		// Verify input data files are uploaded separately
		for _, expectedInputFile := range expectedInputDataFiles {
			assert.Contains(t, files, expectedInputFile, "Should contain input data file uploaded separately: %s", expectedInputFile)
		}

		// Verify input and output config URIs are properly constructed with bucket URI and default paths
		modelDeployment := AIBatch.GetModelDeployment()
		require.NotNil(t, modelDeployment, "Model deployment should not be nil")

		// Verify input and output config URIs are properly constructed with bucket URI and paths

		// Verify input config URI construction
		batchJob := AIBatch.GetBatchPredictionJob()
		require.NotNil(t, batchJob, "Batch prediction job should not be nil")

		// Extract input config GCS URIs
		inputConfigCh := make(chan []string, 1)
		defer close(inputConfigCh)
		batchJob.InputConfig.GcsSource().Uris().ApplyT(func(uris []string) error {
			inputConfigCh <- uris

			return nil
		})
		inputConfigURIs := <-inputConfigCh
		require.Len(t, inputConfigURIs, 1, "Should have exactly one input URI")

		expectedInputURI := "gs://test-vertex-batch-vertex-model-bucket/inputs/*.jsonl"
		assert.Equal(t, expectedInputURI, inputConfigURIs[0], "Input config URI should point to inputs directory in bucket")

		// Extract output config GCS URI prefix
		outputConfigCh := make(chan string, 1)
		defer close(outputConfigCh)
		batchJob.OutputConfig.GcsDestination().OutputUriPrefix().ApplyT(func(uriPrefix string) error {
			outputConfigCh <- uriPrefix

			return nil
		})
		outputConfigURI := <-outputConfigCh

		expectedOutputURI := "gs://test-vertex-batch-vertex-model-bucket/predictions/"
		assert.Equal(t, expectedOutputURI, outputConfigURI, "Output config URI should be bucket URI + output data path")

		return nil
	}, pulumi.WithMocks("project", "stack", &AIBatchMocks{t: t}))

	if err != nil {
		t.Fatalf("Pulumi WithMocks failed: %v", err)
	}
}

func TestNewAIBatch_RetainJobOnDeleteAndUniqueName(t *testing.T) {
	t.Parallel()

	tempModelDir := createTempModelDir(t)
	tempInputDataDir := createTempInputDataDir(t)

	var firstJobName string

	// First run: RetainJobOnDelete = true
	retainOnDeleteRun1 := true
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {

		// Ensure some time gap at the end to simulate different runs
		defer time.Sleep(1 * time.Millisecond)

		args := &gcp.AIBatchArgs{
			Project:                         testProjectName,
			Region:                          testRegion,
			ModelDir:                        tempModelDir,
			ModelPredictionInputSchemaPath:  "input_schema.yaml",
			ModelPredictionOutputSchemaPath: "output_schema.yaml",
			InputDataPath:                   tempInputDataDir,
			RetainJobOnDelete:               retainOnDeleteRun1,
		}

		aiBatch, err := gcp.NewAIBatch(ctx, "test-retain-job", args)
		require.NoError(t, err)

		job := aiBatch.GetBatchPredictionJob()
		require.NotNil(t, job)

		// Capture the job name
		jobNameCh := make(chan string, 1)
		defer close(jobNameCh)
		job.Name.ApplyT(func(name string) error {
			jobNameCh <- name

			return nil
		})
		firstJobName = <-jobNameCh
		assert.NotEmpty(t, firstJobName, "Job name should not be empty on first run")
		assert.Contains(t, firstJobName, "test-retain-job-", "Job name should be prefixed with the component name")

		return nil
	}, pulumi.WithMocks("project", "stack", &AIBatchMocks{t: t}))
	require.NoError(t, err)

	retainOnDeleteRun2 := false
	err = pulumi.RunErr(func(ctx *pulumi.Context) error {
		args := &gcp.AIBatchArgs{
			Project:                         testProjectName,
			Region:                          testRegion,
			ModelDir:                        tempModelDir,
			ModelPredictionInputSchemaPath:  "input_schema.yaml",
			ModelPredictionOutputSchemaPath: "output_schema.yaml",
			InputDataPath:                   tempInputDataDir,
			RetainJobOnDelete:               retainOnDeleteRun2,
		}

		// same component name to test a subsequent pulumi up op
		aiBatch, err := gcp.NewAIBatch(ctx, "test-retain-job", args)
		require.NoError(t, err)

		job := aiBatch.GetBatchPredictionJob()
		require.NotNil(t, job)

		// Capture and compare the job name
		jobNameCh := make(chan string, 1)
		defer close(jobNameCh)
		job.Name.ApplyT(func(name string) error {
			jobNameCh <- name

			return nil
		})
		secondJobName := <-jobNameCh
		assert.NotEmpty(t, secondJobName, "Job name should not be empty on second run")
		assert.NotEqual(t, firstJobName, secondJobName, "Job name should be unique on each run")

		return nil
	}, pulumi.WithMocks("project", "stack", &AIBatchMocks{t: t}))
	require.NoError(t, err)
}

func TestNewAIBatch_WithModelFromTheGarden(t *testing.T) {
	t.Parallel()

	// Create separate temporary input data directory (no model directory needed for garden models)
	tempInputDataDir := createTempInputDataDir(t)

	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		args := &gcp.AIBatchArgs{
			Project:              testProjectName,
			Region:               testRegion,
			ModelName:            "publishers/google/models/gemma-2b-it", // Model from the garden
			MachineType:          pulumi.String("n1-standard-2"),
			JobDisplayName:       pulumi.String("test-garden-batch-job"),
			InputDataPath:        tempInputDataDir,
			InputFormat:          "jsonl",
			OutputDataPath:       pulumi.String("garden-model/predictions/"),
			OutputFormat:         pulumi.String("jsonl"),
			StartingReplicaCount: pulumi.Int(1),
			MaxReplicaCount:      pulumi.Int(2),
			BatchSize:            pulumi.Int(50),
			Labels: map[string]string{
				"environment": "test",
				"model-type":  "garden",
			},
		}

		AIBatch, err := gcp.NewAIBatch(ctx, "test-garden-batch", args)
		require.NoError(t, err)

		// Verify basic properties
		assert.Equal(t, testProjectName, AIBatch.Project)
		assert.Equal(t, testRegion, AIBatch.Region)
		assert.Equal(t, "publishers/google/models/gemma-2b-it", AIBatch.ModelName)
		assert.Empty(t, AIBatch.ModelDir, "Model directory should be empty for garden models")

		// Verify schema paths are empty (not required for garden models)
		assert.Empty(t, AIBatch.ModelPredictionInputSchemaPath, "Input schema path should be empty for garden models")
		assert.Empty(t, AIBatch.ModelPredictionOutputSchemaPath, "Output schema path should be empty for garden models")
		assert.Empty(t, AIBatch.ModelPredictionBehaviorSchemaPath, "Behavior schema path should be empty for garden models")

		// Verify machine type
		machineTypeCh := make(chan string, 1)
		defer close(machineTypeCh)
		AIBatch.MachineType.ApplyT(func(machineType string) error {
			machineTypeCh <- machineType

			return nil
		})
		assert.Equal(t, "n1-standard-2", <-machineTypeCh, "Machine type should match")

		// Verify batch job specific fields
		jobDisplayNameCh := make(chan string, 1)
		defer close(jobDisplayNameCh)
		AIBatch.JobDisplayName.ApplyT(func(displayName string) error {
			jobDisplayNameCh <- displayName

			return nil
		})
		assert.Equal(t, "test-garden-batch-job", <-jobDisplayNameCh, "Job display name should match")

		inputDataURICh := make(chan string, 1)
		defer close(inputDataURICh)
		AIBatch.InputDataPath.ApplyT(func(uri string) error {
			inputDataURICh <- uri

			return nil
		})
		assert.Equal(t, tempInputDataDir, <-inputDataURICh, "Input data path should match the input data directory")

		outputDataURIPrefixCh := make(chan string, 1)
		defer close(outputDataURIPrefixCh)
		AIBatch.OutputDataPath.ApplyT(func(uri string) error {
			outputDataURIPrefixCh <- uri

			return nil
		})
		assert.Equal(t, "garden-model/predictions/", <-outputDataURIPrefixCh, "Output data URI prefix should match")

		// Verify model service account email is still created
		modelServiceAccountEmail := AIBatch.GetModelServiceAccountEmail()

		// Assert service account email is set correctly
		serviceAccountEmailCh := make(chan string, 1)
		defer close(serviceAccountEmailCh)
		modelServiceAccountEmail.ApplyT(func(email string) error {
			serviceAccountEmailCh <- email

			return nil
		})
		expectedEmail := "test-garden-batch-model-account@test-project.iam.gserviceaccount.com"
		assert.Equal(t, expectedEmail, <-serviceAccountEmailCh, "Model service account email should match expected pattern")

		// Verify batch prediction job is created
		batchPredictionJob := AIBatch.GetBatchPredictionJob()
		require.NotNil(t, batchPredictionJob, "Batch prediction job should not be nil")

		// Verify NO model deployment is created for garden models
		modelDeployment := AIBatch.GetModelDeployment()
		assert.Nil(t, modelDeployment, "Model deployment should be nil for garden models")

		// Verify uploaded files (only input data files, no model artifacts)
		uploadedFiles := AIBatch.GetUploadedModelArtifacts()
		filesCh := make(chan []string, 1)
		defer close(filesCh)
		uploadedFiles.ApplyT(func(files []string) error {
			filesCh <- files

			return nil
		})
		files := <-filesCh
		require.Len(t, files, 2, "Should have uploaded exactly 2 files (only input data files, no model artifacts)")

		// Verify only input data files are uploaded (no model artifacts)
		expectedInputDataFiles := []string{
			"inputs/data1.jsonl",
			"inputs/data2.jsonl",
		}
		for _, expectedInputFile := range expectedInputDataFiles {
			assert.Contains(t, files, expectedInputFile, "Should contain input data file: %s", expectedInputFile)
		}

		// Verify no model artifacts are uploaded
		modelArtifactPatterns := []string{"model/", ".pb", ".yaml", "variables/"}
		for _, file := range files {
			for _, pattern := range modelArtifactPatterns {
				assert.NotContains(t, file, pattern, "Should not contain model artifacts for garden models: %s", file)
			}
		}

		// Verify IAM members for batch prediction job (same as regular models)
		iamMembers := AIBatch.GetIAMMembers()
		require.Len(t, iamMembers, 5, "Should have exactly 5 IAM members (storage.bucketViewer, storage.objectCreator, logging.logWriter, monitoring.metricWriter, aiplatform.user)")

		// Check that IAM members have the expected roles
		for _, member := range iamMembers {
			roleCh := make(chan string, 1)
			member.Role.ApplyT(func(role string) error {
				roleCh <- role

				return nil
			})
			role := <-roleCh
			assert.Contains(t, []string{
				"roles/storage.bucketViewer",
				"roles/storage.objectCreator",
				"roles/logging.logWriter",
				"roles/monitoring.metricWriter",
				"roles/aiplatform.user",
			}, role, "IAM member should have expected role")
		}

		// Verify input and output config URIs are properly constructed with bucket URI and paths
		batchJob := AIBatch.GetBatchPredictionJob()
		require.NotNil(t, batchJob, "Batch prediction job should not be nil")

		// Extract input config GCS URIs
		inputConfigCh := make(chan []string, 1)
		defer close(inputConfigCh)
		batchJob.InputConfig.GcsSource().Uris().ApplyT(func(uris []string) error {
			inputConfigCh <- uris

			return nil
		})
		inputConfigURIs := <-inputConfigCh
		require.Len(t, inputConfigURIs, 1, "Should have exactly one input URI")

		expectedInputURI := "gs://test-garden-batch-vertex-model-bucket/inputs/*.jsonl"
		assert.Equal(t, expectedInputURI, inputConfigURIs[0], "Input config URI should point to inputs directory in bucket")

		// Extract output config GCS URI prefix
		outputConfigCh := make(chan string, 1)
		defer close(outputConfigCh)
		batchJob.OutputConfig.GcsDestination().OutputUriPrefix().ApplyT(func(uriPrefix string) error {
			outputConfigCh <- uriPrefix

			return nil
		})
		outputConfigURI := <-outputConfigCh

		expectedOutputURI := "gs://test-garden-batch-vertex-model-bucket/garden-model/predictions/"
		assert.Equal(t, expectedOutputURI, outputConfigURI, "Output config URI should be bucket URI + output data path")

		return nil
	}, pulumi.WithMocks("project", "stack", &AIBatchMocks{t: t}))

	if err != nil {
		t.Fatalf("Pulumi WithMocks failed: %v", err)
	}
}

func TestNewAIBatch_RequiredFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        *gcp.AIBatchArgs
		expectedErr string
	}{
		{
			name: "missing project",
			args: &gcp.AIBatchArgs{
				Region:                          testRegion,
				ModelImageURL:                   pulumi.String("gcr.io/test-project/my-model:latest"),
				MachineType:                     pulumi.String("n1-standard-2"),
				ModelPredictionInputSchemaPath:  "input_schema.yaml",
				ModelPredictionOutputSchemaPath: "output_schema.yaml",
			},
			expectedErr: "project is required",
		},
		{
			name: "missing region",
			args: &gcp.AIBatchArgs{
				Project:                         testProjectName,
				ModelImageURL:                   pulumi.String("gcr.io/test-project/my-model:latest"),
				MachineType:                     pulumi.String("n1-standard-2"),
				ModelPredictionInputSchemaPath:  "input_schema.yaml",
				ModelPredictionOutputSchemaPath: "output_schema.yaml",
			},
			expectedErr: "region is required",
		},
		{
			name: "missing both model directory and model name",
			args: &gcp.AIBatchArgs{
				Project:                         testProjectName,
				Region:                          testRegion,
				ModelImageURL:                   pulumi.String("gcr.io/test-project/my-model:latest"),
				MachineType:                     pulumi.String("n1-standard-2"),
				ModelPredictionInputSchemaPath:  "input_schema.yaml",
				ModelPredictionOutputSchemaPath: "output_schema.yaml",
			},
			expectedErr: "one of model directory or model name is required",
		},
		{
			name: "missing input schema path when using model directory",
			args: &gcp.AIBatchArgs{
				Project:                         testProjectName,
				Region:                          testRegion,
				ModelImageURL:                   pulumi.String("gcr.io/test-project/my-model:latest"),
				ModelDir:                        "dummy-model-dir",
				MachineType:                     pulumi.String("n1-standard-2"),
				ModelPredictionOutputSchemaPath: "output_schema.yaml",
			},
			expectedErr: "model prediction input schema path is required",
		},
		{
			name: "missing output schema path when using model directory",
			args: &gcp.AIBatchArgs{
				Project:                        testProjectName,
				Region:                         testRegion,
				ModelImageURL:                  pulumi.String("gcr.io/test-project/my-model:latest"),
				ModelDir:                       "dummy-model-dir",
				MachineType:                    pulumi.String("n1-standard-2"),
				ModelPredictionInputSchemaPath: "input_schema.yaml",
			},
			expectedErr: "model prediction output schema path is required",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				_, err := gcp.NewAIBatch(ctx, "test-vertex-batch", testCase.args)
				if err != nil {
					assert.Contains(t, err.Error(), testCase.expectedErr)

					return nil // Expected error, test passes
				}
				t.Errorf("Expected error containing '%s', but got no error", testCase.expectedErr)

				return nil
			}, pulumi.WithMocks("project", "stack", &AIBatchMocks{t: t}))

			// We expect the test to complete successfully even when the component creation fails
			assert.NoError(t, err, "Pulumi test should not fail")
		})
	}
}
