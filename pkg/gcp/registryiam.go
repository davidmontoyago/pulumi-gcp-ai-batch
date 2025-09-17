package gcp

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/artifactregistry"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// grantRegistryIAMAccess grants the SA access to the registry source of the model docker image.
func (v *AIBatch) grantRegistryIAMAccess(ctx *pulumi.Context, serviceAccountEmail pulumi.StringOutput) (*artifactregistry.RepositoryIamMember, error) {
	modelImageRepoName := v.ModelImageURL.ApplyT(func(url string) string {
		return strings.Split(url, "/")[2]
	}).(pulumi.StringOutput)

	bindingName := v.NewResourceName("model-registry-access", "iam-member", 63)
	repoMember, err := artifactregistry.NewRepositoryIamMember(ctx, bindingName, &artifactregistry.RepositoryIamMemberArgs{
		Repository: modelImageRepoName,
		Location:   pulumi.String(v.Region),
		Project:    pulumi.String(v.Project),
		Role:       pulumi.String("roles/artifactregistry.reader"),
		Member:     pulumi.Sprintf("serviceAccount:%s", serviceAccountEmail),
	}, pulumi.Parent(v))
	if err != nil {
		return nil, fmt.Errorf("failed to grant registry IAM access: %w", err)
	}

	return repoMember, nil
}
