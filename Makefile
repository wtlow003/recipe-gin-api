container-up:
	@echo "Starting Containers..."
	docker-compose up --build --no-start
	docker-compose restart

container-down:
	@echo "Tearing Containers..."
	docker-compose down --remove-orphans