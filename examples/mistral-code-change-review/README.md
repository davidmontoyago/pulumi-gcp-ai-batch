# Code Change Review with Mistral 7b Instruct v0.2

Mistral model from the garden deployed as a batch job to perform a code review on a git patch.

See:
- https://cloud.google.com/vertex-ai/docs/predictions/get-batch-model-garden

## Deploy

```sh
cd ./example/mistral-code-change-review

# set env vars in ./config/env.example

# init pulumi state and stack
make pulumi-init

# proceed to deploy the AIBatch with Pulumi
make deploy

# see inference results in GCS

# clean up
make undeploy
```

Example prediction output:

```json
{
  "custom_id": "test-request-0",
  "error": "",
  "response": {
    "choices": [
      {
        "finish_reason": "stop",
        "index": 0,
        "logprobs": null,
        "message": {
          "content": " The provided pull request introduces two changes: the addition of a `.gitignore` file modification and the creation of a new `go.work` file. These changes aim to set up a Go workspace for the project.\n\nSynthesis of the changes:\n1. The `.gitignore` file was modified, removing the exclusion of the `go.work` file and keeping the existing exclusions. This change enables the Go workspace file to be included in the Git repository.\n2. A new `go.work` file was created, which specifies the Go version and the directories to be used in the workspace.\n\nApproach taken:\nThe developer, David Montoya, has set up a Go workspace for the project by adding the necessary files and updates to the `.gitignore` file. This allows the team to manage their dependencies more effectively and maintain a consistent environment.\n\nTest cases:\nSince the changes are primarily infrastructure-related and do not introduce new functionality, test cases are not strictly necessary. However, it is recommended to verify the correct functioning of the Go workspace setup by testing the builds and the behavior of the application with the updated dependencies.\n\nLevel of complexity:\nThe level of complexity for these changes is estimated to be 1, as the changes are straightforward and focused on setting up the Go workspace. The developer has correctly followed best practices for managing Go dependencies and has effectively updated the `.gitignore` file.",
          "role": "assistant",
          "tool_calls": []
        },
        "stop_reason": null
      }
    ],
    "created": 1759285885,
    "id": "cmpl-bdc55950843740afb961d8ca9f4e6417",
    "model": "mistralai/Mistral-7B-Instruct-v0.2",
    "object": "chat.completion",
    "usage": {
      "completion_tokens": 306,
      "prompt_tokens": 317,
      "total_tokens": 623
    }
  },
  "id": "11e3b2b9-61db-4d2d-8a54-1b45c6c3df9d"
}

```
