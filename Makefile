# start the orchestration
start:
	docker compose -f docker-compose.yml up

# stop the orchestration
stop:
	docker compose stop

# remove all the ArangoDB instance containers and relevant images
remove:
	docker compose down -v
	docker rmi arangodb:3.9.10
	docker rmi anti-pattern-graph-checker-single-test:latest

# completely remove the persisting volumes
clean:
	sudo rm -rf logs nodes