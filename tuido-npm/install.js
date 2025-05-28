const { Binary } = require("binary-install");
const os = require("os");
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
  }

  let extension = "";
  if (platform === "windows") {
    extension = ".exe";
  }

  // Format: tuido_0.0.10_linux_amd64.tar.gz
  const filename = `tuido_${version}_${platform}_${arch}${armSuffix}.tar.gz`;
  const url = `https://github.com/NiloCK/tuido/releases/download/v${version}/${filename}`;

  return {
    url,
    name: `tuido${extension}`,
  };
}

function install() {
  try {
    const { url, name } = getBinary();
    console.log(`Installing tuido for ${os.platform()}-${os.arch()}`);
    console.log(`Downloading tuido binary from ${url}`);
    
    const binary = new Binary(name, url);
    binary.install();
    console.log(`Successfully installed tuido binary: ${name}`);
  } catch (e) {
    console.error("Error installing tuido:", e.message || e);
    console.error("Please check:");
    console.error("1. Your internet connection");
    console.error("2. That the release exists on GitHub");
    console.error("3. Your platform/architecture is supported");
    console.error(`Attempted URL: ${getBinary().url}`);
    process.exit(1);
  }
}

install();
