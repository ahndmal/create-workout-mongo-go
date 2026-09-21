deploy:
	gcloud run deploy create-workout-mongo-go \
       --source . \
       --function functions.HelloWorld \
       --base-image java25 \
       --region REGION \
       --allow-unauthenticated
	gcloud functions deploy create-workout-mongo-go \
		 --gen2 \
		 --region=us-central1 \
		 --memory=128Mi \
		 --source=. \
		 --runtime=go121 \
		 --entry-point=createWorkout \
		 --allow-unauthenticated
update:
	go get -u ./...
