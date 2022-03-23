IMAGE = ohmnyom
REPOSITORY = asia-northeast3-docker.pkg.dev/ohmnyom/server

build :
	docker build -t $(REPOSITORY)/$(IMAGE):latest .

push :
	docker push $(REPOSITORY)/$(IMAGE):latest

deploy :
	gcloud config set project ohmnyom
	gcloud config set run/region asia-northeast3
	gcloud run services replace deployment/service.yaml
