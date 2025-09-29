package gcp

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// grantModelIAMRoles grants necessary IAM roles to the model service account.
func (v *AIBatch) grantModelIAMRoles(ctx *pulumi.Context, serviceAccountEmail pulumi.StringOutput) ([]*projects.IAMMember, error) {
	// IAM roles specific to what the batch prediction job needs to operate
	roles := []string{
		"roles/storage.bucketViewer",    // List and get buckets
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

// createModelServiceAccount creates a service account for Vertex AI operations.
func (v *AIBatch) createModelServiceAccount(ctx *pulumi.Context) (pulumi.StringOutput, error) {
	accountID := v.NewResourceName("model-account", "", 30)

	modelServiceAccount, err := serviceaccount.NewAccount(ctx, v.NewResourceName("model-account", "", 63), &serviceaccount.AccountArgs{
		Project:     pulumi.String(v.Project),
		AccountId:   pulumi.String(accountID),
		DisplayName: pulumi.Sprintf("%s Vertex AI Service Account", v.ModelDisplayName),
		Description: pulumi.String("Service account for deployed model operations"),
	}, pulumi.Parent(v))
	if err != nil {
		return pulumi.StringOutput{}, fmt.Errorf("failed to create model service account: %w", err)
	}

	return modelServiceAccount.Email, nil
}
