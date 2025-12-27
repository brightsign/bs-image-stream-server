 #!/bin/bash

  set -x  # Enable debug mode - shows every command

  # Get the current git branch name (works on older git and detached HEAD)
  branch_name=$(git rev-parse --abbrev-ref HEAD 2>/dev/null)

  echo "DEBUG: branch_name=[$branch_name]"

  # Check if we're in a git repository
  if [ -z "$branch_name" ] || [ "$branch_name" = "HEAD" ]; then
      echo "Error: Not in a git repository or no branch checked out"
      exit 1
  fi

  # Get commit message from command line argument or use default
  if [ -z "$1" ]; then
      commit_message="interim commit"
  else
      commit_message="$1"
  fi

  # Execute git commands
  git add .
  git commit -am "$commit_message"
  echo "DEBUG: About to push to origin $branch_name"
  git push -u origin "$branch_name"
  echo "DEBUG: Push exit code: $?"
