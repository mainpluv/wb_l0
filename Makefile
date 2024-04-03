.SILENT:

run:
	go run src/main.go

build:
	docker-compose up -d --build