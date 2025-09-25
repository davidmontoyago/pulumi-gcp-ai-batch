# Pre-built container to run on CPUs, not GPUs
FROM us-docker.pkg.dev/vertex-ai/prediction/pytorch-cpu.1-12:latest AS builder

USER root

RUN chown -R model-server /home/model-server/

USER model-server

ENV PATH="/home/model-server/.local/bin:${PATH}"

# Install additional dependencies for your predictor
COPY src/cpr/requirements.txt .
RUN pip install --user --no-cache-dir -r requirements.txt

# Runtime stage
FROM us-docker.pkg.dev/vertex-ai/prediction/pytorch-cpu.1-12:latest

# Copy only the installed packages from builder
COPY --from=builder /home/model-server/.local /home/model-server/.local

COPY src/cpr/predictor.py /home/model-server/cpr/
COPY models/hasnain43-bert-stock-sentiment-v1/ /home/model-server/model/

# Environment variables

# The path to the directory containing the Model artifact and any of its
# supporting files. The path is either a GCS uri or the path to a local directory.
# If this parameter is set to a GCS uri:
# (1) `credential_path` must be specified for local prediction.
# (2) The GCS uri will be passed directly to `Predictor.load`.
# If this parameter is a local directory:
# (1) The directory will be mounted to a default temporary model path.
# (2) The mounted path will be passed to `Predictor.load`.
ENV AIP_STORAGE_URI="gs://example-prediction-batch-vertex-model-bucket/model-hasnain43"

# --- config for model server and custom prediction routine ---

# handler is the python program that captures the requests and invokes the predictor
# this is the default handler
# See:
# https://github.com/googleapis/python-aiplatform/blob/v1.90.0/google/cloud/aiplatform/prediction/handler.py
ENV HANDLER_MODULE="google.cloud.aiplatform.prediction.handler"
ENV HANDLER_CLASS="PredictionHandler"

# predictor is the custom prediction routine invoked by the handler to
# pre-process, predict and post-process
ENV PREDICTOR_MODULE="cpr.predictor"
ENV PREDICTOR_CLASS="BertSentimentPredictor"

ENV AIP_HTTP_PORT=8080
ENV AIP_PREDICT_ROUTE=/predict
ENV AIP_HEALTH_ROUTE=/health

# required by model
ENV HF_HOME="/tmp/huggingface"

ENTRYPOINT ["python", "-m", "google.cloud.aiplatform.prediction.model_server"]
