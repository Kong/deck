#!/bin/bash

# Script to get commit SHAs for GitHub Actions
# Usage: ./get-action-commit-shas.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Fetching commit SHAs for GitHub Actions...${NC}"
echo

# Function to get commit SHA for a GitHub action
get_commit_sha() {
    local action_ref=$1
    local repo=$(echo $action_ref | cut -d'@' -f1)
    local tag_or_sha=$(echo $action_ref | cut -d'@' -f2)
    
    # Skip if already a commit SHA (40 characters)
    if [[ ${#tag_or_sha} -eq 40 ]]; then
        echo -e "${GREEN}$action_ref${NC} (already pinned)"
        return
    fi
    
    # Skip master/main branches for now
    if [[ "$tag_or_sha" == "master" || "$tag_or_sha" == "main" ]]; then
        echo -e "${YELLOW}$action_ref${NC} (branch reference - consider pinning)"
        return
    fi
    
    echo -n "Fetching SHA for $repo@$tag_or_sha... "
    
    # Try multiple GitHub API endpoints to get the commit SHA
    local sha=""
    
    # First try as a tag reference
    local api_url="https://api.github.com/repos/$repo/git/refs/tags/$tag_or_sha"
    local response=$(curl -s -w "%{http_code}" "$api_url" 2>/dev/null)
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [[ "$http_code" == "200" ]]; then
        sha=$(echo "$body" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data['object']['sha'] if data['object']['type'] == 'commit' else '')" 2>/dev/null)
        if [[ -z "$sha" ]]; then
            # It's a tag object, get the commit it points to
            local tag_sha=$(echo "$body" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data['object']['sha'])" 2>/dev/null)
            if [[ -n "$tag_sha" ]]; then
                local tag_response=$(curl -s "https://api.github.com/repos/$repo/git/tags/$tag_sha" 2>/dev/null)
                sha=$(echo "$tag_response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data['object']['sha'])" 2>/dev/null)
            fi
        fi
    fi
    
    # If tag approach didn't work, try as a branch/commit reference
    if [[ -z "$sha" ]]; then
        api_url="https://api.github.com/repos/$repo/commits/$tag_or_sha"
        response=$(curl -s -w "%{http_code}" "$api_url" 2>/dev/null)
        http_code="${response: -3}"
        body="${response%???}"
        
        if [[ "$http_code" == "200" ]]; then
            sha=$(echo "$body" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data['sha'])" 2>/dev/null)
        fi
    fi
    
    if [[ -n "$sha" && ${#sha} -eq 40 ]]; then
        echo -e "${GREEN}✓${NC}"
        echo "  $repo@$sha # $tag_or_sha"
    else
        echo -e "${RED}✗ (could not fetch SHA)${NC}"
        echo -e "    ${YELLOW}Manual lookup: https://github.com/$repo/releases/tag/$tag_or_sha${NC}"
    fi
}

# Extract all GitHub Actions from workflow files
echo "Scanning workflow files for GitHub Actions..."
echo

# Find all unique action references
actions=$(grep -h "uses:" .github/workflows/*.yaml | \
          sed 's/.*uses: *//' | \
          sed 's/ *#.*//' | \
          sort -u)

echo "Found the following actions:"
echo "$actions"
echo
echo "Fetching commit SHAs:"
echo

# Process each action
while IFS= read -r action; do
    get_commit_sha "$action"
done <<< "$actions"

echo
echo -e "${YELLOW}Note: You can copy the output above to update your workflow files.${NC}"
echo -e "${YELLOW}Remember to also update the comment with the version tag.${NC}"
