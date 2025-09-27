package gcp

import (
	"fmt"
	"time"

	vertexmodeldeployment "github.com/davidmontoyago/pulumi-gcp-vertex-model-deployment/sdk/go/pulumi-gcp-vertex-model-deployment/resources"
	"github.com/pulumi/pulumi-google-native/sdk/go/google/aiplatform/v1beta1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createBatchPredictionJob creates a Vertex AI Batch Prediction Job.
func (v *AIBatch) createBatchPredictionJob(ctx *pulumi.Context,
	modelDeployment *vertexmodeldeployment.VertexModelDeployment,
	inputDataBucketURI pulumi.StringOutput,
	serviceAccountEmail pulumi.StringOutput) (*v1beta1.BatchPredictionJob, error) {

	var modelName pulumi.StringOutput
	dependencies := []pulumi.Resource{v.artifactsBucket}
	if modelDeployment != nil {
		dependencies = append(dependencies, modelDeployment)
		modelName = modelDeployment.ModelName
	} else {
		// if no model deployment, it's a model from the garden
		modelName = pulumi.String(v.ModelName).ToStringOutput()
	}
	if v.repoIamMember != nil {
		// wait for IAM binding to access a private registry
		dependencies = append(dependencies, v.repoIamMember)
	}

	// Construct the input config
	inputConfig := &v1beta1.GoogleCloudAiplatformV1beta1BatchPredictionJobInputConfigArgs{
		InstancesFormat: v.InputFormat,
		GcsSource: &v1beta1.GoogleCloudAiplatformV1beta1GcsSourceArgs{
			Uris: pulumi.StringArray{
				pulumi.Sprintf("%s/*.jsonl", inputDataBucketURI),
			},
		},
	}

	// Construct the output config
	outputConfig := &v1beta1.GoogleCloudAiplatformV1beta1BatchPredictionJobOutputConfigArgs{
		PredictionsFormat: v.OutputFormat,
		GcsDestination: &v1beta1.GoogleCloudAiplatformV1beta1GcsDestinationArgs{
			OutputUriPrefix: pulumi.Sprintf("gs://%s/%s", v.artifactsBucket.Name, v.OutputDataPath),
		},
	}

	// Construct dedicated resources for the job
	dedicatedResources := &v1beta1.GoogleCloudAiplatformV1beta1BatchDedicatedResourcesArgs{
		MachineSpec: &v1beta1.GoogleCloudAiplatformV1beta1MachineSpecArgs{
			MachineType:      v.MachineType,
			AcceleratorCount: v.AcceleratorCount,
			AcceleratorType: v.AcceleratorType.ApplyT(func(accelType string) v1beta1.GoogleCloudAiplatformV1beta1MachineSpecAcceleratorType {
				return v1beta1.GoogleCloudAiplatformV1beta1MachineSpecAcceleratorType(accelType)
			}).(v1beta1.GoogleCloudAiplatformV1beta1MachineSpecAcceleratorTypeOutput),
		},
		StartingReplicaCount: v.StartingReplicaCount,
		MaxReplicaCount:      v.MaxReplicaCount,
	}

	batchJobArgs := &v1beta1.BatchPredictionJobArgs{
		Project:            pulumi.String(v.Project),
		Location:           pulumi.String(v.Region),
		DisplayName:        v.JobDisplayName,
		Model:              modelName, // Use the deployed model name or the name of a model from the garden
		InputConfig:        inputConfig,
		OutputConfig:       outputConfig,
		DedicatedResources: dedicatedResources,
		ServiceAccount:     serviceAccountEmail,
		ManualBatchTuningParameters: &v1beta1.GoogleCloudAiplatformV1beta1ManualBatchTuningParametersArgs{
			BatchSize: v.BatchSize,
		},
		Labels: pulumi.ToStringMap(v.Labels),
	}

	// every pulumi up operation is a new launch
	jobName := fmt.Sprintf("%s-%d", v.NewResourceName("batch-prediction-job", "", 63), time.Now().UnixMilli())

	batchPredictionJob, err := v1beta1.NewBatchPredictionJob(ctx,
		jobName,
		batchJobArgs,
		pulumi.Parent(v),
		pulumi.DependsOn(dependencies),
		pulumi.RetainOnDelete(v.retainJobOnDelete),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch prediction job: %w", err)
	}

	return batchPredictionJob, nil
}
