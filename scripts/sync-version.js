#!/usr/bin/env node

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

function syncVersionFromGitTag() {
  try {
    // Get the latest git tag
    const latestTag = execSync('git describe --tags --abbrev=0', { encoding: 'utf8' }).trim();
    
    // Remove 'v' prefix if present
    const version = latestTag.startsWith('v') ? latestTag.slice(1) : latestTag;
    
    // Read package.json
    const packagePath = path.join(__dirname, '..', 'package.json');
    const packageJson = JSON.parse(fs.readFileSync(packagePath, 'utf8'));
    
    // Check if version needs updating
    if (packageJson.version === version) {
      console.log(`✅ package.json version (${packageJson.version}) already matches git tag (${latestTag})`);
      return;
    }
    
    // Update version
    packageJson.version = version;
    
    // Write back to package.json
    fs.writeFileSync(packagePath, JSON.stringify(packageJson, null, 2) + '\n');
    
    console.log(`🔄 Updated package.json version from ${packageJson.version} to ${version} (git tag: ${latestTag})`);
    
  } catch (error) {
    console.error('❌ Error syncing version from git tag:', error.message);
    process.exit(1);
  }
}

syncVersionFromGitTag();