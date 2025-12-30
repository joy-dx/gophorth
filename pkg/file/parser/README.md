parser


A small, production-hardened Go library to parse app release filenames using reverse templates compiled to regular expressions with named capture groups.

You define templates with placeholders like {name}, {version}, {os}, {arch}, optional fields {field?}, and optional segments [...]?. The parser compiles these into efficient regular expressions, applies field-specific normalization (e.g., x86_64 -> amd64, mac -> darwin), and returns extracted fields.

This is ideal for parsing common release artifact names, for example:


- UV-x86_64-unknown-linux-gnu.tar.gz

- UV-aarch64-apple-darwin.tar.gz

- cpython-3.10.16+20250212-aarch64-apple-darwin-pgo+lto-full.tar.zst

- hugo_0.151.1_linux-amd64.tar.gz

- golangci-lint-1.64.6-darwin-arm64.tar.gz

- frankenphp-linux-x86_64

- docker-credential-secretservice-v0.9.3.linux-amd64

Highlights


- Reverse template DSL to quickly write filename parsers.

- Named capture groups and field-level normalization.

- Bounded regex patterns to avoid catastrophic backtracking.

- Input validation and timeout protection to harden parsing.

- Thread-safe with read/write locking and a “sealed” state.

- Builder API for custom configurations.

- Golden-table tests and benchmarks.

Installation


- Go 1.20+ recommended.

- Add to your module and import as package parser (this repo).

Quick Start


1. Create a release parser with sensible defaults

These include common patterns for version, os, arch, and normalizers.