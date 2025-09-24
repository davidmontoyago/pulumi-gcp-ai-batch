package gcp

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Helper functions for setting default values
func setDefaultString(input pulumi.StringInput, defaultValue string) pulumi.StringOutput {
	if input == nil {
		return pulumi.String(defaultValue).ToStringOutput()
	}

	return input.ToStringOutput()
}

func setDefaultInt(input pulumi.IntInput, defaultValue int) pulumi.IntOutput {
	if input == nil {
		return pulumi.Int(defaultValue).ToIntOutput()
	}

	return input.ToIntOutput()
}
