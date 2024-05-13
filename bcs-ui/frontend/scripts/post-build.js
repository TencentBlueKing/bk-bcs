const fs = require('fs')

// 生成.gitkeep，golang embed lint 需要
fs.writeFileSync('./dist/.gitkeep', "")
