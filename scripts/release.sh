#!/bin/bash
# GitHub Release Script
# Usage: ./scripts/release.sh v1.1.0

set -e

VERSION="$1"

if [ -z "$VERSION" ]; then
    echo "❌ Version is required"
    echo "Usage: ./scripts/release.sh v1.1.0"
    exit 1
fi

echo "🚀 Creating GitHub Release: $VERSION"
echo "======================================"

# Check if gh CLI is installed
if command -v gh &> /dev/null; then
    echo "✅ GitHub CLI found, triggering workflow..."
    
    # Ask for pre-release confirmation
    read -p "Is this a pre-release? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        PRERELEASE="true"
    else
        PRERELEASE="false"
    fi
    
    # Trigger workflow
    gh workflow run release.yml \
        -f version="$VERSION" \
        -f prerelease="$PRERELEASE"
    
    echo "✅ Release workflow triggered!"
    echo "🌐 Check status: https://github.com/Madraka/GONews/actions"
    
else
    echo "⚠️  GitHub CLI not found"
    echo "🌐 Please go to: https://github.com/Madraka/GONews/actions/workflows/release.yml"
    echo "🎯 Click 'Run workflow' and enter:"
    echo "   - Version: $VERSION"
    echo "   - Pre-release: false (unless this is a pre-release)"
fi

echo ""
echo "📝 Don't forget to update CHANGELOG.md!"
echo "🎉 Release $VERSION will be created automatically"
