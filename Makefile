compose-up: ### Run docker-compose
	docker-compose up --build -d --scale scrapper=5 && docker-compose logs -f
.PHONY: compose-up

compose-down: ### Down docker-compose
	docker-compose down --remove-orphans -v
.PHONY: compose-down