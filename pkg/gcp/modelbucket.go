package gcp

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// uploadDirectoryToBucket traverses a directory and uploads all files to a GCS bucket.
func (v *AIBatch) uploadDirectoryToBucket(ctx *pulumi.Context, localDir, baseObjectPath string) ([]pulumi.Resource, error) {
	if localDir == "" {
		// no model artifacts to upload. skip
		return []pulumi.Resource{}, nil
	}

	var bucketObjects []*storage.BucketObject

	err := filepath.Walk(localDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking path %s: %w", filePath, err)
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip hidden files and system files
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// Calculate relative path from the base directory to preserve directory structure
		relPath, err := filepath.Rel(localDir, filePath)
		if err != nil {
			return fmt.Errorf("error calculating relative path: %w", err)
		}

		// Convert to GCS object key (this preserves the original filename and path structure)
		gcsObjectName := strings.ReplaceAll(relPath, string(filepath.Separator), "/")

		// Detect content type
		contentType := detectContentType(filePath)

		// Create a unique resource name by replacing path separators with hyphens
		resourceName := fmt.Sprintf("file-%s", strings.ReplaceAll(gcsObjectName, "/", "-"))
		resourceName = strings.ReplaceAll(resourceName, ".", "-")

		// Prepend the base object path if provided
		if baseObjectPath != "" {
			gcsObjectName = filepath.Join(baseObjectPath, gcsObjectName)
			gcsObjectName = strings.ReplaceAll(gcsObjectName, string(filepath.Separator), "/")
		}

		// Create BucketObject resource
		bucketObject, err := storage.NewBucketObject(ctx, resourceName, &storage.BucketObjectArgs{
			Name:        pulumi.String(gcsObjectName),
			Bucket:      v.artifactsBucket.Name,
			Source:      pulumi.NewFileAsset(filePath),
			ContentType: pulumi.String(contentType),
		}, pulumi.Parent(v))
		if err != nil {
			return fmt.Errorf("error creating bucket object for %s: %w", filePath, err)
		}

		bucketObjects = append(bucketObjects, bucketObject)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error uploading directory %s: %w", localDir, err)
	}

	uploadedResources := make([]pulumi.Resource, len(bucketObjects))
	for i, bucketObject := range bucketObjects {
		uploadedResources[i] = bucketObject
	}

	return uploadedResources, nil
}

// detectContentType determines the MIME type of a file based on its extension
func detectContentType(filePath string) string {
	ext := filepath.Ext(filePath)
	contentType := mime.TypeByExtension(ext)

	if contentType == "" {
		// Default to binary if type cannot be determined
		contentType = "application/octet-stream"
	}

	return contentType
}

// setupModelBucket creates a bucket for model artifacts and uploads the model directory if any.
// It returns the GCS URI of the uploaded model artifacts and the uploaded objects for dependency tracking.
func (v *AIBatch) setupModelBucket(ctx *pulumi.Context, modelDir string, modelBucketBasePath string, labels map[string]string) (pulumi.StringOutput, []pulumi.Resource, error) {
	// Create the bucket for model artifacts
	bucketName := v.NewResourceName("vertex-model", "bucket", 63)

	// Merge default labels with provided labels
	bucketLabels := pulumi.StringMap{
		"purpose": pulumi.String("model-storage"),
	}

	// Add user-provided labels
	for key, value := range labels {
		bucketLabels[key] = pulumi.String(value)
	}

	artifactsBucket, err := storage.NewBucket(ctx, bucketName, &storage.BucketArgs{
		Name:         pulumi.String(bucketName),
		Location:     pulumi.String(v.Region),
		Project:      pulumi.String(v.Project),
		ForceDestroy: pulumi.Bool(true), // Model data is part of the pipeline, safe to implode.
		// Enable Uniform Bucket Level Access (UBLA) for enhanced security
		// This is required for SBOMs and prevents ACL-based access control
		UniformBucketLevelAccess: pulumi.Bool(true),
		Versioning: &storage.BucketVersioningArgs{
			Enabled: pulumi.Bool(true), // Enable versioning for audit trail
		},
		Labels: bucketLabels,
	}, pulumi.Parent(v))
	if err != nil {
		return pulumi.StringOutput{}, nil, fmt.Errorf("failed to create artifacts bucket: %w", err)
	}
	v.artifactsBucket = artifactsBucket

	// No luck with https://github.com/pulumi/pulumi-synced-folder /o\

	// Upload the model artifacts, if any
	uploadedObjects, err := v.uploadDirectoryToBucket(ctx, modelDir, modelBucketBasePath)
	if err != nil {
		return pulumi.StringOutput{}, nil, fmt.Errorf("failed to upload model artifacts: %w", err)
	}

	modelArtifactsURI := pulumi.Sprintf("gs://%s/%s", artifactsBucket.Name, modelBucketBasePath)

	return modelArtifactsURI, uploadedObjects, nil
}

// uploadInputDataToBucket uploads the input data to the bucket.
func (v *AIBatch) uploadInputDataToBucket(ctx *pulumi.Context, inputDataDir string, inputDataBasePath string) (pulumi.StringOutput, []pulumi.Resource, error) {
	uploadedDataObjects, err := v.uploadDirectoryToBucket(ctx, inputDataDir, inputDataBasePath)
	if err != nil {
		return pulumi.StringOutput{}, nil, fmt.Errorf("failed to upload input data to bucket: %w", err)
	}

	inputDataBucketURI := pulumi.Sprintf("gs://%s/%s", v.artifactsBucket.Name, inputDataBasePath)

	return inputDataBucketURI, uploadedDataObjects, nil
}

// collectBucketObjectNames collects the names of the uploaded model artifacts and data objects.
func collectBucketObjectNames(
	uploadedModelArtifacts []pulumi.Resource,
	uploadedDataObjects []pulumi.Resource,
) pulumi.StringArrayOutput {
	uploadedObjectNames := pulumi.StringArray{}

	for _, resource := range uploadedModelArtifacts {
		if bucketObject, ok := resource.(*storage.BucketObject); ok {
			uploadedObjectNames = append(uploadedObjectNames, bucketObject.Name.ApplyT(func(name string) string {
				return name
			}).(pulumi.StringOutput))
		}
	}

	for _, resource := range uploadedDataObjects {
		if bucketObject, ok := resource.(*storage.BucketObject); ok {
			uploadedObjectNames = append(uploadedObjectNames, bucketObject.Name.ApplyT(func(name string) string {
				return name
			}).(pulumi.StringOutput))
		}
	}

	return uploadedObjectNames.ToStringArrayOutput()
}
