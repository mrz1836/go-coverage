# ⚡ Quickstart Guide

Get **go-coverage** running in your project in under 5 minutes. This self-contained coverage system replaces Codecov with zero external dependencies.

## 📦 Installation

### Option 1: Install CLI Tool
```bash
go install github.com/mrz1836/go-coverage/cmd/go-coverage@latest
```

### Option 2: Use as Library
```bash
go get -u github.com/mrz1836/go-coverage
```

## 🚀 Quick Setup

### 1. Configure GitHub Pages

Run the setup command to configure GitHub Pages automatically:

```bash
# Auto-detect repository from git remote
go-coverage setup-pages

# Or specify repository explicitly
go-coverage setup-pages owner/repo

# Preview changes without making them
go-coverage setup-pages --dry-run
```

### 2. Add to GitHub Actions

Create or update `.github/workflows/coverage.yml`:

```yaml
name: Coverage
on: [push, pull_request]

jobs:
  coverage:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Run Tests with Coverage
        run: go test -coverprofile=coverage.txt ./...

      - name: Generate Coverage Reports
        run: go-coverage complete -i coverage.txt
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./coverage
```

### 3. First Run

Commit and push your changes. Your coverage reports will be available at:

- **Dashboard**: `https://yourname.github.io/yourrepo/`
- **Badge**: `https://yourname.github.io/yourrepo/coverage.svg`

## 🎯 Basic Usage

### Generate Complete Coverage Report

```bash
# Run tests and generate coverage
go test -coverprofile=coverage.txt ./...

# Process coverage with complete pipeline
go-coverage complete -i coverage.txt
```

### Individual Commands

```bash
# Parse coverage data only
go-coverage parse -i coverage.txt

# Generate PR comment
go-coverage comment --pr 123 --coverage coverage.txt

# View coverage history
go-coverage history --branch main --days 30

# Check for updates
go-coverage upgrade --check
```

## 📊 What You Get

After setup, your project automatically generates:

- 🏷️ **Live Coverage Badge** - Real-time SVG badge for README
- 📈 **Interactive Dashboard** - Beautiful coverage visualization
- 📊 **Detailed Reports** - File-level coverage analysis
- 🔄 **PR Comments** - Automatic coverage analysis on pull requests
- 📈 **History Tracking** - Coverage trends over time

## ⚙️ Quick Configuration

Create `.go-coverage.json` for custom settings:

```json
{
  "coverage": {
    "threshold": 80.0,
    "exclude_paths": ["vendor/", "test/"],
    "exclude_files": ["*.pb.go", "*_gen.go"]
  },
  "badge": {
    "style": "flat",
    "logo": "go"
  },
  "report": {
    "title": "My Project Coverage",
    "theme": "dark"
  }
}
```

## 🆘 Troubleshooting

### Setup Issues

```bash
# Verify installation
go-coverage --version

# Test with dry run
go-coverage setup-pages --dry-run

# Enable debug mode
go-coverage complete --debug -i coverage.txt
```

### Common Problems

- **No coverage badge**: Ensure GitHub Pages is enabled and deployed
- **Missing reports**: Check GitHub Actions logs for errors
- **Permission errors**: Verify `GITHUB_TOKEN` has required permissions

## 📚 Next Steps

- [**User Guide**](user-guide.md) - Complete usage documentation
- [**CLI Reference**](cli-reference.md) - All command options
- [**Configuration**](configuration.md) - Advanced configuration options
- [**Contributing**](contributing.md) - How to contribute to the project

## 💡 Pro Tips

1. **Use the badge in your README**:
   ```markdown
   ![Coverage](https://yourname.github.io/yourrepo/coverage.svg)
   ```

2. **Set up branch protection** to require coverage thresholds

3. **Customize themes** and styles to match your project branding

4. **Use PR comments** to track coverage changes in code reviews

---

That's it! You now have a complete, self-hosted coverage system with zero external dependencies. 🎉
