# Sentiment Analysis with llama 3.2

Llama model from the garden deployed as a batch job. It skips model upload to the model registry.

See:
- https://cloud.google.com/vertex-ai/docs/predictions/get-batch-model-garden
- https://cloud.google.com/vertex-ai/generative-ai/docs/partner-models/llama/llama-batch#cloudstorage

## Deploy

```sh
cd ./example/llama-sentiment-analysis

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
          "content": "**Summary of the Discussion:**\n\nA customer expressed extreme frustration and anger towards a company due to a billing issue that had been unresolved for three weeks. The customer felt ignored and unvalued by the company's customer service representatives, who provided inconsistent information. The customer demanded that the representative directly address the issue without being put on hold or transferred to another person.\n\n**Breakdown of Each Comment with Sentiment Analysis:**\n\n1. Customer: \"I am absolutely furious right now!\"\n   - Sentiment Analysis: **Hot (100)** - The customer uses strong language to express their extreme frustration and anger.\n\n2. Customer: \"I've been trying to resolve this billing issue for THREE WEEKS and nobody at your company seems to care or know what they're doing.\"\n   - Sentiment Analysis: **Hot (100)** - The customer expresses a deep sense of frustration and betrayal towards the company's lack of attention to their issue.\n\n3. Customer: \"I've been charged twice for the same service, my account is completely messed up, and every single person I talk to gives me different information.\"\n   - Sentiment Analysis: **Hot (100)** - The customer feels disrespected and helpless due to the inconsistent and incorrect information provided by the representatives.\n\n4. Customer: \"This is the worst customer service I've ever experienced in my life!\"\n   - Sentiment Analysis: **Hot (100)** - The customer explicitly states their dissatisfaction and disappointment with the company's customer service.\n\n5. Representative: \"I sincerely apologize for the frustration you've experienced, and I understand how upsetting this situation must be.\"\n   - Sentiment Analysis: **Cold (20)** - The representative shows empathy but does not acknowledge the depth of the customer's frustration.\n\n6. Customer: \"Account number 4478-9921-3.\"\n   - Sentiment Analysis: **Neutral (0)** - The customer simply provides the requested information.\n\n7. Customer: \"And don't you dare put me on hold or transfer me to someone else! I've already explained this story four times to different people who clearly don't communicate with each other.\"\n   - Sentiment Analysis: **Hot (100)** - The customer becomes aggressive and demanding, feeling that they are being ignored and disrespected.\n\n8. Customer: \"Your company is a complete joke - I'm paying for services I'm not even receiving properly, and now you're stealing money from me with duplicate charges.\"\n   - Sentiment Analysis: **Hot (100)** - The customer uses strong language to express their outrage and betrayal, feeling that the company is taking advantage of them.\n\nThe conversation is marked with extreme frustration, anger, and a sense of being disrespected and ignored by the company's customer service representatives.",
          "role": "assistant",
          "tool_calls": []
        },
        "stop_reason": null
      }
    ],
    "created": 1759188343,
    "id": "chatcmpl-4eb5c872319144149cada060bdee5be9",
    "model": "meta-llama/Llama-3.2-3B-Instruct",
    "object": "chat.completion",
    "prompt_logprobs": null,
    "usage": {
      "completion_tokens": 558,
      "prompt_tokens": 324,
      "prompt_tokens_details": null,
      "total_tokens": 882
    }
  },
  "id": "ca6661dd-5c29-4b77-a03d-e6196ccb4d19"
}
```
