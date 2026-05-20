const { execSync } = require("child_process");
const { existsSync, renameSync, mkdirSync, rmSync, writeFileSync } = require("fs");
const { join } = require("path");
const https = require("https");

const REPO = "ibrhr/qdoc";

function mapPlatform() {
  const goos = { linux: "linux", darwin: "darwin", win32: "windows" }[process.platform];
  const goarch = { x64: "amd64", arm64: "arm64" }[process.arch];
  if (!goos || !goarch) throw new Error(`Unsupported: ${process.platform}/${process.arch}`);
  return `${goos}_${goarch}`;
}

function fetch(url) {
  return new Promise((resolve, reject) => {
    https.get(url, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        https.get(res.headers.location, (rr) => {
          const c = []; rr.on("data", (d) => c.push(d));
          rr.on("end", () => resolve(Buffer.concat(c)));
          rr.on("error", reject);
        });
        return;
      }
      const c = []; res.on("data", (d) => c.push(d));
      res.on("end", () => resolve(Buffer.concat(c)));
      res.on("error", reject);
    }).on("error", reject);
  });
}

async function main() {
  const pkgDir = join(__dirname);
  const ext = process.platform === "win32" ? ".exe" : "";
  const dest = join(pkgDir, "qdoc_bin" + ext);

  if (existsSync(dest)) return;

  const platform = mapPlatform();
  const base = `https://github.com/${REPO}/releases/latest/download/qdoc_${platform}`;
  const url = process.platform === "win32" ? `${base}.zip` : `${base}.tar.gz`;

  process.stderr.write(`qdoc: downloading ${url}\n`);
  const buf = await fetch(url);

  const tmp = join(pkgDir, ".qdoc_tmp");
  mkdirSync(tmp, { recursive: true });

  try {
    if (process.platform === "win32") {
      writeFileSync(join(tmp, "qdoc.zip"), buf);
      execSync(`tar -xf qdoc.zip`, { cwd: tmp, stdio: "ignore" });
      rmSync(join(tmp, "qdoc.zip"));
    } else {
      writeFileSync(join(tmp, "qdoc.tar.gz"), buf);
      execSync(`tar xzf qdoc.tar.gz`, { cwd: tmp, stdio: "ignore" });
      rmSync(join(tmp, "qdoc.tar.gz"));
    }

    // Find the extracted binary and move it to dest
    const { readdirSync } = require("fs");
    const files = readdirSync(tmp).filter((f) => f.startsWith("qdoc"));
    if (files.length === 0) throw new Error("no binary found in archive");
    renameSync(join(tmp, files[0]), dest);
  } finally {
    rmSync(tmp, { recursive: true, force: true });
  }

  if (process.platform !== "win32") {
    execSync(`chmod +x "${dest}"`, { stdio: "ignore" });
  }
}

main().catch((err) => {
  console.error("qdoc install failed:", err.message);
  process.exit(1);
});
