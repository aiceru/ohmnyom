IMAGE = ohmnyom
REPOSITORY = asia-northeast3-docker.pkg.dev/ohmnyom/server
SERVICE = ohmnyom-grpc-server

build :
	docker build -t $(REPOSITORY)/$(IMAGE):latest .

push :
	docker push $(REPOSITORY)/$(IMAGE):latest

# create service at first time
create :
	gcloud config set project ohmnyom
	gcloud config set run/region asia-northeast3
	gcloud run services deploy deployment/service.yaml

deploy :
	gcloud config set project ohmnyom
	gcloud config set run/region asia-northeast3
	gcloud run deploy $(SERVICE) --image=$(REPOSITORY)/$(IMAGE):latest && \
	gcloud run services update-traffic $(SERVICE) --to-latest
