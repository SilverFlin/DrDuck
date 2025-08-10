#!/usr/bin/env node

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

function release(versionType = 'patch') {
  try {
    // Ensure we're on main/master branch
    const currentBranch = execSync('git branch --show-current', { encoding: 'utf8' }).trim();
    if (!['main', 'master'].includes(currentBranch)) {
      throw new Error(`Must be on main/master branch. Currently on: ${currentBranch}`);
    }
    
    // Ensure working directory is clean
    const status = execSync('git status --porcelain', { encoding: 'utf8' }).trim();
    if (status) {
      throw new Error('Working directory not clean. Commit your changes first.');
    }
    
    // Read current version
    const packagePath = path.join(__dirname, '..', 'package.json');
    const packageJson = JSON.parse(fs.readFileSync(packagePath, 'utf8'));
    const currentVersion = packageJson.version;
    
    console.log(`📦 Current version: ${currentVersion}`);
    
    // Calculate new version
    const [major, minor, patch] = currentVersion.split('.').map(Number);
    let newVersion;
    
    switch (versionType) {
      case 'major':
        newVersion = `${major + 1}.0.0`;
        break;
      case 'minor':
        newVersion = `${major}.${minor + 1}.0`;
        break;
      case 'patch':
      default:
        newVersion = `${major}.${minor}.${patch + 1}`;
        break;
    }
    
    console.log(`🚀 Releasing version: ${newVersion}`);
    
    // Update package.json
    packageJson.version = newVersion;
    fs.writeFileSync(packagePath, JSON.stringify(packageJson, null, 2) + '\n');
    
    // Commit version bump
    execSync(`git add package.json`);
    execSync(`git commit -m "chore: bump version to ${newVersion}"`);
    
    // Create and push tag
    execSync(`git tag v${newVersion}`);
    execSync(`git push origin main`);
    execSync(`git push origin v${newVersion}`);
    
    console.log(`✅ Tagged and pushed v${newVersion}`);
    
    // Publish to npm
    execSync('npm publish', { stdio: 'inherit' });
    
    console.log(`🎉 Released ${newVersion} successfully!`);
    
  } catch (error) {
    console.error('❌ Release failed:', error.message);
    process.exit(1);
  }
}

// Parse command line arguments
const versionType = process.argv[2] || 'patch';
if (!['major', 'minor', 'patch'].includes(versionType)) {
  console.error('Usage: node scripts/release.js [major|minor|patch]');
  process.exit(1);
}

release(versionType);