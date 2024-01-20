include .env

run:
	@air

tailwind-build:
	@~/tailwindcss -i ./templates/static/css/main.css -o ./templates/static/css/tailwind.css

tailwind-watch:
	@~/tailwindcss -i ./templates/static/css/main.css -o ./templates/static/css/tailwind.css --watch

docker-up:
	@docker compose -f .docker/docker-compose.yml up -d

migrate-up:
	@migrate -database $(POSTGRESQL_URL) -path db/migrations up

migrate-down:
	@migrate -database $(POSTGRESQL_URL) -path db/migrations down
