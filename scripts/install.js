#!/usr/bin/env node

const https = require('https');
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');
const packageJson = require('../package.json');

// Platform mapping
const platformMap = {
  darwin: 'Darwin',
  linux: 'Linux', 
  win32: 'Windows'
};

const archMap = {
  x64: 'x86_64',
  ia32: 'i386',
  arm64: 'arm64'
};

function getPlatformInfo() {
  const platform = process.platform;
  const arch = process.arch;
  
  if (!platformMap[platform]) {
    throw new Error(`Unsupported platform: ${platform}`);
  }
  
  if (!archMap[arch]) {
    throw new Error(`Unsupported architecture: ${arch}`);
  }
  
  const goos = platformMap[platform];
  const goarch = archMap[arch];
  
  // Special case for macOS - use universal binary
  if (platform === 'darwin') {
    return {
      filename: 'drduck_Darwin_all.tar.gz',
      binary: 'drduck'
    };
  }
  
  // Windows binary has .exe extension
  const binaryExt = platform === 'win32' ? '.exe' : '';
  const archiveExt = platform === 'win32' ? 'zip' : 'tar.gz';
  
  return {
    filename: `drduck_${goos}_${goarch}.${archiveExt}`,
    binary: `drduck${binaryExt}`
  };
}

function downloadFile(url, destination) {
  return new Promise((resolve, reject) => {
    console.log(`📥 Downloading: ${url}`);
    
    const file = fs.createWriteStream(destination);
    
    https.get(url, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // Handle redirects
        file.close();
        fs.unlinkSync(destination);
        return downloadFile(response.headers.location, destination).then(resolve).catch(reject);
      }
      
      if (response.statusCode !== 200) {
        file.close();
        fs.unlinkSync(destination);
        return reject(new Error(`Download failed: ${response.statusCode} ${response.statusMessage}`));
      }
      
      response.pipe(file);
      
      file.on('finish', () => {
        file.close();
        resolve();
      });
      
      file.on('error', (err) => {
        file.close();
        fs.unlinkSync(destination);
        reject(err);
      });
    }).on('error', (err) => {
      file.close();
      fs.unlinkSync(destination);
      reject(err);
    });
  });
}

function extractArchive(archivePath, extractPath, binaryName) {
  console.log(`📦 Extracting: ${archivePath}`);
  
  try {
    if (archivePath.endsWith('.tar.gz')) {
      execSync(`tar -xzf "${archivePath}" -C "${extractPath}"`, { stdio: 'inherit' });
    } else if (archivePath.endsWith('.zip')) {
      // Use unzip on Unix-like systems, or fall back to node module
      try {
        execSync(`unzip -q "${archivePath}" -d "${extractPath}"`, { stdio: 'inherit' });
      } catch (error) {
        // Fallback for systems without unzip
        const AdmZip = require('adm-zip');
        const zip = new AdmZip(archivePath);
        zip.extractAllTo(extractPath, true);
      }
    }
    
    // Make binary executable
    const binaryPath = path.join(extractPath, binaryName);
    if (fs.existsSync(binaryPath)) {
      fs.chmodSync(binaryPath, 0o755);
      return binaryPath;
    } else {
      throw new Error(`Binary ${binaryName} not found in extracted archive`);
    }
  } catch (error) {
    throw new Error(`Extraction failed: ${error.message}`);
  }
}

async function installBinary() {
  try {
    console.log('🦆 Installing DrDuck...');
    
    const { filename, binary } = getPlatformInfo();
    const version = `v${packageJson.version}`;
    const downloadUrl = `https://github.com/SilverFlin/DrDuck/releases/download/${version}/${filename}`;
    
    // Create directories
    const binDir = path.join(__dirname, '..', 'bin');
    const tempDir = path.join(__dirname, '..', 'temp');
    
    if (!fs.existsSync(binDir)) {
      fs.mkdirSync(binDir, { recursive: true });
    }
    if (!fs.existsSync(tempDir)) {
      fs.mkdirSync(tempDir, { recursive: true });
    }
    
    // Download archive
    const archivePath = path.join(tempDir, filename);
    await downloadFile(downloadUrl, archivePath);
    
    // Extract binary
    const extractedBinaryPath = extractArchive(archivePath, tempDir, binary);
    
    // Move binary to bin directory
    const finalBinaryPath = path.join(binDir, binary);
    fs.copyFileSync(extractedBinaryPath, finalBinaryPath);
    fs.chmodSync(finalBinaryPath, 0o755);
    
    // Clean up
    fs.rmSync(tempDir, { recursive: true, force: true });
    
    console.log('✅ DrDuck installed successfully!');
    console.log(`📍 Binary location: ${finalBinaryPath}`);
    
    // Test the binary
    try {
      const output = execSync(`"${finalBinaryPath}" --version`, { encoding: 'utf8' });
      console.log(`🔍 Version: ${output.trim()}`);
    } catch (error) {
      console.warn('⚠️  Could not verify installation');
    }
    
  } catch (error) {
    console.error('❌ Installation failed:', error.message);
    console.error('');
    console.error('Alternative installation methods:');
    console.error('1. go install github.com/SilverFlin/DrDuck@latest');
    console.error('2. Download manually from: https://github.com/SilverFlin/DrDuck/releases');
    process.exit(1);
  }
}

// Only run if this script is executed directly
if (require.main === module) {
  installBinary();
}

module.exports = { installBinary };