build:
	docker build -t extrator:v1.0 .
	docker pull mysql:8.0

run: build
	mkdir -p application db
	docker-compose up -d

clean:
	docker-compose down