import torch
from google.cloud.aiplatform.prediction import Predictor
from transformers import AutoTokenizer, AutoModelForSequenceClassification

class BertSentimentPredictor(Predictor):
    def __init__(self):
        return

    def load(self, artifacts_uri: str) -> None:
        print(f"model artifacts bucket: {artifacts_uri}")

        self._tokenizer = AutoTokenizer.from_pretrained("hasnain43/bert-stock-sentiment-v1")
        self._model = AutoModelForSequenceClassification.from_pretrained("hasnain43/bert-stock-sentiment-v1")
        self._model.eval()
        self._loaded = True

    def preprocess(self, prediction_input):
        print("preprocessing...")

        instances = prediction_input.get("instances", [])
        texts = [instance["text"] for instance in instances]

        encoded = self._tokenizer(
            texts,
            truncation=True,
            padding=True,
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

        probabilities = outputs.logits
        predicted_classes = torch.argmax(probabilities, dim=1)
        # confidences = probabilities.max(dim=-1).values

        # prediction results that will be passed to postprocess
        return [
            {
                "predicted_class": pred.item(),
                # "probabilities": prob.tolist(),
                # "confidence": conf.item(),
            }
            for prob, pred in zip(probabilities, predicted_classes)
        ]

    def postprocess(self, predictions):
        print("postprocessing...")

        sentiment_labels = {0: "negative", 1: "neutral", 2: "positive"}

        processed_predictions = []
        for prediction in predictions:
            predicted_class = prediction["predicted_class"]
            # probabilities = prediction["probabilities"]
            # confidence = prediction["confidence"]

            processed_predictions.append(
                {
                    "sentiment": sentiment_labels[predicted_class],
                    # "confidence": confidence,
                    # "probabilities": probabilities,
                }
            )

        return {"predictions": processed_predictions}
