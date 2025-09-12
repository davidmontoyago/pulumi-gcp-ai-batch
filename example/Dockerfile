# Start from pre-built container
FROM us-docker.pkg.dev/vertex-ai/prediction/pytorch-cpu.1-12:latest

# Install additional dependencies for your predictor
COPY requirements.txt .
RUN pip install -r requirements.txt

# Copy your custom predictor code
COPY predictor.py /opt/prediction/
COPY model_artifacts/ /opt/model/

# Set environment variables for CPR
ENV AIP_PREDICTOR_CLASS="predictor.BertSentimentPredictor"
ENV AIP_HTTP_PORT=8080

# Use CPR's HTTP server
ENTRYPOINT ["python", "-m", "google.cloud.aiplatform.prediction.model_server"]
