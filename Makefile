buildAPP:
	docker build -t app ./app
publisAPP:
	docker tag app localhost:4566/app
	docker push localhost:4566/app
runApp:
	docker run -it --rm -p 3000:3000 --volume appData:/etc/todos app:latest
runAppEcr:
	docker run -d --network host --volume appData:/etc/todos localhost:4566/app
.PHONY: buildAPP runApp