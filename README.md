`gcloud beta functions deploy CloudBuildNotifier --set-env-vars USERNAME=$BITBUCKET_API_USER,PASSWORD=$API_USER_PASSWORD --runtime go111 --trigger-topic cloud-builds`
