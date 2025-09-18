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

# Copy application code and models
COPY src/cpr/predictor.py /opt/prediction/
# COPY models/bert-tensorflow2-bert-en-uncased-l-10-h-128-a-2-v2/ /opt/model/

# Environment variables
ENV PREDICTOR_CLASS="predictor.BertSentimentPredictor"
ENV AIP_HTTP_PORT=8080
ENV AIP_PREDICT_ROUTE=/predict
ENV AIP_HEALTH_ROUTE=/health

ENTRYPOINT ["python", "-m", "google.cloud.aiplatform.prediction.model_server"]
