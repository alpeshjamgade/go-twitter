AUTH_BINARY=authApp
USER_BINARY=userApp

## up: starts all containers in the background without forcing build
up:
	@echo "Starting Docker images..."
	docker-compose up -d
	@echo "Docker images started!"

## up_build: stops docker-compose (if running), builds all projects and starts docker compose
up_build: build_user build_auth
	@echo "Stopping docker images (if running...)"
	docker-compose down
	@echo "Building (when required) and starting docker images..."
	docker-compose up --build -d
	@echo "Docker images built and started!"

## down: stop docker compose
down:
	@echo "Stopping docker compose..."
	docker-compose down
	@echo "Done!"

## build_auth: builds the authentication service as linux executable
build_auth:
	@echo "Building auth binary"
	cd authentication-service && GOOS=linux CGO_ENABLED=0 go build -o ./build/${AUTH_BINARY} ./cmd/api
	@echo "Done!"

## build_user: builds the user service as a linux executable
build_user:
	@echo "Building user binary..."
	cd user-service && env GOOS=linux CGO_ENABLED=0 go build -o ./build/${USER_BINARY} ./cmd/api
	@echo "Done!"