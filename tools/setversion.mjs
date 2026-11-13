import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const version = (process.argv[2] || "").trim().replace(/^[vV]/, "");
if (!/^\d+\.\d+\.\d+(?:[-.+][0-9A-Za-z.-]+)?$/.test(version)) {
  console.error(`Invalid version: ${process.argv[2] || "(empty)"}`);
  process.exit(1);
}

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const windowsVersion = version.split(/[-+]/)[0];

function read(relative) {
  return fs.readFileSync(path.join(root, relative), "utf8");
}

function write(relative, content) {
  fs.writeFileSync(path.join(root, relative), content);
}

// build/config.yml 只替换 info 下的 version，保留顶层 version: '3'。
const configPath = "build/config.yml";
const config = read(configPath);
const infoVersion = /^(\s{2}version:\s*)"[^"]*"/m;
if (!infoVersion.test(config)) {
  throw new Error(`未找到 build/config.yml 的 info.version`);
}
write(configPath, config.replace(infoVersion, `$1"${version}"`));

// Windows 资源信息（exe 属性中的文件/产品版本）。
const infoPath = "build/windows/info.json";
const info = JSON.parse(read(infoPath));
info.fixed.file_version = windowsVersion;
info.fixed.product_version = windowsVersion;
for (const locale of Object.values(info.info || {})) {
  locale.FileVersion = windowsVersion;
  locale.ProductVersion = windowsVersion;
}
write(infoPath, `${JSON.stringify(info, null, "\t")}\n`);

// 应用清单与 NSIS 安装器脚本中的显示版本。
const manifestPath = "build/windows/wails.exe.manifest";
const manifest = read(manifestPath);
const manifestVersion = /(name="com\.ba4rhh\.cracexam"\s+version=")[^"]*(")/;
if (!manifestVersion.test(manifest)) {
  throw new Error(`未找到 ${manifestPath} 的 assemblyIdentity.version`);
}
write(manifestPath, manifest.replace(manifestVersion, `$1${windowsVersion}$2`));

const nsisPath = "build/windows/nsis/wails_tools.nsh";
const nsis = read(nsisPath);
const nsisVersion = /(!define INFO_PRODUCTVERSION\s+")[^"]*(")/;
if (!nsisVersion.test(nsis)) {
  throw new Error(`未找到 ${nsisPath} 的 INFO_PRODUCTVERSION`);
}
write(nsisPath, nsis.replace(nsisVersion, `$1${windowsVersion}$2`));

// 前端包版本，保持 package.json 与 lockfile 一致。
const packagePath = "frontend/package.json";
const pkg = JSON.parse(read(packagePath));
pkg.version = version;
write(packagePath, `${JSON.stringify(pkg, null, 2)}\n`);

const lockPath = "frontend/package-lock.json";
const lock = JSON.parse(read(lockPath));
lock.version = version;
if (lock.packages?.[""]) lock.packages[""].version = version;
write(lockPath, `${JSON.stringify(lock, null, 2)}\n`);

console.log(`Version metadata updated to ${version}`);
