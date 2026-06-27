#!/usr/bin/env node
// Thin launcher: exec the downloaded `zebra` binary, forwarding args, stdio,
// and exit code. The binary is fetched into ./bin by install.js at postinstall.
"use strict";

const path = require("path");
const { spawnSync } = require("child_process");

const ext = process.platform === "win32" ? ".exe" : "";
const bin = path.join(__dirname, "bin", `zebra${ext}`);

const res = spawnSync(bin, process.argv.slice(2), { stdio: "inherit" });

if (res.error) {
  if (res.error.code === "ENOENT") {
    console.error("zebra: binary not found. Reinstall with: npm install -g zebra-tui");
  } else {
    console.error(`zebra: ${res.error.message}`);
  }
  process.exit(1);
}

process.exit(res.status ?? 1);
