import json
import torch
import numpy as np
from typing import Type, Optional, Any, Dict
from fastapi import HTTPException
from google.cloud.aiplatform.prediction import PredictionHandler, Predictor
from transformers import BertTokenizer, BertForSequenceClassification


class BertSentimentPredictor(Predictor):
    def __init__(self):
        return

    def load(self, artifacts_uri: str) -> None:
        print(f"model artifacts bucket: {artifacts_uri}")

        self._tokenizer = BertTokenizer.from_pretrained("bert-base-uncased")
        self._model = BertForSequenceClassification.from_pretrained(
            "bert-base-uncased", num_labels=3  # positive, negative, neutral
        )
        self._model.eval()
        self._loaded = True

    def preprocess(self, prediction_input):
        print("preprocessing...")

        instances = prediction_input.get("instances", [])
        texts = [instance["text"] for instance in instances]

        encoded = self._tokenizer(
            texts,
            truncation=True,
            padding="max_length",
            max_length=512,
            return_tensors="pt",
        )
        return {
            "input_ids": encoded["input_ids"],
            "attention_mask": encoded["attention_mask"],
            "token_type_ids": encoded.get(
                "token_type_ids", torch.zeros_like(encoded["input_ids"])
            ),
        }

    def predict(self, instances):
        # Return dummy predictions for testing
        print("predicting...")

        with torch.no_grad():
            outputs = self._model(
                input_ids=instances["input_ids"],
                attention_mask=instances["attention_mask"],
                token_type_ids=instances["token_type_ids"],
            )

        probabilities = torch.softmax(outputs.logits, dim=-1)
        predicted_classes = torch.argmax(probabilities, dim=-1)
        confidences = probabilities.max(dim=-1).values

        # prediction results that will be passed to postprocess
        return [
            {
                "probabilities": prob.tolist(),
                "predicted_class": pred.item(),
                "confidence": conf.item(),
            }
            for prob, pred, conf in zip(probabilities, predicted_classes, confidences)
        ]

    def postprocess(self, predictions):
        print("postprocessing...")

        sentiment_labels = ["negative", "neutral", "positive"]

        processed_predictions = []
        for prediction in predictions:
            probabilities = prediction["probabilities"]
            predicted_class = prediction["predicted_class"]
            confidence = prediction["confidence"]

            processed_predictions.append(
                {
                    "sentiment": sentiment_labels[predicted_class],
                    "confidence": confidence,
                    "probabilities": probabilities,
                }
            )

        return {"predictions": processed_predictions}
