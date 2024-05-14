import JSZip from 'jszip';

export default () =>
  new Promise((resolve, reject) => {
    // 生成JSON文件内容
    const jsonData = {
      string_demo: { kv_type: 'string', value: 'blueking' },
      number_demo: { kv_type: 'number', value: 12345 },
      text_demo: { kv_type: 'text', value: 'line 1\nline 2' },
      json_demo: {
        kv_type: 'json',
        value: '{"name": "John Doe", "age": 30, "city": "New York", "hobbies": ["reading", "travelling", "sports"]}',
      },
      xml_demo: {
        kv_type: 'xml',
        value:
          '<person>\n  <name>John Doe</name>\n  <age>30</age>\n  <city>New York</city>\n  <hobbies>\n    <hobby>reading</hobby>\n    <hobby>travelling</hobby>\n    <hobby>sports</hobby>\n  </hobbies>\n</person>',
      },
      yaml_demo: {
        kv_type: 'yaml',
        value: 'name: John Doe\nage: 30\ncity: New York\nhobbies:\n  - reading\n  - travelling\n  - sports',
      },
    };

    const jsonContent = JSON.stringify(jsonData, null, 2); // 格式化JSON字符串

    // 生成YAML文件内容
    const yamlData = `string_demo:
  kv_type: string
  value: "blueking"

number_demo:
  kv_type: number
  value: 12345

text_demo:
  kv_type: text
  value: |-
    line 1
    line 2

json_demo:
  kv_type: json
  value: |-
    {
      "name": "John Doe",
      "age": 30,
      "city": "New York",
      "hobbies": ["reading", "travelling", "sports"]
    }

xml_demo:
  kv_type: xml
  value: |-
    <person>
    <name>John Doe</name>
    <age>30</age>
    <city>New York</city>
    <hobbies>
      <hobby>reading</hobby>
      <hobby>travelling</hobby>
      <hobby>sports</hobby>
    </hobbies>
    </person>

yaml_demo:
  kv_type: yaml
  value: |-
    name: John Doe
    age: 30
    city: New York
    hobbies:
      - reading
      - travelling
      - sports
    `;
    // 创建JSZip实例
    const zip = new JSZip();

    // 将JSON和YAML文件添加到压缩包
    zip.file('json_demo.json', jsonContent);
    zip.file('yaml_demo.yaml', yamlData);

    // 生成压缩包
    zip
      .generateAsync({ type: 'blob' })
      .then((content) => {
        // 创建下载链接
        const href = URL.createObjectURL(content);
        resolve(href);
      })
      .catch((error) => {
        reject(error);
      });
  });
