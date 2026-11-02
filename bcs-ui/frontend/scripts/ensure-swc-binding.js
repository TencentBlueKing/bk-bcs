// npm 会静默跳过下载失败的 optionalDependencies，@swc/core 的平台二进制包缺失时构建会报 Failed to load native binding
const { execSync } = require('child_process');
const path = require('path');

const canLoad = () => {
  try {
    execSync('node -e "require(\'@swc/core\')"', { stdio: 'ignore' });
    return true;
  } catch (e) {
    return false;
  }
};

const isMusl = () => {
  const report = process.report && process.report.getReport();
  return !(report && report.header && report.header.glibcVersionRuntime);
};

const platformPackage = () => {
  const { platform, arch } = process;
  if (platform === 'linux') return `@swc/core-linux-${arch}-${isMusl() ? 'musl' : 'gnu'}`;
  if (platform === 'win32') return `@swc/core-win32-${arch}-msvc`;
  return `@swc/core-${platform}-${arch}`;
};

if (!canLoad()) {
  const { version } = require(path.resolve(__dirname, '../node_modules/@swc/core/package.json'));
  const pkg = `${platformPackage()}@${version}`;
  console.warn(`@swc/core native binding missing, installing ${pkg}`);
  execSync(`npm install --no-save --legacy-peer-deps ${pkg}`, { stdio: 'inherit' });
  if (!canLoad()) {
    console.error(`@swc/core native binding still unavailable after installing ${pkg}`);
    process.exit(1);
  }
}
