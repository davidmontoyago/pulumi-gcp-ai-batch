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

		predictionBatch, err := gcp.NewAIBatch(ctx, "example-prediction-batch", args)
		if err != nil {
			return err
		}

		ctx.Export("batchPredictionJobName", predictionBatch.GetBatchPredictionJob().Name)
		ctx.Export("batchPredictionJobModelVersionId", predictionBatch.GetBatchPredictionJob().ModelVersionId)
		ctx.Export("modelServiceAccountEmail", predictionBatch.GetModelServiceAccount().Email)

		return nil
	})
}
