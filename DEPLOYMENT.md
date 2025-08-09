# DrDuck Deployment Guide

This guide explains how to deploy DrDuck using GoReleaser for cross-platform distribution.

## Prerequisites

1. **GoReleaser installed**:
   ```bash
   # macOS
   brew install goreleaser
   
   # Linux
   curl -sfL https://goreleaser.com/static/run | bash
   
   # Or install via Go
   go install github.com/goreleaser/goreleaser/v2@latest
   ```

2. **GitHub repository setup**:
   - Repository: `https://github.com/SilverFlin/DrDuck`
   - GitHub token with repo permissions

3. **Optional: Package manager repositories**:
   - Homebrew tap: `https://github.com/SilverFlin/homebrew-tap`
   - Scoop bucket: `https://github.com/SilverFlin/scoop-bucket`

## Manual Release Process

### 1. Test the Release Configuration

```bash
# Dry run to check configuration
goreleaser check

# Build snapshot (test without releasing)
goreleaser build --snapshot --clean
```

### 2. Create and Push a Tag

```bash
# Create a new tag
git tag v0.1.0
git push origin v0.1.0

# Or create a tag with message
git tag -a v0.1.0 -m "Release v0.1.0: Initial DrDuck release"
git push origin v0.1.0
```

### 3. Run GoReleaser

```bash
# Set GitHub token
export GITHUB_TOKEN="your_github_token"

# Optional: Set package manager tokens
export HOMEBREW_TAP_GITHUB_TOKEN="your_homebrew_token"
export SCOOP_GITHUB_TOKEN="your_scoop_token"

# Run release
goreleaser release --clean
```

## Automated Release via GitHub Actions

The repository includes GitHub Actions workflows that automatically:

1. **On every push/PR** (`.github/workflows/test.yml`):
   - Run tests
   - Build binaries
   - Validate GoReleaser configuration

2. **On tag push** (`.github/workflows/release.yml`):
   - Build cross-platform binaries
   - Create GitHub release
   - Upload release assets
   - Update package managers (Homebrew, Scoop)

### GitHub Secrets Required

Add these secrets to your GitHub repository:

```bash
# Required
GITHUB_TOKEN                    # Automatic, provided by GitHub

# Optional (for package managers)
HOMEBREW_TAP_GITHUB_TOKEN      # Token for homebrew-tap repo
SCOOP_GITHUB_TOKEN             # Token for scoop-bucket repo
```

## What Gets Built

GoReleaser builds DrDuck for:

### Platforms
- **Linux**: amd64, 386, arm, arm64
- **macOS**: amd64, arm64 (Apple Silicon)
- **Windows**: amd64, 386

### Distribution Formats
- **Binary archives**: `.tar.gz` (Linux/macOS), `.zip` (Windows)
- **Homebrew formula**: For `brew install silverflin/tap/drduck`
- **Scoop manifest**: For `scoop install drduck`
- **Checksums**: SHA256 verification

### GitHub Release Assets
```
drduck_Darwin_x86_64.tar.gz
drduck_Darwin_arm64.tar.gz
drduck_Linux_x86_64.tar.gz
drduck_Linux_i386.tar.gz
drduck_Linux_arm64.tar.gz
drduck_Windows_x86_64.zip
drduck_Windows_i386.zip
checksums.txt
```

## Post-Release

After a successful release:

1. **GitHub Release** is created automatically with:
   - Release notes from git commits
   - Cross-platform binaries
   - Installation instructions

2. **Package Managers** are updated:
   - Homebrew users can: `brew install silverflin/tap/drduck`
   - Scoop users can: `scoop install drduck`

3. **Manual Installation**:
   - Users can download binaries from GitHub releases
   - Extract and add to PATH

## Version Numbering

Follow semantic versioning:
- `v1.0.0` - Major release
- `v1.1.0` - Minor release (new features)
- `v1.1.1` - Patch release (bug fixes)
- `v1.0.0-rc.1` - Release candidate
- `v1.0.0-beta.1` - Beta release

## Troubleshooting

### Common Issues

1. **GoReleaser fails to build**:
   ```bash
   # Check Go modules
   go mod tidy
   go mod verify
   ```

2. **GitHub token issues**:
   ```bash
   # Test token permissions
   curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user
   ```

3. **Package manager updates fail**:
   - Check repository permissions
   - Verify token scopes
   - Ensure tap/bucket repositories exist

### Local Testing

```bash
# Build for current platform only
go build -o drduck main.go

# Test version injection
go build -ldflags "-X main.version=v0.1.0 -X main.commit=$(git rev-parse HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o drduck main.go
./drduck --version
```

## Next Steps

1. Set up Homebrew tap repository
2. Create Scoop bucket repository  
3. Configure package manager tokens
4. Test release process with a pre-release tag