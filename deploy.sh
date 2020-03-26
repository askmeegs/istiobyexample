#!/bin/sh

# If a command fails then the deploy stops
set -e

printf "\033[0;32mDeploying updates to GitHub...\033[0m\n"

# Build the project.
#   -t: set theme
#   --minift: compaction
#   --enableGitInfo: use commit timestamp as lastmod
hugo -t hugo-notepadium --minify --enableGitInfo

# Go To Public folder
cd docs

# Add changes to git.
git add .

# Commit changes.
msg="rebuilding site $(date)"
if [ -n "$*" ]; then
	msg="$*"
fi
git commit -m "$msg"

# Push source and build repos.
git push origin master
