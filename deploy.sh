#!/bin/bash

# Build Docker image
docker build -t us-central1-docker.pkg.dev/aicoexist-446217/aisense-repo/aisense_portal_snapshot .

# Push Docker image
docker push us-central1-docker.pkg.dev/aicoexist-446217/aisense-repo/aisense_portal_snapshot

# Deploy to Cloud Run
gcloud run deploy aisense-app \
  --image=us-central1-docker.pkg.dev/aicoexist-446217/aisense-repo/aisense_portal_snapshot \
  --region=us-central1 \
  --platform=managed \
  --vpc-connector=aisense-vpc-connector \
  --vpc-egress=all-traffic \
  --env-vars-file .env.yaml \
  --port=8080 \
  --allow-unauthenticated \
  --timeout=500s
