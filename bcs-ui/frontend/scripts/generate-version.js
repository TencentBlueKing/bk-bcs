const crypto = require('crypto');
const fs = require('fs')

// 生成一个随机的Buffer
const randomBuffer = crypto.randomBytes(256);

// 创建一个哈希算法实例，比如SHA-256
const hash = crypto.createHash('sha256');

// 使用随机Buffer更新哈希值
hash.update(randomBuffer);

// 计算哈希值，输出为16进制字符串
const hashDigest = hash.digest('hex');

fs.writeFileSync('./static/static_version.txt',hashDigest)

