#!/bin/bash

# Generate shell completions for DrDuck
set -e

# Build the binary first
echo "Building DrDuck..."
go build -o drduck main.go

# Create completions directory
mkdir -p completions

# Generate completions for different shells
echo "Generating shell completions..."

# Bash completion
echo "Generating bash completion..."
./drduck completion bash > completions/drduck.bash

# Zsh completion  
echo "Generating zsh completion..."
./drduck completion zsh > completions/drduck.zsh

# Fish completion
echo "Generating fish completion..."
./drduck completion fish > completions/drduck.fish

# PowerShell completion
echo "Generating PowerShell completion..."
./drduck completion powershell > completions/drduck.ps1

echo "Completions generated in ./completions/"
echo "- bash: completions/drduck.bash"
echo "- zsh: completions/drduck.zsh" 
echo "- fish: completions/drduck.fish"
echo "- powershell: completions/drduck.ps1"

# Clean up binary
rm -f drduck