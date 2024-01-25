import JSZip from 'jszip';

export default () => new Promise((resolve, reject) => {
  // 生成JSON文件内容
  const jsonData = {
    name: 'John Doe',
    age: 30,
  };

  const jsonContent = JSON.stringify(jsonData, null, 2); // 格式化JSON字符串

  // 生成YAML文件内容
  const yamlData = `
      name: John Doe
      age: 30
    `;
    // 创建JSZip实例
  const zip = new JSZip();

  // 将JSON和YAML文件添加到压缩包
  zip.file('data.json', jsonContent);
  zip.file('data.yaml', yamlData);

  // 生成压缩包
  zip.generateAsync({ type: 'blob' }).then((content) => {
    // 创建下载链接
    const href = URL.createObjectURL(content);
    resolve(href);
  })
    .catch((error) => {
      reject(error);
    });
});
