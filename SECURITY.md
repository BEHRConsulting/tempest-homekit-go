# Security and Sensitive Data Cleanup Guide

## Overview

This document explains how to handle sensitive data (API tokens, station names) that was previously committed to the Git repository.

## What Was Changed (October 2025)

### Sanitized Files
The following files were updated to remove hardcoded sensitive data:

1. **pkg/config/config.go**
 - Changed default token from hardcoded value to empty string
 - Changed default station name from "Chino Hills" to empty string
 - Updated help text to indicate these are required fields

2. **.env and .env.example**
 - Replaced actual token with placeholder: `your-token-here`
 - Replaced station name with placeholder: `Your Station Name`

3. **README.md**
 - Updated documentation to show fields as required
 - Removed examples with real station names

### Protected Files
The following files are in `.gitignore` to prevent accidental commits:
- `.env` - Your personal configuration (never commit this!)
- `.env.local`
- `.env.*.local`
- `db/` - HomeKit pairing data (contains device-specific data)

## Git History Cleanup

### Option 1: Rewrite History (Recommended for Private Repos)

If this is a private repository and you want to completely remove sensitive data from history:

```bash
# Install git-filter-repo (if not already installed)
# macOS: brew install git-filter-repo
# Linux: pip install git-filter-repo

# Backup your repository first!
cd /path/to/tempest-homekit-go
git clone . ../tempest-homekit-go-backup

# Replace sensitive data in history
git filter-repo --replace-text <(cat <<EOF
b88edc78-6261-414e-8042-86a4d4f9ba15=your-token-here
Chino Hills=Your Station Name
178915=12345
EOF
)

# Force push to remote (WARNING: This rewrites history!)
git push --force --all origin
git push --force --tags origin
```

**WARNING**: This rewrites Git history and requires all collaborators to re-clone the repository!

### Option 2: Use Git-Secrets (Prevent Future Leaks)

Install git-secrets to prevent accidentally committing sensitive data:

```bash
# macOS
brew install git-secrets

# Linux
git clone https://github.com/awslabs/git-secrets
cd git-secrets
make install

# Install in this repo
cd /path/to/tempest-homekit-go
git secrets --install
git secrets --register-aws # Optional: AWS patterns

# Add custom patterns for this project
git secrets --add 'TEMPEST_TOKEN=[A-Za-z0-9-]+'
git secrets --add '[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}'
```

### Option 3: Rotate Credentials (Recommended for Public Repos)

If this will be a public repository:

1. **Revoke the exposed token:**
 - Go to https://tempestwx.com/settings/tokens
 - Delete the old token: `b88edc78-6261-414e-8042-86a4d4f9ba15`
 - Generate a new token

2. **Update your local .env:**
 ```bash
 # Edit your local .env file
 nano .env
 # Replace with new token
 TEMPEST_TOKEN=new-token-here
 ```

3. **Accept that old commits contain the old token:**
 - Since you've revoked it, it's no longer a security risk
 - Future commits won't contain sensitive data
 - Document this in your README

### Option 4: Keep as Example/Test Repo

If this token was always meant for examples:

1. **Document it clearly:**
 - Add note in README that token is for testing/examples only
 - Clarify it's not a real production token
 - Add warning about not using example credentials

2. **No action needed** - Just update documentation

## Current Protection Measures

### Files Protected by .gitignore
```
.env # Your personal config
.env.local # Local overrides
.env.*.local # Environment-specific
db/ # HomeKit database
*.log # Log files
```

### Configuration Priority
1. Command-line flags (highest priority)
2. Environment variables from `.env`
3. Default values (now empty strings for sensitive data)

### Required Setup for New Users

New users must create their own `.env` file:

```bash
# 1. Copy example
cp .env.example .env

# 2. Edit with real values
nano .env

# 3. Add your token and station name
TEMPEST_TOKEN=your-actual-token
TEMPEST_STATION_NAME=Your Actual Station
```

## Verification

Check that sensitive data is not in current code:

```bash
# Search for hardcoded tokens (should find none in code)
git grep "b88edc78-6261-414e-8042-86a4d4f9ba15" -- '*.go' '*.md'

# Verify .env is in .gitignore
git check-ignore .env # Should output: .env

# Check what would be committed
git status --ignored
```

## Best Practices Moving Forward

1. **Never hardcode credentials** in source files
2. **Always use .env files** for sensitive configuration
3. **Keep .env in .gitignore** (already done)
4. **Use .env.example** for documentation only
5. **Rotate tokens** if accidentally committed
6. **Use git-secrets** to prevent future accidents

## References

- [Git Filter-Repo Documentation](https://github.com/newren/git-filter-repo)
- [Git-Secrets](https://github.com/awslabs/git-secrets)
- [GitHub: Removing Sensitive Data](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/removing-sensitive-data-from-a-repository)
- [WeatherFlow Token Management](https://tempestwx.com/settings/tokens)
