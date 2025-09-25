---
license: mit
tags:
  - sentiment-analysis
  - finance
  - bert
  - stock-news
  - classification
---

# BERT Stock Sentiment Classifier (Fine-Tuned)

This is a BERT-based model fine-tuned on a dataset of stock market news headlines to perform **sentiment analysis**. The labels are:

- `positive`
- `neutral`
- `negative`

The model is intended for **financial news and headlines**, especially useful for trading, sentiment scoring, or market analysis pipelines.

## 🧾 How It Was Trained

- **Base model**: `bert-base-uncased`
- **Dataset**: Custom scraped Finviz news headlines
- **Labels**: Generated using FinBERT, mapped to `positive`, `neutral`, `negative`
- **Training**: 3 epochs, batch size 16, learning rate 2e-5

## 🛠 Usage

```python
from transformers import AutoTokenizer, AutoModelForSequenceClassification
import torch

tokenizer = AutoTokenizer.from_pretrained("hasnain43/bert-stock-sentiment-v1")
model = AutoModelForSequenceClassification.from_pretrained("hasnain43/bert-stock-sentiment-v1")
model.eval()

label_map = {0: "negative", 1: "neutral", 2: "positive"}

def predict_sentiment(text):
    inputs = tokenizer(text, return_tensors="pt", truncation=True, padding=True)
    with torch.no_grad():
        outputs = model(**inputs)
        logits = outputs.logits
        prediction = torch.argmax(logits, dim=1).item()
    return label_map[prediction]

predict_sentiment("Tesla stock drops after disappointing delivery numbers.")
