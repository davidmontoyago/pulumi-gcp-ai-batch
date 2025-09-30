# BERT model with Custom Prediction Routine (CPR)

BERT deployment with custom docker image to use Custom Prediction Routines for model input preprocessing and output translation.

BERT model is https://huggingface.co/nlptown/bert-base-multilingual-uncased-sentiment which predicts products review stars from 1 to 5.

The custom prediction routine uses a BERT sequence classification transformer with `torch.argmax` to get the highest probability of the review falling within one of the 5 review classes. It's packaged in a custom container and deployed with the GCP-provided [CprModelServer](https://github.com/googleapis/python-aiplatform/blob/18a55590c5679b8ea7536c4c3c73566ba006bf36/google/cloud/aiplatform/prediction/model_server.py#L48).

See:
- https://github.com/googleapis/python-aiplatform/tree/v1.116.0/google/cloud/aiplatform/prediction
- https://cloud.google.com/blog/topics/developers-practitioners/simplify-model-serving-custom-prediction-routines-vertex-ai
- https://github.com/googleapis/python-aiplatform/blob/18a55590c5679b8ea7536c4c3c73566ba006bf36/google/cloud/aiplatform/prediction/model_server.py#L48
- https://www.geeksforgeeks.org/nlp/sentiment-classification-using-bert/

## Deploy

```sh
cd ./example/bert-sentiment-analysis-with-cpr

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

# see inference results in GCS

# clean up
make undeploy
```
