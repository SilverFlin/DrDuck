#!/usr/bin/env node

const path = require('path');
const fs = require('fs');

console.log('');
console.log('🦆 DrDuck installed successfully!');
console.log('');
console.log('📋 Quick start:');
console.log('   drduck init        # Initialize project');  
console.log('   drduck new -n "my-decision"  # Create ADR');
console.log('   drduck list        # List all ADRs');
console.log('');
console.log('📚 Documentation: https://github.com/SilverFlin/DrDuck');
console.log('');

// Check if binary exists and is executable
const binaryPath = path.join(__dirname, '..', 'bin', process.platform === 'win32' ? 'drduck.exe' : 'drduck');

if (!fs.existsSync(binaryPath)) {
  console.log('⚠️  Warning: DrDuck binary not found. Installation may have failed.');
  console.log('   Try: npm install -g drduck --force');
} else {
  // Ensure binary is executable
  try {
    fs.chmodSync(binaryPath, 0o755);
  } catch (error) {
    // Ignore chmod errors on Windows
  }
}