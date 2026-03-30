# Web服务器演示站点

这是一个用于演示Go语言Web服务器功能的示例站点。

## 目录结构

```
demo_site/
├── index.html          # 首页
├── about.html          # 关于我们
├── products/
│   ├── index.html      # 产品列表
│   ├── electronics/    # 电子产品
│   │   └── index.html
│   ├── books/          # 图书
│   │   └── index.html
│   ├── images/         # 图片资源
│   │   ├── index.html
│   │   ├── logo.svg
│   │   └── app.js
│   ├── css/
│   │   └── style.css
│   ├── js/
│   │   └── main.js
│   └── data.json
```

## 文件类型

| 类型 | 说明 | 浏览器行为 |
|------|------|-----------|
| HTML | 网页内容 | 直接显示 |
| CSS | 样式表 | 直接显示 |
| JS | 脚本代码 | 直接显示 |
| SVG | 矢量图形 | 直接显示 |
| JSON | 数据格式 | 直接显示 |
| PNG/JPG | 图片 | 直接显示 |
| 其他 | 未知格式 | 下载 |

## 使用方法

1. 将此目录拷贝到服务器的资源根目录
2. 在控制面板设置根目录指向父目录
3. 启动服务器后访问 `/demo_site/`
