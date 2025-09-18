# BERT model with Custom Prediction Routine (CPR)

BERT deployment with custom docker image to use Custom Prediction Routines for model input preprocessing and output translation.

The custom prediction routine is a BERT predictor for sentiment analysis. It's packaged in a custom container and deployed with the GCP-provided [CprModelServer](https://github.com/googleapis/python-aiplatform/blob/18a55590c5679b8ea7536c4c3c73566ba006bf36/google/cloud/aiplatform/prediction/model_server.py#L48).

See:
- https://cloud.google.com/blog/topics/developers-practitioners/simplify-model-serving-custom-prediction-routines-vertex-ai

## Deploy

```sh
cd ./example

# set env vars in ./config/env.example

# configure the registry for the custom model server
export ARTIFACT_REGISTRY=<>

# build & push the image
make image

make push-image

# init pulumi state and stack
make pulumi-init

# proceed to deploy the AIBatch with Pulumi
make deploy
```
