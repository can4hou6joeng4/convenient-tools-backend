start:
	@docker compose -p convenient-tools up --build

stop:
	@docker compose -p convenient-tools rm -v --force --stop
	@docker rmi convenient-tools