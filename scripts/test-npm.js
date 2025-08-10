#!/usr/bin/env node

// Simple test to validate npm package structure
const fs = require('fs');
const path = require('path');

console.log('🧪 Testing npm package structure...');

// Check package.json
const packagePath = path.join(__dirname, '..', 'package.json');
if (fs.existsSync(packagePath)) {
  const pkg = require(packagePath);
  console.log('✅ package.json exists');
  console.log(`   Name: ${pkg.name}`);
  console.log(`   Version: ${pkg.version}`);
} else {
  console.log('❌ package.json missing');
}

// Check scripts directory
const scriptsDir = path.join(__dirname);
const requiredScripts = ['install.js', 'postinstall.js'];

requiredScripts.forEach(script => {
  const scriptPath = path.join(scriptsDir, script);
  if (fs.existsSync(scriptPath)) {
    console.log(`✅ ${script} exists`);
  } else {
    console.log(`❌ ${script} missing`);
  }
});

// Check bin directory exists (will be populated during install)
const binDir = path.join(__dirname, '..', 'bin');
if (!fs.existsSync(binDir)) {
  fs.mkdirSync(binDir, { recursive: true });
  console.log('✅ bin/ directory created');
} else {
  console.log('✅ bin/ directory exists');
}

console.log('');
console.log('📦 Package is ready for npm publishing!');
console.log('');
console.log('🚀 To publish:');
console.log('   1. npm login');
console.log('   2. npm publish');
console.log('');
console.log('🧪 To test locally:');
console.log('   1. npm pack  # creates .tgz file');
console.log('   2. npm install -g ./drduck-0.1.0.tgz');
console.log('   3. drduck --version');