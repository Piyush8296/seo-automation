.PHONY: setup build-ui build run dev clean

# Install all dependencies (Go modules + frontend npm packages)
setup:
	cd ui && npm install

# Build the React frontend
build-ui:
	cd ui && npm run build

# Build the Go binary
build:
	go build -o seo-audit .

# Build everything and start the server
run: build build-ui
	./seo-audit serve --port 8080 --ui-dir ui/dist

# Development mode: start Go server + Vite dev server concurrently
dev: build
	./seo-audit serve --port 8080 & \
	cd ui && npm run dev & \
	wait

# Build and run with Docker
docker:
	docker compose up --build

# Remove build artifacts
clean:
	rm -f seo-audit
	rm -rf ui/dist ui/node_modules
