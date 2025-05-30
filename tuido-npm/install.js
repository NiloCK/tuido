const { Binary } = require("binary-install");
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

function getBinary() {
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

async function manualDownload(url, name) {
  console.log(`\n📥 Attempting manual download as fallback...`);
  
  const binaryDir = path.join(__dirname, "binary");
  const tarPath = path.join(__dirname, "temp.tar.gz");
  const finalBinaryPath = path.join(binaryDir, name);
  
  // Ensure binary directory exists
  if (!fs.existsSync(binaryDir)) {
    fs.mkdirSync(binaryDir, { recursive: true });
    console.log(`Created binary directory: ${binaryDir}`);
  }
  
  return new Promise((resolve, reject) => {
    console.log(`Downloading from: ${url}`);
    
    const file = fs.createWriteStream(tarPath);
    
    https.get(url, (response) => {
      if (response.statusCode !== 200) {
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
        
        console.log(`Extraction completed`);
        
        // Clean up the tar file
        fs.unlinkSync(tarPath);
        
        // Make binary executable on Unix systems
        if (process.platform !== 'win32') {
          fs.chmodSync(finalBinaryPath, 0o755);
          console.log(`Made binary executable`);
        }
        
        resolve();
      });
      
      file.on('error', (err) => {
        fs.unlink(tarPath, () => {}); // Clean up on error
        reject(err);
      });
      
    }).on('error', (err) => {
      reject(err);
    });
  });
}

async function install() {
  console.log(`\n=== Tuido Binary Installation ===`);
  console.log(`Package version: ${version}`);
  console.log(`Node.js version: ${process.version}`);
  console.log(`Installation directory: ${__dirname}`);
  
  try {
    const { url, name } = getBinary();
    
    console.log(`\nCreating Binary instance...`);
    const binary = new Binary(name, url);
    
    console.log(`Binary object created successfully`);
    console.log(`Expected binary name: ${name}`);
    console.log(`Expected install location: ${path.join(__dirname, "binary", name)}`);
    
    console.log(`\nStarting download and installation...`);
    
    let installSuccess = false;
    
    // Try binary-install first
    try {
      // Wrap the install call to catch any errors
      const installPromise = new Promise((resolve, reject) => {
        try {
          const result = binary.install();
          
          // Handle both promise and callback style returns
          if (result && typeof result.then === 'function') {
            result.then(resolve).catch(reject);
          } else {
            // Assume synchronous success
            resolve(result);
          }
        } catch (error) {
          reject(error);
        }
      });
      
      await installPromise;
      installSuccess = true;
      console.log(`\n✅ Successfully installed tuido binary using binary-install: ${name}`);
      
    } catch (binaryInstallError) {
      console.log(`\n⚠️  binary-install failed: ${binaryInstallError.message}`);
      console.log(`Trying manual download fallback...`);
      
      try {
        await manualDownload(url, name);
        installSuccess = true;
        console.log(`\n✅ Successfully installed tuido binary using manual download: ${name}`);
      } catch (manualError) {
        console.error(`\n❌ Manual download also failed: ${manualError.message}`);
        throw manualError;
      }
    }
    
    // Verify the binary was actually installed
    const binaryPath = path.join(__dirname, "binary", name);
    const fs = require("fs");
    
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
    console.error(`2. Verify the release exists: ${getBinary().url}`);
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