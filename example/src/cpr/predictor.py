import os
import pickle
import torch
from transformers import BertTokenizer, BertForSequenceClassification
from google.cloud.aiplatform.prediction.predictor import Predictor
from google.cloud.aiplatform.utils import prediction_utils

class BertSentimentPredictor(Predictor):
    def __init__(self):
        return

    def load(self, artifacts_uri: str) -> None:
        """Load BERT model and tokenizer."""
        prediction_utils.download_model_artifacts(artifacts_uri)

        # Load tokenizer and model
        self._tokenizer = BertTokenizer.from_pretrained('bert-base-uncased')
        self._model = BertForSequenceClassification.from_pretrained(
            artifacts_uri,
            num_labels=3  # positive, negative, neutral
        )
        self._model.eval()

    def preprocess(self, prediction_input):
        """Handle BERT-specific preprocessing."""
        instances = prediction_input.instances

        preprocessed_instances = []
        for instance in instances:
            text = instance["text"]

            # BERT tokenization with required encoding
            encoded = self._tokenizer(
                text,
                truncation=True,
                padding='max_length',
                max_length=512,
                return_tensors='pt'
            )

            preprocessed_instances.append({
                'input_ids': encoded['input_ids'].squeeze().tolist(),
                'attention_mask': encoded['attention_mask'].squeeze().tolist(),
                'token_type_ids': encoded.get('token_type_ids',
                    torch.zeros_like(encoded['input_ids'])).squeeze().tolist()
            })

        return prediction_utils.PredictionInput(
            instances=preprocessed_instances
        )

    def predict(self, instances):
        """Run BERT inference."""
        predictions = []

        for instance in instances:
            input_ids = torch.tensor([instance['input_ids']])
            attention_mask = torch.tensor([instance['attention_mask']])
            token_type_ids = torch.tensor([instance['token_type_ids']])

            with torch.no_grad():
                outputs = self._model(
                    input_ids=input_ids,
                    attention_mask=attention_mask,
                    token_type_ids=token_type_ids
                )

            probabilities = torch.softmax(outputs.logits, dim=-1)
            predicted_class = torch.argmax(probabilities, dim=-1).item()
            confidence = probabilities.max().item()

            sentiment_labels = ["negative", "neutral", "positive"]
            predictions.append({
                "sentiment": sentiment_labels[predicted_class],
                "confidence": confidence,
                "probabilities": probabilities.squeeze().tolist()
            })

        return predictions
