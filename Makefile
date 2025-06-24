

run:
	docker compose -f docker-compose.yml up --build --watch

stop:
	docker compose -f docker-compose.yml down
