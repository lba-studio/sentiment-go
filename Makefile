registry_uri = public.ecr.aws/t5m8k1a3
image_name = sentiment-go
repository = ${image_name}
repository_uri = ${registry_uri}/${repository}

build:
	go build -o sentiment-go main.go

docker_login:
	aws ecr-public get-login-password --region us-east-1 | docker login --username AWS --password-stdin $(registry_uri)

docker_build:
	docker build -t ${image_name} .

publish: docker_login docker_build
	docker tag ${image_name}:latest ${repository_uri}:latest
	docker push ${repository_uri}