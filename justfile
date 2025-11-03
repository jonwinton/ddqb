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

release:
	#!/bin/bash
	set -e

	echo "-- Analyzing commits to determine next version..."

	# Get current version
	current_version=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
	echo "-- Current version: $current_version"

	# Calculate next version based on conventional commits
	if [ -n "$1" ]; then
		# Manual override
		case $1 in
			major)
				next_version=$(svu major)
				;;
			minor)
				next_version=$(svu minor)
				;;
			patch)
				next_version=$(svu patch)
				;;
			*)
				echo "-- Invalid argument. Use: major, minor, or patch"
				exit 1
				;;
		esac
		echo "-- Manual version bump: $next_version ($1)"
	else
		# Automatic detection based on conventional commits
		next_version=$(svu next)
		echo "-- Auto-detected next version: $next_version"
	fi

	# Preview changelog
	echo ""
	echo "-- Preview of changes for $next_version (since $current_version):"
	echo "--------------------------------------------------------------------------------"
	git-cliff --config cliff.toml "$current_version..HEAD" --tag "$next_version" --strip all
	echo "--------------------------------------------------------------------------------"
	echo ""

	# Confirm
	read -p "-- Create release $next_version? [y/N] " -n 1 -r
	echo
	if [[ ! $REPLY =~ ^[Yy]$ ]]; then
		echo "Release cancelled"
		exit 1
	fi

	# Create and push tag
	echo "-- Creating tag $next_version..."
	git tag "$next_version"

	echo "-- Pushing tag to origin..."
	git push origin "$next_version"

	echo "-- Changelog for $next_version (single release):"
	echo "--------------------------------------------------------------------------------"
	git-cliff --config cliff.toml "$current_version..$next_version" --tag "$next_version" --strip all
	echo "--------------------------------------------------------------------------------"

	echo ""
	echo "-- Release $next_version initiated!"
	echo ""
	echo "-- GitHub Actions will now:"
	echo "   1. Generate changelog"
	echo "   2. Create GitHub release"
	echo "   3. Publish release notes"
	echo ""
	echo "-- Monitor progress: https://github.com/jonwinton/ddqb/actions"

