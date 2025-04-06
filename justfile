DOCKER_REPO := "subotic/valuation-go"
CARGO_VERSION := `cargo metadata --format-version=1 --no-deps | jq --raw-output '.packages[].version'`
COMMIT_HASH := `git log --pretty=format:'%h' -n 1`
GIT_TAG := `git describe --tags --exact-match 2>/dev/null || true`
IMAGE_TAG := if GIT_TAG == "" { CARGO_VERSION + "-" + COMMIT_HASH } else { CARGO_VERSION }
DOCKER_IMAGE := DOCKER_REPO + ":" + IMAGE_TAG

# List all recipes
default:
    just --list --unsorted

# 🔧 Install required CLI tools
setup:
    @echo "Installing dependencies..."
    go install github.com/a-h/templ/cmd/templ@latest
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    @echo "Optional: install entr for file watching: https://github.com/eradman/entr"
    @echo "Try: brew install entr  # macOS"
    @echo "Or:  sudo apt install entr  # Debian/Ubuntu"

# 📦 Tidy go.mod and fetch dependencies
deps:
    @echo "Tidying Go modules..."
    go mod tidy

# 🧱 Build binaries
build:
    @echo "Building CLI and HTTP binaries..."
    go build -o bin/cli ./cmd/cli
    go build -o bin/http ./cmd/http

# 🚀 Run the HTTP server
run-http:
    @echo "Running HTTP server..."
    go run ./cmd/http

# 🖥️ Run the CLI
run-cli:
    @echo "Running CLI..."
    go run ./cmd/cli

# 🧪 Run tests
test:
    @echo "Running tests..."
    go test ./...

# 🧹 Format code
fmt:
    @echo "Formatting code..."
    go fmt ./...

# 🔍 Lint with golangci-lint
lint:
    @echo "Linting code..."
    golangci-lint run ./...

# 🧼 Clean build artifacts
clean:
    @echo "Cleaning up..."
    rm -rf bin/

# 🛠️ Generate code from .templ files
generate:
    @echo "Generating Templ code..."
    templ generate

# 🔄 Dev mode: generate + run
dev:
    just generate
    just run-http

# 👀 Watch .templ files and regenerate (requires entr)
watch:
    @echo "Watching .templ files for changes..."
    find . -name '*.templ' | entr just generate

# Build linux/amd64 Docker image locally
docker-build-amd64:
    docker buildx build --platform linux/amd64 -t {{ DOCKER_IMAGE }}-amd64 --load .

# Push previously build linux/amd64 image to Docker hub
docker-push-amd64:
    docker push {{ DOCKER_IMAGE }}-amd64

# Build linux/arm64 Docker image locally
docker-build-arm64:
    docker buildx build --platform linux/arm64 -t {{ DOCKER_IMAGE }}-arm64 --load .

# Push previously build linux/arm64 image to Docker hub
docker-push-arm64:
    docker push {{ DOCKER_IMAGE }}-arm64

# Publish Docker manifest combining aarch64 and x86 published images
docker-publish-manifest:
    docker manifest create {{ DOCKER_IMAGE }} --amend {{ DOCKER_IMAGE }}-amd64 --amend {{ DOCKER_IMAGE }}-arm64
    docker manifest annotate --arch amd64 --os linux {{ DOCKER_IMAGE }} {{ DOCKER_IMAGE }}-amd64
    docker manifest annotate --arch arm64 --os linux {{ DOCKER_IMAGE }} {{ DOCKER_IMAGE }}-arm64
    docker manifest inspect {{ DOCKER_IMAGE }}
    docker manifest push {{ DOCKER_IMAGE }}

# Output the BUILD_TAG
docker-image-tag:
    @echo {{ IMAGE_TAG }}
