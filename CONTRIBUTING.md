# Contributing to Tempus

First off, thank you for considering contributing to Tempus! 🎉

Tempus is a neurodivergent-friendly calendar event generator, and we welcome contributions that help make it even better for the ADHD and neurodivergent community.

## Code of Conduct

This project adheres to the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When you create a bug report, include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples** (command you ran, input you provided, etc.)
- **Describe the behavior you observed** and what you expected
- **Include your environment details**: OS, Go version, Tempus version

**Example bug report:**
```
Title: "Template create medication fails with duration > 1 hour"

Steps to reproduce:
1. Run `tempus template create medication`
2. Enter duration "2h"
3. Command exits with error

Expected: Event created with 2-hour duration
Actual: Error: "duration must be <= 1h"

Environment:
- OS: Ubuntu 22.04
- Go: 1.24
- Tempus: v0.5.0
```

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion:

- **Use a clear and descriptive title**
- **Provide a detailed description** of the suggested enhancement
- **Explain why this enhancement would be useful** to most Tempus users
- **Include examples** of how it would work

**ADHD-friendly features are particularly welcome!** If your enhancement helps neurodivergent users, please highlight that.

### Pull Requests

1. **Fork the repo** and create your branch from `main`:
   ```bash
   git checkout -b feature/my-awesome-feature
   ```

2. **Follow the coding standards**:
   - Run `gofmt` on your code
   - Follow Go best practices
   - Add godoc comments for exported functions
   - Keep functions small and focused

3. **Write tests** for new functionality:
   - Add unit tests in `*_test.go` files
   - Ensure all tests pass: `go test ./...`
   - Aim for >80% coverage on new code

4. **Update documentation**:
   - Update README.md if adding new commands or features
   - Update relevant docs in `docs/` directory
   - Add examples for new templates

5. **Commit your changes** with clear messages:
   ```
   # Good commit messages:
   feat: add pomodoro template for focus sessions
   fix: correct timezone handling in appointment template
   docs: expand Google API setup instructions
   test: add tests for duration parsing edge cases

   # Less helpful:
   update stuff
   fixes
   changes
   ```

6. **Run the full test suite** before submitting:
   ```bash
   # Format code
   gofmt -w .

   # Vet code
   go vet ./...

   # Run tests
   go test ./... -v

   # Build to ensure no compilation errors
   go build .
   ```

7. **Push to your fork** and submit a pull request to `main`

8. **Wait for review** - maintainers will review your PR and may suggest changes

## Development Setup

### Prerequisites

- Go 1.24 or later
- Git

### Getting Started

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/tempus.git
cd tempus

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o tempus .

# Run locally
./tempus --help
```

### Project Structure

```
tempus/
├── main.go                 # CLI commands and entry point
├── internal/
│   ├── calendar/          # ICS generation (RFC 5545)
│   ├── config/            # Configuration management
│   ├── normalizer/        # Date/time normalization
│   ├── prompts/           # Interactive UI helpers
│   ├── templates/         # Template system
│   └── utils/             # Shared utilities
├── docs/                  # Documentation
├── examples/              # Example ICS files
└── locales/               # Internationalization
```

### Adding a New Template

Templates are defined in [internal/templates/templates.go](internal/templates/templates.go). To add a new template:

1. **Define the template structure** in the `templates` map:
   ```go
   tm.templates["my-template"] = &Template{
       Name: "my-template",
       Description: "Brief description (ADHD-friendly if applicable)",
       Fields: []Field{
           {Key: "title", Name: "Event Title", Type: "text", Required: true},
           // ... more fields
       },
       Generator: generateMyTemplateEvent,
   }
   ```

2. **Implement the generator function**:
   ```go
   func generateMyTemplateEvent(fields map[string]string, tz string) (*calendar.Event, error) {
       // Parse fields, create event
       // Add ADHD-friendly features (alarms, duration defaults, etc.)
   }
   ```

3. **Add tests** in `main_template_test.go`:
   ```go
   func TestTemplateCreateMyTemplate(t *testing.T) {
       // Test happy path and edge cases
   }
   ```

4. **Update documentation** in README.md under "Templates" section

### ADHD-Friendly Design Principles

When contributing, keep these principles in mind:

- ✅ **Provide sensible defaults** - minimize decision fatigue
- ✅ **Support time-only input** - "10:30" instead of full datetime
- ✅ **Parse human durations** - "45m", "1h30m", "90 minutes"
- ✅ **Add multiple alarms** - help with time blindness
- ✅ **Use visual indicators** - emojis for categories
- ✅ **Include transition time** - buffer between activities
- ✅ **Show required fields clearly** - with `*` marker
- ✅ **Avoid overwhelming choices** - progressive disclosure

## Coding Standards

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Keep functions under 50 lines when possible
- Prefer simple, readable code over clever code

### Error Handling

```go
// Good: specific error messages
if err != nil {
    return fmt.Errorf("failed to parse duration %q: %w", input, err)
}

// Less helpful: generic errors
if err != nil {
    return err
}
```

### Testing

```go
// Use table-driven tests for multiple cases
func TestSlugify(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"simple", "hello", "hello"},
        {"spaces", "hello world", "hello-world"},
        // ...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Slugify(tt.input)
            if result != tt.expected {
                t.Errorf("got %q, want %q", result, tt.expected)
            }
        })
    }
}
```

## Security

- **Never commit secrets** - use environment variables
- **Validate all user input** - prevent injection attacks
- **Report security issues privately** - see [SECURITY.md](SECURITY.md)

## Questions?

- Open an issue with the `question` label
- Check existing issues and discussions
- Read the [documentation](docs/)

## Recognition

Contributors will be recognized in the project's README. Thank you for making Tempus better! 🙏
