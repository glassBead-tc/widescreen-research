# Contributing to widescreen-research

Thank you for your interest in contributing! This guide will help you get started.

## Development Setup

### Prerequisites
- Go 1.22 or later
- Node.js 18+ LTS (for examples only)
- Docker (optional, for container builds)
- Git

### Initial Setup

1. Clone the repository:
```bash
git clone https://github.com/glassBead-tc/widescreen-research.git
cd widescreen-research
```

2. Install development tools and Git hooks:
```bash
make install-hooks
```

This will install:
- pre-commit hooks for code quality
- golangci-lint for Go linting
- hadolint for Dockerfile linting
- Security scanning tools

## Quality Standards

This project uses automated quality checks via Git hooks. All code must pass these checks before being merged.

### Code Formatting
- Go code is formatted with `gofmt` and `goimports`
- YAML files follow yamllint rules
- Markdown follows markdownlint standards
- All files use LF line endings

### Linting
Run all linters:
```bash
make lint
```

Run specific linters:
```bash
make lint-go      # Go linting with golangci-lint
make lint-docker  # Dockerfile linting with hadolint
```

### Testing
All code changes must include tests:
```bash
make test
```

### Security Scanning
Run security checks:
```bash
make security-scan
```

### Commit Message Format
We use Conventional Commits format:
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `style:` Formatting changes (no code logic changes)
- `refactor:` Code restructuring
- `perf:` Performance improvements
- `test:` Test additions or fixes
- `build:` Build system changes
- `ci:` CI/CD changes
- `chore:` Maintenance tasks

Examples:
```
feat: add distributed task execution to coordinator
fix: resolve memory leak in drone worker pool
docs: update API documentation for new endpoints
```

### Pre-Push Checklist
Before pushing code, run:
```bash
make pre-push
```

This runs:
1. All linters
2. All tests
3. Security scans

## Development Workflow

### 1. Create a feature branch
```bash
git checkout -b feature/your-feature-name
```

### 2. Make your changes
- Write code following Go best practices
- Add tests for new functionality
- Update documentation as needed

### 3. Run quality checks
```bash
make lint
make test
```

### 4. Commit your changes
```bash
git add .
git commit -m "feat: your feature description"
```

The pre-commit hooks will automatically:
- Format your code
- Run linters
- Check for security issues
- Validate commit message format

### 5. Push and create a PR
```bash
git push origin feature/your-feature-name
```

## Project Structure

```
.
├── cmd/                    # Application entry points
│   ├── coordinator/        # Coordinator service
│   └── drone/              # Drone worker
├── pkg/                    # Reusable packages
│   ├── mcp/                # MCP server implementation
│   ├── gcp/                # GCP client utilities
│   └── coordinator/        # Coordinator logic
├── examples/               # Example implementations
│   └── node/               # Node.js examples
├── scripts/                # Development scripts
└── docs/                   # Documentation
```

## Running Services Locally

### Coordinator
```bash
make run-coordinator
```

### Drone
```bash
make run-drone
```

### With Docker
```bash
make docker
docker run --rm -p 8080:8080 widescreen/coordinator:dev
```

## VS Code Setup

This project includes VS Code configuration for optimal development experience:

1. Install recommended extensions when prompted
2. Go formatting and imports are configured to run on save
3. Debug configurations are available for coordinator and drone

## Getting Help

- Check existing issues before creating new ones
- Use discussions for questions and ideas
- Join our community chat (link TBD)

## Code of Conduct

Please be respectful and professional in all interactions. We're committed to providing a welcoming and inclusive environment for all contributors.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.