package gcp

import (
	vertexmodeldeployment "github.com/davidmontoyago/pulumi-gcp-vertex-model-deployment/sdk/go/pulumi-gcp-vertex-model-deployment/resources"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

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
		ServiceAccount:                 serviceAccountEmail,
		// TODO make me configurable
		PredictRoute: pulumi.String("/predict"),
		HealthRoute:  pulumi.String("/health"),
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
