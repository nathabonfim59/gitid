#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Ensure we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo -e "${RED}Error: This is not a git repository${NC}"
    exit 1
fi

# Check if working directory is clean
if ! git diff-index --quiet HEAD --; then
    echo -e "${RED}Error: Working directory is not clean. Please commit or stash changes first.${NC}"
    exit 1
fi

# Get the current version from the latest tag
current_version=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

# Parse the version numbers
major=$(echo $current_version | sed 's/v\([0-9]*\).\([0-9]*\).\([0-9]*\)/\1/')
minor=$(echo $current_version | sed 's/v\([0-9]*\).\([0-9]*\).\([0-9]*\)/\2/')
patch=$(echo $current_version | sed 's/v\([0-9]*\).\([0-9]*\).\([0-9]*\)/\3/')

# Show current version and ask for the type of release
echo "Current version: $current_version"
echo "Select release type:"
echo "1) Major (v$((major+1)).0.0)"
echo "2) Minor (v$major.$((minor+1)).0)"
echo "3) Patch (v$major.$minor.$((patch+1)))"
read -p "Enter choice [1-3]: " choice

case $choice in
    1)
        new_version="v$((major+1)).0.0"
        ;;
    2)
        new_version="v$major.$((minor+1)).0"
        ;;
    3)
        new_version="v$major.$minor.$((patch+1))"
        ;;
    *)
        echo -e "${RED}Invalid choice${NC}"
        exit 1
        ;;
esac

# Confirm the new version
read -p "Create release $new_version? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${RED}Release cancelled${NC}"
    exit 1
fi

# Create and push the new tag
echo "Creating new tag $new_version..."
git tag -a "$new_version" -m "Release $new_version"
git push origin "$new_version"

# Build the release
echo "Building release..."
make release

# Create GitHub release if gh CLI is available
if command -v gh &> /dev/null; then
    echo "Creating GitHub release..."
    gh release create "$new_version" \
        --title "Release $new_version" \
        --generate-notes \
        release/*

    echo -e "${GREEN}Successfully created release $new_version${NC}"
else
    echo -e "${RED}Warning: GitHub CLI (gh) not found. Skipping GitHub release creation.${NC}"
    echo -e "${GREEN}Successfully created tag $new_version${NC}"
    echo "Please create the release manually on GitHub and upload the artifacts from the release directory."
fi