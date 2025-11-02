[private]
default: help

[private]
help:
	@just --list


# Serves mkdocs locally
docs-serve:
	python3 -m pip install -r requirements.txt
	mkdocs serve

# Builds mkdocs
docs-build:
	python3 -m pip install -r requirements.txt
	mkdocs build

# Deploys mkdocs to GitHub Pages
docs-deploy:
	python3 -m pip install -r requirements.txt
	mkdocs gh-deploy --force

# Generates API docs
docs-api:
	gomarkdoc --output docs/api.md ./...

# Lints the code
lint:
	@echo "Checking formatting with gofumpt..."
	gofumpt -l -d .

	@echo "Running golangci-lint..."
	golangci-lint run

# Runs tests
test:
	gotestsum -f standard-verbose
