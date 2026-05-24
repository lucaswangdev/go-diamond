# go-diamond Multi-Language Client Examples

go-diamond 是一个分布式配置中心，提供 HTTP REST API，可以被任何语言的项目使用。

## API 概述

| 功能 | HTTP Method | URL Pattern |
|------|-------------|-------------|
| 获取配置 | GET | `/api/v1/configs/{namespace}/{group}/{dataId}` |
| 监听配置变更 | GET | `/api/v1/watch/{namespace}/{group}/{dataId}?md5=xxx&timeout=30` |
| 批量监听 | POST | `/api/v1/watch/batch` |
| 创建配置 | POST | `/admin/v1/configs` (需认证) |
| 更新配置 | PUT | `/admin/v1/configs/{namespace}/{group}/{dataId}` (需认证) |
| 删除配置 | DELETE | `/admin/v1/configs/{namespace}/{group}/{dataId}` (需认证) |

## 返回格式

```json
{
  "code": 0,
  "data": {
    "dataId": "app.json",
    "group": "DEFAULT_GROUP",
    "namespace": "default",
    "content": "{\"key\": \"value\"}",
    "contentMd5": "abc123...",
    "version": 1,
    "format": "json",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
}
```

## Python 示例

使用标准库（无需安装额外依赖）：

```bash
python3 python_example_simple.py
```

使用 requests 库（更丰富功能）：

```bash
pip install requests
python3 python_example.py
```

**主要功能：**

```python
from diamond import DiamondConfigClient

# 创建客户端
client = DiamondConfigClient(
    server_url="http://127.0.0.1:8080",
    namespace="default",
    group="DEFAULT_GROUP",
    data_id="app.json"
)

# 获取配置
config = client.get_config()

# 添加变更监听器
client.add_listener(lambda content: print(f"Config changed: {content}"))

# 启动监听
client.start_watch()

# 停止监听
client.stop_watch()
```

## Java 示例

纯标准库实现，无需额外依赖：

```bash
cd examples/java
javac DiamondExample.java
java DiamondExample
```

**主要功能：**

```java
DiamondExample client = new DiamondExample(
    "http://127.0.0.1:8080", "default", "DEFAULT_GROUP", "app.json");

// 获取配置
String config = client.getConfig();

// 添加变更监听器
client.addListener(newContent -> {
    System.out.println("Config changed: " + newContent);
});

// 启动监听
client.startWatch();

// 批量监听
List<Map<String, String>> items = new ArrayList<>();
items.add(createItem("default", "DEFAULT_GROUP", "app.json"));
client.batchWatch(items, 30);
```

## 其他语言

由于 go-diamond 提供标准 HTTP REST API，任何支持 HTTP 的语言都可以使用：

**Node.js:**
```javascript
const http = require('http');

function getConfig() {
  return new Promise((resolve, reject) => {
    http.get('http://127.0.0.1:8080/api/v1/configs/default/DEFAULT_GROUP/app.json', (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => resolve(JSON.parse(data)));
    }).on('error', reject);
  });
}
```

**PHP:**
```php
$response = file_get_contents('http://127.0.0.1:8080/api/v1/configs/default/DEFAULT_GROUP/app.json');
$data = json_decode($response, true);
$config = $data['data']['content'] ?? null;
```

**Go (官方客户端):**
```go
client := diamond.NewClient("http://127.0.0.1:8080", "default")
content, err := client.GetConfig("DEFAULT_GROUP", "app.json")
```

## 认证

管理接口需要 Bearer Token 认证：

```python
headers = {"Authorization": "Bearer your-admin-token"}
requests.post(url, json=payload, headers=headers)
```

```java
conn.setRequestProperty("Authorization", "Bearer your-admin-token");
```

## 配置说明

- **namespace**: 命名空间，用于环境隔离（如 dev/test/prod）
- **group**: 配置分组，如 DEFAULT_GROUP、database 等
- **dataId**: 配置项唯一标识，如 app.json、db.json

## 注意事项

1. 服务端默认端口 8080
2. 管理接口默认 token 为 `change-me-in-production`
3. 建议生产环境修改 token 并使用 HTTPS
4. 监听超时建议 30 秒，避免 HTTP keepalive 问题
5. 客户端会自动缓存配置并使用 MD5 检测变更