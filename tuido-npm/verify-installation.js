#!/usr/bin/env node

const fs = require("fs");
const path = require("path");
const os = require("os");
const https = require("https");

console.log("=== Tuido Installation Verification ===\n");

const packageDir = __dirname;
const binaryDir = path.join(packageDir, "binary");
const platform = os.platform();
const arch = os.arch();
const expectedBinaryName = platform === "win32" ? "tuido.exe" : "tuido";
const expectedBinaryPath = path.join(binaryDir, expectedBinaryName);

console.log(`Platform: ${platform}`);
console.log(`Architecture: ${arch}`);
console.log(`Package directory: ${packageDir}`);
console.log(`Expected binary path: ${expectedBinaryPath}\n`);

// Step 1: Check if package.json exists and has correct version
try {
  const packageJson = JSON.parse(fs.readFileSync(path.join(packageDir, "package.json"), "utf8"));
  console.log(`✅ Package version: ${packageJson.version}`);
} catch (e) {
  console.log(`❌ Cannot read package.json: ${e.message}`);
  process.exit(1);
}

// Step 2: Check if binary directory exists
if (fs.existsSync(binaryDir)) {
  console.log(`✅ Binary directory exists`);
  
  const binaryContents = fs.readdirSync(binaryDir);
  console.log(`   Contents: ${binaryContents.join(", ")}`);
  
  // Step 3: Check if expected binary exists
  if (fs.existsSync(expectedBinaryPath)) {
    console.log(`✅ Binary file exists: ${expectedBinaryName}`);
    
    const stats = fs.statSync(expectedBinaryPath);
    console.log(`   Size: ${stats.size} bytes`);
    console.log(`   Modified: ${stats.mtime}`);
    
    // Check permissions
    const isExecutable = !!(stats.mode & parseInt('111', 8));
    if (isExecutable || platform === "win32") {
      console.log(`✅ Binary is executable`);
    } else {
      console.log(`❌ Binary is not executable (permissions: ${stats.mode.toString(8)})`);
    }
    
  } else {
    console.log(`❌ Expected binary not found: ${expectedBinaryName}`);
    console.log(`   Available files: ${binaryContents.join(", ")}`);
  }
  
} else {
  console.log(`❌ Binary directory does not exist: ${binaryDir}`);
  
  // Check what files do exist in package directory
  if (fs.existsSync(packageDir)) {
    const packageContents = fs.readdirSync(packageDir);
    console.log(`   Package contents: ${packageContents.join(", ")}`);
  }
}

// Step 4: Check if install.js exists
const installScript = path.join(packageDir, "install.js");
if (fs.existsSync(installScript)) {
  console.log(`✅ Install script exists`);
} else {
  console.log(`❌ Install script missing: install.js`);
}

// Step 5: Check if CLI wrapper exists
const cliScript = path.join(packageDir, "bin", "tuido-cli.js");
if (fs.existsSync(cliScript)) {
  console.log(`✅ CLI wrapper exists`);
} else {
  console.log(`❌ CLI wrapper missing: bin/tuido-cli.js`);
}

// Step 6: Test if GitHub release URL is accessible
function getPlatformName() {
  const mappings = {
    win32: "windows",
    darwin: "darwin",
    linux: "linux",
    freebsd: "freebsd",
    openbsd: "openbsd",
    netbsd: "netbsd",
  };
  return mappings[platform] || platform;
}

function getArchName() {
  const mappings = {
    x64: "amd64",
    ia32: "386",
    arm: "arm",
    arm64: "arm64",
  };
  return mappings[arch] || arch;
}

const packageJson = JSON.parse(fs.readFileSync(path.join(packageDir, "package.json"), "utf8"));
const version = packageJson.version;
const platformName = getPlatformName();
const archName = getArchName();
const filename = `tuido_${version}_${platformName}_${archName}.tar.gz`;
const releaseUrl = `https://github.com/NiloCK/tuido/releases/download/v${version}/${filename}`;

console.log(`\nChecking GitHub release availability...`);
console.log(`URL: ${releaseUrl}`);

https.request(releaseUrl, { method: 'HEAD' }, (res) => {
  if (res.statusCode === 200) {
    console.log(`✅ GitHub release is accessible (HTTP ${res.statusCode})`);
    console.log(`   Content-Length: ${res.headers['content-length'] || 'unknown'}`);
    console.log(`   Content-Type: ${res.headers['content-type'] || 'unknown'}`);
  } else {
    console.log(`❌ GitHub release not accessible (HTTP ${res.statusCode})`);
  }
}).on('error', (err) => {
  console.log(`❌ Network error checking GitHub release: ${err.message}`);
}).end();

// Step 7: Summary and recommendations
setTimeout(() => {
  console.log(`\n=== Summary ===`);
  
  if (fs.existsSync(expectedBinaryPath)) {
    console.log(`✅ Installation appears successful`);
    console.log(`\nTo test the binary:`);
    console.log(`   node bin/tuido-cli.js --help`);
  } else {
    console.log(`❌ Installation appears incomplete`);
    console.log(`\nTo fix:`);
    console.log(`   1. Run: node install.js`);
    console.log(`   2. Check the output for errors`);
    console.log(`   3. If that fails, run: node test-install.js`);
    console.log(`   4. As a last resort, manually download the binary from:`);
    console.log(`      ${releaseUrl}`);
  }
}, 2000);