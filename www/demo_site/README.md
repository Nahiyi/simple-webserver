# Web服务器演示站点

这是一个用于演示Go语言Web服务器功能的示例站点。

## 目录结构

```
demo_site/
├── index.html              # 首页
├── about.html              # 关于我们
├── README.md               # 说明文档
└── products/
    ├── index.html           # 产品列表
    ├── data.json           # JSON数据示例
    ├── electronics/         # 电子产品
    │   └── index.html
    ├── books/             # 图书列表
    │   └── index.html
    ├── images/            # 图片资源
    │   ├── index.html
    │   ├── logo.svg       # SVG矢量图(实际文件)
    │   ├── logo.png       # PNG图片(待填充)
    │   ├── banner.jpg     # JPG图片(待填充)
    │   ├── icon.gif       # GIF图片(待填充)
    │   ├── bg.webp        # WebP图片(待填充)
    │   └── photo.bmp      # BMP图片(待填充)
    ├── css/
    │   └── style.css       # CSS样式示例
    └── js/
        ├── main.js         # JavaScript示例
        └── app.js          # 应用脚本示例
```

## 使用方法

1. 将此目录拷贝到服务器的资源根目录（如 `/root`）
2. 在控制面板设置根目录指向父目录
3. 启动服务器后访问 `/demo_site/`
4. 浏览各层级目录，测试不同文件类型的显示/下载行为

## 文件类型支持

| 类型 | 扩展名 | 浏览器行为 |
|------|--------|-----------|
| 网页 | .html | 直接显示 |
| 样式 | .css | 直接显示 |
| 脚本 | .js | 直接显示 |
| 图形 | .svg | 直接显示 |
| 图片 | .png/.jpg/.gif/.webp/.bmp | 直接显示 |
| 数据 | .json | 直接显示 |
| 文档 | .md | 直接显示 |
| 其他 | 其他格式 | 下载 |
