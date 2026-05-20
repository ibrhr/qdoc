#!/usr/bin/env node
const { execFileSync } = require("child_process");
const { join } = require("path");

const ext = process.platform === "win32" ? ".exe" : "";
const bin = join(__dirname, "..", "qdoc_bin" + ext);

try {
  execFileSync(bin, process.argv.slice(2), { stdio: "inherit" });
} catch (err) {
  process.exit(err.status || 1);
}
