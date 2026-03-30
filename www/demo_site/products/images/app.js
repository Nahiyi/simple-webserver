// 演示JavaScript文件
console.log('Web服务器演示 - app.js 已加载');

// 显示欢迎信息
function showWelcome() {
    const messages = [
        '欢迎访问Web服务器演示站点',
        '这是一个3层嵌套目录结构',
        '支持多种文件类型直接显示'
    ];
    console.log(messages.join('\n'));
}

// 定时更新
setInterval(() => {
    const time = new Date().toLocaleTimeString();
    document.title = `演示站点 - ${time}`;
}, 1000);

showWelcome();
