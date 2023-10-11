deploy:
	gcloud functions deploy create-workout-mongo-go \
		 --gen2 \
		 --region=us-central1 \
		 --memory=128Mi \
		 --source=. \
		 --runtime=go121 \
		 --entry-point=createWorkout \
		 --allow-unauthenticated
