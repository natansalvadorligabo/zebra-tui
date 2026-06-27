// Postinstall: download the platform-specific `zebra` binary from the matching
// GitHub Release and place it in ./bin. The asset names here must match the
// GoReleaser config (.goreleaser.yaml) exactly: zebra-<os>-<arch>[.exe].
"use strict";

const fs = require("fs");
const path = require("path");
const crypto = require("crypto");

const REPO = "natansalvadorligabo/zebra-tui";
const { version } = require("./package.json");

// node platform/arch -> GoReleaser goos/goarch.
const OS = { win32: "windows", darwin: "darwin", linux: "linux" };
const ARCH = { x64: "amd64", arm64: "arm64" };

async function fetchBuffer(url, what) {
  const res = await fetch(url, { redirect: "follow" });
  if (!res.ok) {
    throw new Error(`failed to download ${what} (${res.status} ${res.statusText})\n  ${url}`);
  }
  return Buffer.from(await res.arrayBuffer());
}

// Parse a GoReleaser checksums.txt ("<sha256>  <filename>" per line) and return
// the expected hash for assetName, or null if absent.
function expectedSha(checksums, assetName) {
  for (const line of checksums.toString("utf8").split("\n")) {
    const [sha, name] = line.trim().split(/\s+/);
    if (name === assetName) return sha;
  }
  return null;
}

async function main() {
  const goos = OS[process.platform];
  const goarch = ARCH[process.arch];
  if (!goos || !goarch) {
    throw new Error(
      `unsupported platform ${process.platform}/${process.arch}. ` +
        `Install from source instead: go install github.com/${REPO}@latest`
    );
  }

  const ext = goos === "windows" ? ".exe" : "";
  const assetName = `zebra-${goos}-${goarch}${ext}`;
  const tag = `v${version}`;
  const base = `https://github.com/${REPO}/releases/download/${tag}`;

  console.log(`zebra-tui: downloading ${assetName} (${tag})`);
  const binary = await fetchBuffer(`${base}/${assetName}`, assetName);

  // Verify the checksum when checksums.txt is available; skip gracefully if not.
  try {
    const checksums = await fetchBuffer(`${base}/checksums.txt`, "checksums.txt");
    const want = expectedSha(checksums, assetName);
    if (want) {
      const got = crypto.createHash("sha256").update(binary).digest("hex");
      if (got !== want) {
        throw new Error(`checksum mismatch for ${assetName}\n  expected ${want}\n  got      ${got}`);
      }
      console.log("zebra-tui: checksum verified");
    }
  } catch (err) {
    if (/checksum mismatch/.test(err.message)) throw err;
    console.warn(`zebra-tui: skipping checksum verification (${err.message})`);
  }

  const binDir = path.join(__dirname, "bin");
  fs.mkdirSync(binDir, { recursive: true });
  const dest = path.join(binDir, `zebra${ext}`);
  fs.writeFileSync(dest, binary, { mode: 0o755 });
  console.log(`zebra-tui: installed ${dest}`);
}

main().catch((err) => {
  console.error(`zebra-tui: install failed: ${err.message}`);
  process.exit(1);
});
