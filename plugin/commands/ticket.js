const fs = require('fs');
const path = require('path');
const https = require('https');
const { spawnSync } = require('child_process');

// 1. Detect platform and arch
const platform = process.platform;
const arch = process.arch;

let osName = '';
if (platform === 'win32') osName = 'windows';
else if (platform === 'darwin') osName = 'darwin';
else if (platform === 'linux') osName = 'linux';
else {
  console.error(`[ticket] Unsupported platform: ${platform}`);
  process.exit(1);
}

let archName = '';
if (arch === 'x64') archName = 'amd64';
else if (arch === 'arm64') archName = 'arm64';
else {
  console.error(`[ticket] Unsupported architecture: ${arch}`);
  process.exit(1);
}

const exeSuffix = osName === 'windows' ? '.exe' : '';
const binaryName = `ticket-${osName}-${archName}${exeSuffix}`;
const binaryPath = path.join(__dirname, binaryName);

// 2. Check if binary already exists
if (fs.existsSync(binaryPath)) {
  runBinary(binaryPath);
} else {
  downloadAndRun();
}

function runBinary(binPath) {
  const args = process.argv.slice(2);
  const result = spawnSync(binPath, args, { stdio: 'inherit' });
  process.exit(result.status ?? 0);
}

function downloadAndRun() {
  console.warn(`[ticket] Binary not found. Downloading ${binaryName} from GitHub Releases...`);
  const url = `https://github.com/deepziyu/ticket-cli-plugin/releases/latest/download/${binaryName}`;
  
  downloadFile(url, binaryPath, (err) => {
    if (err) {
      console.error(`[ticket] Download failed: ${err.message}`);
      // Clean up the empty or corrupted file if it exists
      if (fs.existsSync(binaryPath)) {
        try { fs.unlinkSync(binaryPath); } catch (e) {}
      }
      process.exit(1);
    }
    
    // Set executable permissions on Unix/macOS
    if (osName !== 'windows') {
      try {
        fs.chmodSync(binaryPath, 0o755);
      } catch (chmodErr) {
        console.error(`[ticket] Failed to set permissions: ${chmodErr.message}`);
      }
    }
    
    console.warn(`[ticket] Download complete. Running command...`);
    runBinary(binaryPath);
  });
}

function downloadFile(url, dest, callback) {
  const file = fs.createWriteStream(dest);
  
  const request = https.get(url, (response) => {
    // Handle redirects
    if (response.statusCode >= 300 && response.statusCode < 400 && response.headers.location) {
      file.close();
      fs.unlink(dest, () => {
        downloadFile(response.headers.location, dest, callback);
      });
      return;
    }
    
    if (response.statusCode !== 200) {
      file.close();
      fs.unlink(dest, () => {
        callback(new Error(`Server responded with status code ${response.statusCode}`));
      });
      return;
    }
    
    response.pipe(file);
    
    file.on('finish', () => {
      file.close(callback);
    });
  });
  
  request.on('error', (err) => {
    file.close();
    fs.unlink(dest, () => {
      callback(err);
    });
  });
}
