const { execFileSync } = require("child_process");
const { existsSync } = require("fs");
const { join } = require("path");

const ext = process.platform === "win32" ? ".exe" : "";
const bin = join(__dirname, "..", "qdoc_bin" + ext);

if (!existsSync(bin)) {
  console.error("qdoc: binary not found. Try reinstalling: npm install -g qdoc-agent");
  process.exit(1);
}

try {
  execFileSync(bin, process.argv.slice(2), { stdio: "inherit" });
} catch (err) {
  process.exit(err.status || 1);
}
