# Audit

Audits a project's Node packages and/or Ruby gems.

## Installation

```bash
brew install go
```

```bash
go install github.com/friendsoftheweb/audit@latest
```

Add the following line to your `.zshrc` or `.bashrc`:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

## Running

```bash
audit
```

## Building

```bash
go build
```
