compose-up: ### Run docker-compose
	docker-compose up --build -d postgres rabbitmq && docker-compose logs -f
.PHONY: compose-up

compose-down: ### Down docker-compose
	docker-compose down --remove-orphans
.PHONY: compose-down