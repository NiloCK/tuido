const os = require("os");
const path = require("path");
const fs = require("fs");
const https = require("https");
const tar = require("tar");
const { version } = require("./package.json");

// Map Node's os.platform() to Go's GOOS values
function getPlatform() {
  const platform = os.platform();
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

// Map Node's os.arch() to Go's GOARCH values
function getArch() {
  const arch = os.arch();
  const mappings = {
    x64: "amd64",
    ia32: "386",
    arm: "arm",
    arm64: "arm64",
  };
  return mappings[arch] || arch;
}

function getBinaryInfo() {
  const platform = getPlatform();
  const arch = getArch();

  console.log(`Detecting platform: ${os.platform()} -> ${platform}`);
  console.log(`Detecting architecture: ${os.arch()} -> ${arch}`);

  // Handle special cases like M1 Macs and 32-bit systems
  if (platform === "darwin" && arch === "386") {
    console.error("macOS 32-bit is not supported");
    process.exit(1);
  }

  // Handle ARM variants appropriately
  let armSuffix = "";
  if (arch === "arm" && platform !== "darwin") {
    // Default to ARMv7 for compatibility
    const armVersion = process.env.TUIDO_ARMV || "7";
    armSuffix = `v${armVersion}`;
    console.log(`ARM platform detected, using version: ${armVersion}`);
  }

  let extension = "";
  if (platform === "windows") {
    extension = ".exe";
  }

  // Format: tuido_0.0.14_linux_amd64.tar.gz
  const filename = `tuido_${version}_${platform}_${arch}${armSuffix}.tar.gz`;
  const url = `https://github.com/NiloCK/tuido/releases/download/v${version}/${filename}`;

  console.log(`Binary filename: ${filename}`);
  console.log(`Download URL: ${url}`);

  return {
    url,
    name: `tuido${extension}`,
  };
}

async function downloadAndInstall(url, name) {
  console.log(`\n📥 Downloading and installing binary...`);
  
  const binaryDir = path.join(__dirname, "binary");
  const tarPath = path.join(__dirname, "temp.tar.gz");
  const finalBinaryPath = path.join(binaryDir, name);
  

  
  // Ensure binary directory exists
  if (!fs.existsSync(binaryDir)) {
    fs.mkdirSync(binaryDir, { recursive: true });
  }
  
  return new Promise((resolve, reject) => {
    console.log(`Downloading from: ${url}`);
    
    const downloadFromUrl = (downloadUrl, redirectCount = 0) => {
      if (redirectCount > 5) {
        reject(new Error('Too many redirects'));
        return;
      }
      
      const file = fs.createWriteStream(tarPath);
      
      https.get(downloadUrl, (response) => {
        if (response.statusCode === 302 || response.statusCode === 301) {
          console.log(`Following redirect to: ${response.headers.location}`);
          file.close();
          fs.unlink(tarPath, () => {});
          downloadFromUrl(response.headers.location, redirectCount + 1);
          return;
        }
        
        if (response.statusCode !== 200) {
          file.close();
          fs.unlink(tarPath, () => {});
          reject(new Error(`HTTP ${response.statusCode}: ${response.statusMessage}`));
          return;
        }
      
        console.log(`Download started, content-length: ${response.headers['content-length'] || 'unknown'}`);
        
        response.pipe(file);
        
        file.on('finish', () => {
          file.close();
          console.log(`Download completed, extracting...`);
          
          // Extract the tar.gz file
          tar.extract({
            file: tarPath,
            cwd: binaryDir,
            sync: true
          });
          
          // Clean up the tar file
          fs.unlinkSync(tarPath);
          
          // Make binary executable on Unix systems
          if (process.platform !== 'win32' && fs.existsSync(finalBinaryPath)) {
            fs.chmodSync(finalBinaryPath, 0o755);
          }
          
          // Verify installation
          if (fs.existsSync(finalBinaryPath)) {
            console.log(`✅ Binary successfully installed`);
            resolve();
          } else {
            reject(new Error(`Binary not found after extraction: ${finalBinaryPath}`));
          }
        });
        
        file.on('error', (err) => {
          fs.unlink(tarPath, () => {}); // Clean up on error
          reject(err);
        });
        
      }).on('error', (err) => {
        reject(err);
      });
    };
    
    downloadFromUrl(url);
  });
}

async function install() {
  console.log(`\n=== Tuido Binary Installation ===`);
  console.log(`Package version: ${version}`);
  console.log(`Node.js version: ${process.version}`);
  console.log(`Installation directory: ${__dirname}`);
  
  try {
    const { url, name } = getBinaryInfo();
    
    console.log(`Expected binary name: ${name}`);
    console.log(`Expected install location: ${path.join(__dirname, "binary", name)}`);
    
    console.log(`\nStarting download and installation...`);
    
    // Download and install the binary
    await downloadAndInstall(url, name);
    
    // Verify the binary was actually installed
    const binaryPath = path.join(__dirname, "binary", name);
    
    if (fs.existsSync(binaryPath)) {
      console.log(`✅ Binary verified at: ${binaryPath}`);
      const stats = fs.statSync(binaryPath);
      console.log(`Binary size: ${stats.size} bytes`);
      console.log(`Binary permissions: ${stats.mode.toString(8)}`);
    } else {
      console.error(`❌ Binary not found at expected location: ${binaryPath}`);
      
      // List what files were actually created
      const binaryDir = path.join(__dirname, "binary");
      if (fs.existsSync(binaryDir)) {
        const files = fs.readdirSync(binaryDir);
        console.log(`Files in binary directory: ${files.join(", ")}`);
      } else {
        console.log(`Binary directory does not exist: ${binaryDir}`);
      }
      
      process.exit(1);
    }
    
  } catch (e) {
    console.error(`\n❌ Error installing tuido:`, e);
    
    if (e.message) {
      console.error(`Error message: ${e.message}`);
    }
    
    if (e.code) {
      console.error(`Error code: ${e.code}`);
    }
    
    if (e.stack) {
      console.error(`Stack trace:\n${e.stack}`);
    }
    
    console.error(`\n🔍 Troubleshooting information:`);
    console.error(`1. Check your internet connection`);
    console.error(`2. Verify the release exists: ${getBinaryInfo().url}`);
    console.error(`3. Check if your platform/architecture is supported`);
    console.error(`4. Try manual download from: https://github.com/NiloCK/tuido/releases`);
    console.error(`5. Check npm/yarn proxy settings if behind corporate firewall`);
    
    process.exit(1);
  }
}

// Add unhandled rejection handler
process.on('unhandledRejection', (reason, promise) => {
  console.error('Unhandled Rejection at:', promise, 'reason:', reason);
  process.exit(1);
});

// Add uncaught exception handler
process.on('uncaughtException', (error) => {
  console.error('Uncaught Exception:', error);
  process.exit(1);
});

install();