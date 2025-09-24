// Package gcp provides Google Cloud Platform infrastructure components for Vertex AI Batch Prediction Jobs.
package gcp

import (
	"fmt"

	namer "github.com/davidmontoyago/commodity-namer"
	vertexmodeldeployment "github.com/davidmontoyago/pulumi-gcp-vertex-model-deployment/sdk/go/pulumi-gcp-vertex-model-deployment/resources"
	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/artifactregistry"
	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/storage"
	"github.com/pulumi/pulumi-google-native/sdk/go/google/aiplatform/v1beta1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// AIBatch represents a GCP Vertex AI Model Deployment and Batch Prediction Job.
type AIBatch struct {
	pulumi.ResourceState
	namer.Namer

	Project                           string
	Region                            string
	ModelImageURL                     pulumi.StringOutput
	ModelDir                          string
	ModelPredictionInputSchemaPath    string
	ModelPredictionOutputSchemaPath   string
	ModelPredictionBehaviorSchemaPath string
	ModelBucketBasePath               string
	MachineType                       pulumi.StringOutput
	JobDisplayName                    pulumi.StringOutput
	ModelDisplayName                  pulumi.StringOutput

	// Batch prediction job specific fields
	InputDataPath        pulumi.StringOutput
	InputFormat          pulumi.StringOutput
	OutputDataPath       pulumi.StringOutput
	OutputFormat         pulumi.StringOutput
	StartingReplicaCount pulumi.IntOutput
	MaxReplicaCount      pulumi.IntOutput
	BatchSize            pulumi.IntOutput
	AcceleratorType      pulumi.StringOutput
	AcceleratorCount     pulumi.IntOutput
	Labels               map[string]string

	inputDataLocalDir  string
	inputDataTargetDir string

	retainJobOnDelete bool

	// Core resources
	modelServiceAccount *serviceaccount.Account
	batchPredictionJob  *v1beta1.BatchPredictionJob
	artifactsBucket     *storage.Bucket
	modelDeployment     *vertexmodeldeployment.VertexModelDeployment
	uploadedModelFiles  pulumi.StringArrayOutput
	jobState            pulumi.StringOutput

	// IAM bindings for the model service account
	iamMembers    []*projects.IAMMember
	repoIamMember *artifactregistry.RepositoryIamMember
}

// NewAIBatch creates a new AIBatch instance with the provided configuration.
func NewAIBatch(ctx *pulumi.Context, name string, args *AIBatchArgs, opts ...pulumi.ResourceOption) (*AIBatch, error) {
	if args.Project == "" {
		return nil, fmt.Errorf("project is required")
	}
	if args.Region == "" {
		return nil, fmt.Errorf("region is required")
	}
	if args.ModelDir == "" {
		return nil, fmt.Errorf("model directory is required")
	}
	if args.ModelPredictionInputSchemaPath == "" {
		return nil, fmt.Errorf("model prediction input schema path is required")
	}
	if args.ModelPredictionOutputSchemaPath == "" {
		return nil, fmt.Errorf("model prediction output schema path is required")
	}

	if args.ModelBucketBasePath == "" {
		args.ModelBucketBasePath = "model"
	}

	// Model input data defaults
	if args.InputDataPath == "" {
		args.InputDataPath = "inputs" // Default to local "inputs" directory
	}
	if args.InputFormat == "" {
		args.InputFormat = "jsonl"
	}

	AIBatch := &AIBatch{
		Namer:                             namer.New(name, namer.WithReplace()),
		Project:                           args.Project,
		Region:                            args.Region,
		ModelDir:                          args.ModelDir,
		ModelPredictionInputSchemaPath:    args.ModelPredictionInputSchemaPath,
		ModelPredictionOutputSchemaPath:   args.ModelPredictionOutputSchemaPath,
		ModelPredictionBehaviorSchemaPath: args.ModelPredictionBehaviorSchemaPath,
		ModelBucketBasePath:               args.ModelBucketBasePath,

		// Default to the latest TensorFlow 2.15 CPU prediction container
		ModelImageURL:    setDefaultString(args.ModelImageURL, "us-docker.pkg.dev/vertex-ai/prediction/tf2-cpu.2-15:latest"),
		MachineType:      setDefaultString(args.MachineType, "n1-highmem-4"),
		JobDisplayName:   setDefaultString(args.JobDisplayName, name),
		ModelDisplayName: setDefaultString(args.ModelDisplayName, name+"-model"),

		// Model input data
		InputDataPath: pulumi.String(args.InputDataPath).ToStringOutput(),
		InputFormat:   pulumi.String(args.InputFormat).ToStringOutput(),

		// Batch prediction job specific defaults
		OutputDataPath:       setDefaultString(args.OutputDataPath, "predictions/"),
		OutputFormat:         setDefaultString(args.OutputFormat, "jsonl"),
		StartingReplicaCount: setDefaultInt(args.StartingReplicaCount, 1),
		MaxReplicaCount:      setDefaultInt(args.MaxReplicaCount, 3),
		BatchSize:            setDefaultInt(args.BatchSize, 0), // 0 means auto-configure
		AcceleratorType:      setDefaultString(args.AcceleratorType, "ACCELERATOR_TYPE_UNSPECIFIED"),
		AcceleratorCount:     setDefaultInt(args.AcceleratorCount, 1),
		Labels:               args.Labels,

		// Initial job state until we create the job
		jobState: pulumi.String("").ToStringOutput(),

		inputDataLocalDir:  args.InputDataPath,
		inputDataTargetDir: "inputs", // Upload input data to a separate "inputs" directory in bucket

		retainJobOnDelete: args.RetainJobOnDelete,
	}

	err := ctx.RegisterComponentResource("pulumi-ai-batch:gcp:AIBatch", name, AIBatch, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to register component resource: %w", err)
	}

	// Deploy the infrastructure
	err = AIBatch.deploy(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy AI batch: %w", err)
	}

	err = ctx.RegisterResourceOutputs(AIBatch, pulumi.Map{
		"vertex_ai_batch_model_service_account_email":          AIBatch.modelServiceAccount.Email,
		"vertex_ai_batch_job_id":                               AIBatch.batchPredictionJob.ID(),
		"vertex_ai_batch_job_name":                             AIBatch.batchPredictionJob.Name,
		"vertex_ai_batch_job_display_name":                     AIBatch.batchPredictionJob.DisplayName,
		"vertex_ai_batch_job_state":                            AIBatch.batchPredictionJob.State,
		"vertex_ai_batch_model_image_url":                      AIBatch.modelDeployment.ModelImageUrl,
		"vertex_ai_batch_model_artifacts_bucket_uri":           AIBatch.modelDeployment.ModelArtifactsBucketUri,
		"vertex_ai_batch_model_deployment_id":                  AIBatch.modelDeployment.ID(),
		"vertex_ai_batch_deployed_model_id":                    AIBatch.modelDeployment.DeployedModelId,
		"vertex_ai_batch_artifacts_bucket_name":                AIBatch.artifactsBucket.Name,
		"vertex_ai_batch_uploaded_model_files":                 AIBatch.uploadedModelFiles,
		"vertex_ai_batch_model_prediction_input_schema_uri":    AIBatch.modelDeployment.ModelPredictionInputSchemaUri,
		"vertex_ai_batch_model_prediction_output_schema_uri":   AIBatch.modelDeployment.ModelPredictionOutputSchemaUri,
		"vertex_ai_batch_model_prediction_behavior_schema_uri": AIBatch.modelDeployment.ModelPredictionBehaviorSchemaUri,
		"vertex_ai_batch_input_data_uri":                       AIBatch.InputDataPath,
		"vertex_ai_batch_output_data_uri_prefix":               AIBatch.OutputDataPath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register resource outputs: %w", err)
	}

	return AIBatch, nil
}

// deploy provisions all the resources for the Vertex AI Batch Prediction Job.
func (v *AIBatch) deploy(ctx *pulumi.Context, args *AIBatchArgs) error {
	// Create service account for the model deployment
	modelServiceAccount, err := v.createModelServiceAccount(ctx)
	if err != nil {
		return fmt.Errorf("failed to create model service account: %w", err)
	}
	v.modelServiceAccount = modelServiceAccount

	// Grant necessary IAM roles to the model service account
	iamMembers, err := v.grantModelIAMRoles(ctx, modelServiceAccount.Email)
	if err != nil {
		return fmt.Errorf("failed to grant model IAM roles: %w", err)
	}
	v.iamMembers = iamMembers

	if args.EnablePrivateRegistryAccess {
		v.repoIamMember, err = v.grantRegistryIAMAccess(ctx, modelServiceAccount.Email)
		if err != nil {
			return fmt.Errorf("failed to grant registry IAM access: %w", err)
		}
	}

	// Upload model artifacts (including schemas) to bucket
	modelArtifactsURI, uploadedModelArtifacts, err := v.uploadModelToBucket(ctx, args.ModelDir, args.ModelBucketBasePath, args.Labels)
	if err != nil {
		return fmt.Errorf("failed to upload model to bucket: %w", err)
	}

	// Upload input data to bucket
	inputDataBucketURI, uploadedDataObjects, err := v.uploadInputDataToBucket(ctx, v.inputDataLocalDir, v.inputDataTargetDir)
	if err != nil {
		return fmt.Errorf("failed to upload input data to bucket: %w", err)
	}

	// Collect uploaded data file names for outputs
	v.uploadedModelFiles = collectBucketObjectNames(uploadedModelArtifacts, uploadedDataObjects)

	// Deploy the model to get a model ID for the batch prediction job
	modelDeployment, err := v.deployModel(ctx, modelArtifactsURI, modelServiceAccount.Email, uploadedModelArtifacts)
	if err != nil {
		return fmt.Errorf("failed to deploy model /o\\: %w", err)
	}
	v.modelDeployment = modelDeployment

	// Create the batch prediction job
	batchPredictionJob, err := v.createBatchPredictionJob(ctx, modelDeployment, inputDataBucketURI, modelServiceAccount.Email)
	if err != nil {
		return fmt.Errorf("failed to create batch prediction job: %w", err)
	}
	// track the job state to retry on failure
	v.batchPredictionJob = batchPredictionJob
	v.jobState = batchPredictionJob.State

	return nil
}

// Getter methods for accessing internal resources

// GetModelServiceAccount returns the model service account resource.
func (v *AIBatch) GetModelServiceAccount() *serviceaccount.Account {
	return v.modelServiceAccount
}

// GetBatchPredictionJob returns the Vertex AI Batch Prediction Job resource.
func (v *AIBatch) GetBatchPredictionJob() *v1beta1.BatchPredictionJob {
	return v.batchPredictionJob
}

// GetModelDeployment returns the Vertex AI Model Deployment resource.
func (v *AIBatch) GetModelDeployment() *vertexmodeldeployment.VertexModelDeployment {
	return v.modelDeployment
}

// GetIAMMembers returns the IAM member resources.
func (v *AIBatch) GetIAMMembers() []*projects.IAMMember {
	return v.iamMembers
}

// GetUploadedModelArtifacts returns the array of uploaded model artifact names.
func (v *AIBatch) GetUploadedModelArtifacts() pulumi.StringArrayOutput {
	return v.uploadedModelFiles
}
