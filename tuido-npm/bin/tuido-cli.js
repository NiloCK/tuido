#!/usr/bin/env node

const { Binary } = require("binary-install");
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

function getBinaryPath() {
  const platform = getPlatform();
  let extension = "";

  if (platform === "windows") {
    extension = ".exe";
  }

  return path.join(
    __dirname,
    "..",
    "node_modules",
    ".bin",
    `tuido${extension}`,
  );
}

try {
  const binaryPath = getBinaryPath();
  const binary = new Binary("tuido", null, { installDirectory: path.dirname(binaryPath) });
  binary.run(process.argv.slice(2));
} catch (e) {
  console.error("Error running tuido:", e);
  process.exit(1);
}
