// Package main provides a basic example using AIBatch component.
package main

import (
	"github.com/davidmontoyago/pulumi-gcp-ai-batch/pkg/gcp"
	"github.com/davidmontoyago/pulumi-gcp-ai-batch/pkg/gcp/config"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		args := cfg.ToAIBatchArgs()

		predictionBatch, err := gcp.NewAIBatch(ctx, "mistral-prediction-batch", args)
		if err != nil {
			return err
		}

		ctx.Export("batchPredictionJobName", predictionBatch.GetBatchPredictionJob().Name)
		ctx.Export("batchPredictionJobDisplayName", predictionBatch.GetBatchPredictionJob().DisplayName)
		ctx.Export("batchPredictionJobModelVersionId", predictionBatch.GetBatchPredictionJob().ModelVersionId)
		ctx.Export("modelServiceAccountEmail", predictionBatch.GetModelServiceAccountEmail())
		if predictionBatch.GetModelDeployment() != nil {
			ctx.Export("modelArtifactsBucketUri", predictionBatch.GetModelDeployment().ModelArtifactsBucketUri)
			ctx.Export("modellOutputsBucketUri", pulumi.Sprintf("gs://%s/%s",
				predictionBatch.GetModelDeployment().ModelArtifactsBucketUri,
				predictionBatch.OutputDataPath,
			))
			ctx.Export("modellInputsBucketUri", pulumi.Sprintf("gs://%s/%s",
				predictionBatch.GetModelDeployment().ModelArtifactsBucketUri,
				predictionBatch.InputDataPath,
			))
		}

		return nil
	})
}
