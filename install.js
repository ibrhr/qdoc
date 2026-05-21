const { execSync } = require("child_process");
const { existsSync, renameSync, mkdirSync, rmSync, writeFileSync } = require("fs");
const { join } = require("path");
const https = require("https");

const REPO = "ibrhr/qdoc";
const VERSION = require("./package.json").version;

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

  if (existsSync(dest)) {
    try {
      const out = execSync(`"${dest}" --version`, { encoding: "utf-8" });
      const match = out.match(/\d+\.\d+\.\d+/);
      if (match && match[0] === VERSION) return;
    } catch (_) {}
  }

  const platform = mapPlatform();
  const tag = `v${VERSION}`;
  const base = `https://github.com/${REPO}/releases/download/${tag}/qdoc_${platform}`;
  const url = process.platform === "win32" ? `${base}.zip` : `${base}.tar.gz`;

  process.stderr.write(`qdoc: downloading ${url}\n`);
  const buf = await fetch(url);

  const tmp = join(pkgDir, ".qdoc_tmp");
  mkdirSync(tmp, { recursive: true });

  try {
    if (process.platform === "win32") {
      writeFileSync(join(tmp, "qdoc.zip"), buf);
      execSync(`tar -xf qdoc.zip`, { cwd: tmp, stdio: "ignore" });
    } else {
      writeFileSync(join(tmp, "qdoc.tar.gz"), buf);
      execSync(`tar xzf qdoc.tar.gz`, { cwd: tmp, stdio: "ignore" });
    }

    const src = join(tmp, "qdoc" + ext);
    if (!existsSync(src)) throw new Error("binary not found in archive");
    renameSync(src, dest);
  } finally {
    rmSync(tmp, { recursive: true, force: true });
  }

  if (process.platform !== "win32") {
    execSync(`chmod +x "${dest}"`, { stdio: "ignore" });
  }
}

main().catch((err) => {
  console.error(`qdoc install failed (v${VERSION}):`, err.message);
  process.exit(1);
});
