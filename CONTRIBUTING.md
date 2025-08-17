# Contributing to Cloudlet

Thank you for your interest in contributing to Cloudlet! We welcome contributions from the community and are excited to see what you can bring to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Guidelines](#contributing-guidelines)
- [Coding Standards](#coding-standards)
- [Submitting Changes](#submitting-changes)
- [Issue Reporting](#issue-reporting)
- [Feature Requests](#feature-requests)

## Code of Conduct

This project adheres to a simple code of conduct:

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Maintain a professional and friendly environment

## Getting Started

### Prerequisites

- Go 1.23.4 or later
- Git
- Basic knowledge of Go programming
- Familiarity with RESTful APIs
- Understanding of SQLite databases

### Development Setup

1. **Fork the repository**

   ```bash
   git clone https://github.com/yourusername/cloudlet.git
   cd cloudlet
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Build the application**

   ```bash
   go build -o cloudlet ./cmd/cloudlet
   ```

4. **Run tests**

   ```bash
   go test ./...
   ```

5. **Start the development server**
   ```bash
   ./cloudlet
   ```

The server will be available at `http://localhost:8080`.

## Contributing Guidelines

### Types of Contributions

We welcome several types of contributions:

- **🐛 Bug Fixes**: Fix issues and improve stability
- **🚀 New Features**: Add functionality that aligns with project goals
- **📖 Documentation**: Improve or add documentation
- **🧹 Refactoring**: Improve code quality and maintainability
- **🏎️ Performance**: Optimize existing functionality

### Before You Start

1. **Check existing issues** to see if your idea is already being discussed
2. **Create an issue** for significant changes to discuss the approach
3. **Keep changes focused** - one feature or fix per pull request
4. **Follow the coding standards** outlined below

## Coding Standards

### Go Code Style

- **Follow Go conventions**: Use `gofmt`, `go vet`, and `golint`
- **Use meaningful names**: Variables, functions, and packages should have clear, descriptive names
- **Keep functions small**: Aim for functions that do one thing well
- **Add comments**: Document public functions and complex logic
- **Handle errors properly**: Always handle errors explicitly

### Project Structure

Follow the existing architecture:

```
internal/
 handlers/          # HTTP request handlers
 models/           # Data structures and API contracts
 repository/       # Database operations
 server/           # HTTP server configuration
 services/         # Business logic
 utils/            # Utility functions
```

### Naming Conventions

- **Files**: Use snake_case (e.g., `file_service.go`)
- **Functions**: Use camelCase (e.g., `createDirectory`)
- **Constants**: Use UPPER_SNAKE_CASE (e.g., `MAX_FILE_SIZE`)
- **Interfaces**: Use descriptive names ending with 'er' when appropriate

### Error Handling

```go
// Good
result, err := someOperation()
if err != nil {
    return fmt.Errorf("failed to perform operation: %w", err)
}

// Avoid
result, _ := someOperation() // Never ignore errors
```

### API Design

- **Follow RESTful principles**
- **Use appropriate HTTP methods** (GET, POST, PUT, DELETE)
- **Return consistent JSON responses**
- **Include proper HTTP status codes**
- **Validate input data**

## Submitting Changes

### Pull Request Process

1. **Create a feature branch**

   ```bash
   git checkout -b feature/amazing-feature
   ```

2. **Make your changes**

   - Write code following the coding standards
   - Add or update tests
   - Update documentation if necessary

3. **Commit your changes**

   ```bash
   git add .
   git commit -m "Add amazing feature

   - Implement feature X
   - Add tests for feature X
   - Update documentation"
   ```

4. **Push to your fork**

   ```bash
   git push origin feature/amazing-feature
   ```

5. **Create a Pull Request**
   - Use a clear title and description
   - Reference any related issues
   - Include screenshots for UI changes
   - Ensure CI checks pass

### Commit Message Guidelines

Use clear, descriptive commit messages:

```
Add file upload validation

- Implement MIME type checking
- Add file size validation
- Include tests for validation logic
- Update API documentation

Fixes #123
```

### Pull Request Checklist

- [ ] Code follows project coding standards
- [ ] New tests added for new functionality
- [ ] Documentation updated if necessary
- [ ] Commit messages are clear and descriptive
- [ ] No breaking changes (or clearly documented)
- [ ] Security considerations reviewed

## Issue Reporting

### Before Reporting

1. **Search existing issues** to avoid duplicates
2. **Check the documentation** for solutions

### Bug Report Template

```
**Describe the bug**
A clear description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. See error

**Expected behavior**
What you expected to happen.

**Environment:**
- OS: [e.g. Windows, macOS, Linux]
- Go version: [e.g. 1.23.4]
- Cloudlet version: [e.g. v1.0.0]

**Additional context**
Add any other context about the problem here.
```

## Feature Requests

We love new ideas! When suggesting features:

1. **Check the roadmap** in README.md
2. **Describe the use case** clearly
3. **Explain the benefit** to other users
4. **Consider implementation complexity**
5. **Be open to discussion** about the approach

### Feature Request Template

```
**Is your feature request related to a problem?**
A clear description of what the problem is.

**Describe the solution you'd like**
A clear description of what you want to happen.

**Describe alternatives you've considered**
Any alternative solutions or features you've considered.

**Additional context**
Add any other context or screenshots about the feature request here.
```

## Versioning

We use [Semantic Versioning](https://semver.org/):

- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [REST API Design Best Practices](https://restfulapi.net/)

## Getting Help

If you need help or have questions:

1. **Check the documentation** first
2. **Look through existing issues** for similar questions
3. **Create a new issue** with the "question" label
4. **Be specific** about what you're trying to achieve

## Recognition

Contributors will be recognized in:

- README.md acknowledgments
- Release notes for significant contributions
- GitHub contributor statistics

Thank you for contributing to Cloudlet! 🙏
