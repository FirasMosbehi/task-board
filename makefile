.PHONY: rebuild up

rebuild-local:
	docker compose -f docker-compose.local.yml up -d --force-recreate --build

up-local:
	docker compose -f docker-compose.local.yml up -d --force-recreate

rebuild:
	docker compose build --no-cache

up:
	docker compose up -d	