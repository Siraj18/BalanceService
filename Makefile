build_service_images:
	docker-compose build

run_service: build_service_images
	docker-compose up

run_tests:
	go test ./... -cover