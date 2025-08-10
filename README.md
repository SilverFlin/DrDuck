# DrDuck 🦆

**DocOps CLI tool for automated documentation workflows**

DrDuck is a command-line tool that integrates with AI coding assistants (Claude Code CLI, Cursor) to automate the creation and management of Architectural Decision Records (ADRs) and other documentation following DocOps principles.

## Features

- 🤖 **AI Integration**: Works with Claude Code CLI and Cursor for intelligent ADR generation
- 📝 **ADR Management**: Create, list, and manage Architectural Decision Records
- 🔄 **Git Hooks**: Optional pre-commit and pre-push hooks for documentation validation  
- 📁 **Flexible Storage**: Store ADRs in the same repo or a separate documentation repository
- 🎨 **Templates**: Support for Nygard, MADR, simple, and custom ADR templates
- ⚡ **Interactive Setup**: Beautiful CLI prompts for configuration

## Quick Start

### Installation

```bash
# npm (recommended - works on all platforms, no Go required)
npm install -g drduck

# Go install (if you have Go installed)
go install github.com/SilverFlin/DrDuck@latest

# Manual download (all platforms)
# Download from: https://github.com/SilverFlin/DrDuck/releases
```

### Initialize a Project

```bash
# Interactive setup
drduck init

# This will prompt you to choose:
# - AI provider (Claude Code CLI or Cursor)
# - Documentation storage (same repo or separate repo)
# - ADR template format (Nygard, MADR, simple, custom)
# - Git hooks (pre-commit, pre-push)
```

### Create Your First ADR

```bash
# Create a new ADR
drduck new -n "use-postgresql-database"

# List all ADRs
drduck list
```

## Commands

- `drduck init` - Initialize DrDuck in the current project
- `drduck new -n "name"` - Create a new ADR
- `drduck list` - List all ADRs with status
- `drduck --version` - Show version information
- `drduck --help` - Show help information

## Configuration

DrDuck stores configuration in `.drduck/config.yml`:

```yaml
ai_provider: "claude-code"     # or "cursor"
doc_storage: "same-repo"       # or "separate-repo"
adr_template: "nygard"        # or "madr", "simple", "custom"
hooks:
  pre_commit: true            # Install pre-commit hook
  pre_push: false            # Install pre-push hook
doc_path: "docs/adrs"        # ADR storage path (same-repo)
separate_repo_url: ""        # Separate repo URL if applicable
```

## Project Structure

After initialization, DrDuck creates:

```
.drduck/
├── config.yml              # Configuration file
├── templates/              # Custom templates
└── hooks/                  # Git hook scripts

docs/adrs/                  # ADRs (if same-repo storage)
├── README.md              # ADR index
├── 0001-use-adr.md        # First ADR
└── 0002-feature-name.md   # Additional ADRs
```

## ADR Templates

### Nygard Template (Default)

Based on [Michael Nygard's original ADR format](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions), includes:
- Status of the decision
- Context that influences or constrains the decision
- The decision we're proposing or implementing
- Consequences of the decision

### MADR Template

Based on [Markdown Any Decision Records](https://adr.github.io/madr/), includes:
- Context and problem statement
- Decision with rationale  
- Consequences (positive, negative, neutral)
- Alternatives considered

### Simple Template

Lightweight format with:
- Problem description
- Solution overview
- Rationale
- Impact assessment

## Integration with AI Assistants

DrDuck is designed to work with:

- **Claude Code CLI**: Integrates with your Claude coding sessions
- **Cursor**: Works with Cursor's AI-powered development environment

The tool can automatically analyze code changes and help complete ADRs based on development context.

## Git Hooks

Optional git hooks help maintain documentation discipline:

- **Pre-commit**: Validates ADR completeness for staged changes
- **Pre-push**: Ensures significant changes have associated ADRs

Hooks can be bypassed with `git commit --no-verify` when needed.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Roadmap

- [ ] Claude Code CLI integration
- [ ] Cursor integration  
- [ ] Separate repository support
- [ ] Custom template system
- [ ] CI/CD pipeline integration
- [ ] Web-based ADR visualization

---

*Built with ❤️ by [SilverFlin](https://github.com/SilverFlin)*
