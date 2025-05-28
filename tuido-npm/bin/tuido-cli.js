#!/usr/bin/env node

const { spawn } = require("child_process");
const fs = require("fs");
const os = require("os");
const path = require("path");

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

function getBinaryPath() {
  const platform = getPlatform();
  let extension = "";

  if (platform === "windows") {
    extension = ".exe";
  }

  // binary-install stores binaries in: node_modules/<package>/binary/<name>
  return path.join(__dirname, "..", "binary", `tuido${extension}`);
}

function findBinaryPath() {
  // Try the standard binary-install location first
  let binaryPath = getBinaryPath();
  
  if (fs.existsSync(binaryPath)) {
    return binaryPath;
  }

  // Fallback: check if binary-install used a different location
  const alternativePaths = [
    path.join(__dirname, "..", "node_modules", ".bin", "tuido"),
    path.join(__dirname, "..", "tuido"),
    path.join(__dirname, "..", "bin", "tuido"),
  ];

  for (const altPath of alternativePaths) {
    const platform = getPlatform();
    const withExt = platform === "windows" ? `${altPath}.exe` : altPath;
    if (fs.existsSync(withExt)) {
      return withExt;
    }
    if (fs.existsSync(altPath)) {
      return altPath;
    }
  }

  return null;
}

try {
  const binaryPath = findBinaryPath();
  
  if (!binaryPath) {
    console.error("Error: tuido binary not found. Please try reinstalling the package.");
    console.error("Expected location:", getBinaryPath());
    process.exit(1);
  }

  // Check if binary is executable
  try {
    fs.accessSync(binaryPath, fs.constants.F_OK | fs.constants.X_OK);
  } catch (err) {
    console.error(`Error: tuido binary found but not executable: ${binaryPath}`);
    console.error("Try reinstalling the package or check file permissions.");
    process.exit(1);
  }

  // Execute the binary
  const child = spawn(binaryPath, process.argv.slice(2), {
    stdio: "inherit",
    windowsHide: false,
  });

  child.on("error", (err) => {
    console.error("Error executing tuido:", err.message);
    process.exit(1);
  });

  child.on("close", (code) => {
    process.exit(code || 0);
  });

  // Handle signals
  process.on("SIGINT", () => {
    child.kill("SIGINT");
  });

  process.on("SIGTERM", () => {
    child.kill("SIGTERM");
  });

} catch (e) {
  console.error("Error running tuido:", e.message);
  process.exit(1);
}