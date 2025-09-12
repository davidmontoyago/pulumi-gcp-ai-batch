// Package gcp provides Google Cloud Platform infrastructure components for Vertex AI Batch Prediction Jobs.
package gcp

import (
	"fmt"

	namer "github.com/davidmontoyago/commodity-namer"
	vertexmodeldeployment "github.com/davidmontoyago/pulumi-gcp-vertex-model-deployment/sdk/go/pulumi-gcp-vertex-model-deployment/resources"
	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/storage"
	"github.com/pulumi/pulumi-google-native/sdk/go/google/aiplatform/v1beta1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// AIBatch represents a GCP Vertex AI Batch Prediction Job.
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
	InputDataURI         pulumi.StringOutput
	InputFormat          pulumi.StringOutput
	OutputDataURIPrefix  pulumi.StringOutput
	OutputFormat         pulumi.StringOutput
	StartingReplicaCount pulumi.IntOutput
	MaxReplicaCount      pulumi.IntOutput
	BatchSize            pulumi.IntOutput
	AcceleratorType      pulumi.StringOutput
	AcceleratorCount     pulumi.IntOutput
	Network              pulumi.StringOutput
	Subnet               pulumi.StringOutput
	Labels               map[string]string

	name string

	// Core resources
	modelServiceAccount    *serviceaccount.Account
	batchPredictionJob     *v1beta1.BatchPredictionJob
	artifactsBucket        *storage.Bucket
	modelDeployment        *vertexmodeldeployment.VertexModelDeployment
	uploadedModelArtifacts pulumi.StringArrayOutput

	// IAM bindings for the model service account
	iamMembers []*projects.IAMMember
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
		args.ModelBucketBasePath = "model/"
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

		// Batch prediction job specific defaults
		InputDataURI:         setDefaultString(args.InputDataURI, ""),
		InputFormat:          setDefaultString(args.InputFormat, "jsonl"),
		OutputDataURIPrefix:  setDefaultString(args.OutputDataURIPrefix, ""),
		OutputFormat:         setDefaultString(args.OutputFormat, "jsonl"),
		StartingReplicaCount: setDefaultInt(args.StartingReplicaCount, 1),
		MaxReplicaCount:      setDefaultInt(args.MaxReplicaCount, 3),
		BatchSize:            setDefaultInt(args.BatchSize, 0), // 0 means auto-configure
		AcceleratorType:      setDefaultString(args.AcceleratorType, ""),
		AcceleratorCount:     setDefaultInt(args.AcceleratorCount, 0),
		Network:              setDefaultString(args.Network, ""),
		Subnet:               setDefaultString(args.Subnet, ""),
		Labels:               args.Labels,

		name: name,
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
		"vertex_ai_batch_uploaded_model_files":                 AIBatch.uploadedModelArtifacts,
		"vertex_ai_batch_model_prediction_input_schema_uri":    AIBatch.modelDeployment.ModelPredictionInputSchemaUri,
		"vertex_ai_batch_model_prediction_output_schema_uri":   AIBatch.modelDeployment.ModelPredictionOutputSchemaUri,
		"vertex_ai_batch_model_prediction_behavior_schema_uri": AIBatch.modelDeployment.ModelPredictionBehaviorSchemaUri,
		"vertex_ai_batch_input_data_uri":                       AIBatch.InputDataURI,
		"vertex_ai_batch_output_data_uri_prefix":               AIBatch.OutputDataURIPrefix,
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

	// Upload model artifacts (including schemas) to bucket
	modelArtifactsURI, uploadedObjects, err := v.uploadModelToBucket(ctx, args.ModelDir, args.ModelBucketBasePath, args.Labels)
	if err != nil {
		return fmt.Errorf("failed to upload model to bucket: %w", err)
	}

	// Collect uploaded object names for tracking
	uploadedObjectNames := pulumi.StringArray{}
	for _, resource := range uploadedObjects {
		if bucketObject, ok := resource.(*storage.BucketObject); ok {
			uploadedObjectNames = append(uploadedObjectNames, bucketObject.Name.ApplyT(func(name string) string {
				return name
			}).(pulumi.StringOutput))
		}
	}
	v.uploadedModelArtifacts = uploadedObjectNames.ToStringArrayOutput()

	// Deploy the model to get a model ID for the batch prediction job
	modelDeployment, err := v.deployModel(ctx, modelArtifactsURI, modelServiceAccount.Email, uploadedObjects)
	if err != nil {
		return fmt.Errorf("failed to deploy model /o\\: %w", err)
	}
	v.modelDeployment = modelDeployment

	// Set default output URI prefix to same bucket under /predictions if not specified
	v.OutputDataURIPrefix = pulumi.All(v.OutputDataURIPrefix, modelArtifactsURI).ApplyT(func(args []interface{}) string {
		uri := args[0].(string)
		bucketURI := args[1].(string)
		if uri == "" {
			return fmt.Sprintf("%s/predictions", bucketURI)
		}
		return uri
	}).(pulumi.StringOutput)

	// Create the batch prediction job
	batchPredictionJob, err := v.createBatchPredictionJob(ctx, modelDeployment, modelServiceAccount.Email)
	if err != nil {
		return fmt.Errorf("failed to create batch prediction job: %w", err)
	}
	v.batchPredictionJob = batchPredictionJob

	return nil
}

// deployModel deploys the model to Vertex AI
// for batch prediction jobs, we only need the model, not an endpoint
func (v *AIBatch) deployModel(ctx *pulumi.Context, modelArtifactsURI pulumi.StringOutput, serviceAccountEmail pulumi.StringOutput, uploadedObjects []pulumi.Resource) (*vertexmodeldeployment.VertexModelDeployment, error) {
	modelDeploymentArgs := &vertexmodeldeployment.VertexModelDeploymentArgs{
		ProjectId:                      pulumi.String(v.Project),
		Region:                         pulumi.String(v.Region),
		ModelArtifactsBucketUri:        modelArtifactsURI,
		ModelImageUrl:                  v.ModelImageURL,
		ModelPredictionInputSchemaUri:  pulumi.Sprintf("%s/%s", modelArtifactsURI, v.ModelPredictionInputSchemaPath),
		ModelPredictionOutputSchemaUri: pulumi.Sprintf("%s/%s", modelArtifactsURI, v.ModelPredictionOutputSchemaPath),
		MachineType:                    v.MachineType,
		MinReplicas:                    v.StartingReplicaCount,
		MaxReplicas:                    v.MaxReplicaCount,
		ServiceAccount:                 serviceAccountEmail,
		// For batch prediction jobs, we don't need an endpoint, just the model
		EndpointId: pulumi.String(""),
	}
	if v.ModelPredictionBehaviorSchemaPath != "" {
		modelDeploymentArgs.ModelPredictionBehaviorSchemaUri = pulumi.Sprintf("%s/%s", modelArtifactsURI, v.ModelPredictionBehaviorSchemaPath)
	}

	// Include dependencies on both the artifacts bucket and uploaded model artifacts
	dependencies := []pulumi.Resource{v.artifactsBucket}
	dependencies = append(dependencies, uploadedObjects...)

	return vertexmodeldeployment.NewVertexModelDeployment(ctx,
		v.NewResourceName("vertex-model-deployment", "", 63),
		modelDeploymentArgs,
		pulumi.Parent(v),
		pulumi.DependsOn(dependencies),
	)
}

// createModelServiceAccount creates a service account for Vertex AI operations.
func (v *AIBatch) createModelServiceAccount(ctx *pulumi.Context) (*serviceaccount.Account, error) {
	accountID := v.NewResourceName("model-sa", "service-account", 30)

	return serviceaccount.NewAccount(ctx, v.NewResourceName("model-sa", "service-account", 63), &serviceaccount.AccountArgs{
		Project:     pulumi.String(v.Project),
		AccountId:   pulumi.String(accountID),
		DisplayName: pulumi.Sprintf("%s Vertex AI Service Account", v.ModelDisplayName),
		Description: pulumi.String("Service account for deployed model operations"),
	}, pulumi.Parent(v))
}

// grantModelIAMRoles grants necessary IAM roles to the model service account.
func (v *AIBatch) grantModelIAMRoles(ctx *pulumi.Context, serviceAccountEmail pulumi.StringOutput) ([]*projects.IAMMember, error) {
	// IAM roles specific to what the batch prediction job needs to operate
	roles := []string{
		"roles/storage.bucketViewer",    // List and get buckets
		"roles/storage.objectViewer",    // For accessing model artifacts in GCS
		"roles/storage.objectCreator",   // For writing prediction results to GCS
		"roles/logging.logWriter",       // For writing logs during prediction
		"roles/monitoring.metricWriter", // For writing custom metrics
		"roles/aiplatform.user",         // For accessing Vertex AI resources
	}

	iamMembers := make([]*projects.IAMMember, len(roles))
	for roleIndex, role := range roles {
		bindingName := v.NewResourceName(fmt.Sprintf("model-sa-iam-%s", role), "", 63)
		member, err := projects.NewIAMMember(ctx, bindingName, &projects.IAMMemberArgs{
			Project: pulumi.String(v.Project),
			Role:    pulumi.String(role),
			Member:  pulumi.Sprintf("serviceAccount:%s", serviceAccountEmail),
		}, pulumi.Parent(v))
		if err != nil {
			return nil, fmt.Errorf("failed to create IAM member for role %s: %w", role, err)
		}
		iamMembers[roleIndex] = member
	}

	return iamMembers, nil
}

// createBatchPredictionJob creates a Vertex AI Batch Prediction Job.
func (v *AIBatch) createBatchPredictionJob(ctx *pulumi.Context, modelDeployment *vertexmodeldeployment.VertexModelDeployment, serviceAccountEmail pulumi.StringOutput) (*v1beta1.BatchPredictionJob, error) {
	// Construct the input config
	inputConfig := &v1beta1.GoogleCloudAiplatformV1beta1BatchPredictionJobInputConfigArgs{
		InstancesFormat: v.InputFormat,
		GcsSource: &v1beta1.GoogleCloudAiplatformV1beta1GcsSourceArgs{
			Uris: pulumi.StringArray{v.InputDataURI},
		},
	}

	// Construct the output config
	outputConfig := &v1beta1.GoogleCloudAiplatformV1beta1BatchPredictionJobOutputConfigArgs{
		PredictionsFormat: v.OutputFormat,
		GcsDestination: &v1beta1.GoogleCloudAiplatformV1beta1GcsDestinationArgs{
			OutputUriPrefix: v.OutputDataURIPrefix,
		},
	}

	// Construct dedicated resources for the job
	dedicatedResources := &v1beta1.GoogleCloudAiplatformV1beta1BatchDedicatedResourcesArgs{
		MachineSpec: &v1beta1.GoogleCloudAiplatformV1beta1MachineSpecArgs{
			MachineType:      v.MachineType,
			AcceleratorCount: v.AcceleratorCount,
			AcceleratorType: v.AcceleratorType.ToStringOutput().ApplyT(func(accelType string) v1beta1.GoogleCloudAiplatformV1beta1MachineSpecAcceleratorType {
				if accelType != "" {
					return v1beta1.GoogleCloudAiplatformV1beta1MachineSpecAcceleratorType(accelType)
				}
				return ""
			}).(v1beta1.GoogleCloudAiplatformV1beta1MachineSpecAcceleratorTypeOutput),
		},
		StartingReplicaCount: v.StartingReplicaCount,
		MaxReplicaCount:      v.MaxReplicaCount,
	}

	batchJobArgs := &v1beta1.BatchPredictionJobArgs{
		Project:            pulumi.String(v.Project),
		Location:           pulumi.String(v.Region),
		DisplayName:        v.JobDisplayName,
		Model:              modelDeployment.DeployedModelId, // Use the deployed model ID
		InputConfig:        inputConfig,
		OutputConfig:       outputConfig,
		DedicatedResources: dedicatedResources,
		ServiceAccount:     serviceAccountEmail,
		ManualBatchTuningParameters: &v1beta1.GoogleCloudAiplatformV1beta1ManualBatchTuningParametersArgs{
			BatchSize: v.BatchSize,
		},
		Labels: pulumi.ToStringMap(v.Labels),
	}

	batchPredictionJob, err := v1beta1.NewBatchPredictionJob(ctx,
		v.NewResourceName("batch-prediction-job", "", 63),
		batchJobArgs,
		pulumi.Parent(v),
		pulumi.DependsOn([]pulumi.Resource{modelDeployment}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch prediction job: %w", err)
	}

	return batchPredictionJob, nil
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

// GetModel returns the Vertex AI Model resource.
// Note: In this version, models are managed through the model deployment component.
func (v *AIBatch) GetModel() interface{} {
	return nil
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
	return v.uploadedModelArtifacts
}
