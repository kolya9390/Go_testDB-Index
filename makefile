build: 
	GOOS=linux GOARCH=amd64 go build -o service_linux cmd/app/main.go

docker-build:
	docker build .

run:
	docker compose -f "docker-compose.yml" up -d --build