buildAPP:
	docker build -t app ./app
runApp:
	docker run -it --rm -p 3000:3000 --volume appData:/etc/todos app:latest
.PHONY: buildAPP runApp