const { execSync } = require("child_process");
const { existsSync, renameSync, mkdirSync, rmSync, writeFileSync } = require("fs");
const { join } = require("path");
const crypto = require("crypto");
const http = require("http");
const https = require("https");

const REPO = "ibrhr/qdoc";
const VERSION = require("./package.json").version;

const MAX_RETRIES = 3;
const RETRY_DELAY_MS = 2000;

function mapPlatform() {
  const goos = { linux: "linux", darwin: "darwin", win32: "windows" }[process.platform];
  const goarch = { x64: "amd64", arm64: "arm64" }[process.arch];
  if (!goos || !goarch) throw new Error(`Unsupported: ${process.platform}/${process.arch}`);
  return `${goos}_${goarch}`;
}

function fetchWithRetry(url, redirects = 0) {
  return new Promise((resolve, reject) => {
    const mod = url.startsWith("https:") ? https : http;

    const req = mod.get(url, { headers: { "User-Agent": "qdoc-installer" } }, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        if (redirects > 5) return reject(new Error("too many redirects"));
        return resolve(fetchWithRetry(res.headers.location, redirects + 1));
      }
      if (res.statusCode !== 200) {
        res.resume();
        return reject(new Error(`HTTP ${res.statusCode}`));
      }
      const chunks = [];
      res.on("data", (d) => chunks.push(d));
      res.on("end", () => resolve(Buffer.concat(chunks)));
      res.on("error", reject);
    });
    req.on("error", reject);
    req.setTimeout(30000, () => { req.destroy(); reject(new Error("request timeout")); });
  });
}

async function fetch(url) {
  let lastErr;
  for (let attempt = 1; attempt <= MAX_RETRIES; attempt++) {
    try {
      return await fetchWithRetry(url);
    } catch (err) {
      lastErr = err;
      if (attempt < MAX_RETRIES) {
        const delay = RETRY_DELAY_MS * Math.pow(2, attempt - 1) * (0.7 + Math.random() * 0.6);
        process.stderr.write(`qdoc: download failed (attempt ${attempt}/${MAX_RETRIES}): ${err.message}, retrying in ${(delay / 1000).toFixed(1)}s\n`);
        await new Promise((r) => setTimeout(r, delay));
      }
    }
  }
  throw lastErr;
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

  const archiveExt = process.platform === "win32" ? "zip" : "tar.gz";
  const archiveName = `qdoc_${platform}.${archiveExt}`;
  process.stderr.write(`qdoc: verifying checksum\n`);
  const checksumUrl = `https://github.com/${REPO}/releases/download/${tag}/checksums.txt`;
  const checksumBuf = await fetch(checksumUrl);
  const checksums = checksumBuf.toString("utf-8");
  const line = checksums.split("\n").find(l => l.includes(archiveName));
  if (!line) throw new Error(`checksum entry not found for ${archiveName}`);
  const expected = line.split(/\s+/)[0];
  const actual = crypto.createHash("sha256").update(buf).digest("hex");
  if (expected !== actual) throw new Error(`checksum mismatch`);

  const tmp = join(pkgDir, ".qdoc_tmp");
  mkdirSync(tmp, { recursive: true });

  try {
    if (process.platform === "win32") {
      writeFileSync(join(tmp, "qdoc.zip"), buf);
      try {
        execSync(`powershell -Command "Expand-Archive -Path '${join(tmp, "qdoc.zip")}' -DestinationPath '${tmp}' -Force"`, { stdio: "ignore" });
      } catch (_) {
        execSync(`tar -xf qdoc.zip`, { cwd: tmp, stdio: "ignore" });
      }
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

  const out = execSync(`"${dest}" --version`, { encoding: "utf-8" }).trim();
  process.stderr.write(`qdoc: installed ${out}\n`);
}

main().catch((err) => {
  if (process.env.CF_PAGES || process.env.CLOUDFLARE_PAGES || (err.message && err.message.includes("404"))) {
    console.warn(`qdoc install notice (v${VERSION}): ${err.message} — skipping binary download`);
    process.exit(0);
  }
  console.error(`qdoc install failed (v${VERSION}):`, err.message);
  process.exit(1);
});
