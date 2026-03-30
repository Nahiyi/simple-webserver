// 主脚本文件
document.addEventListener('DOMContentLoaded', function() {
    console.log('页面加载完成 - main.js');

    // 添加交互效果
    const links = document.querySelectorAll('a');
    links.forEach(link => {
        link.addEventListener('click', function(e) {
            console.log('点击链接:', this.href);
        });
    });
});

// 工具函数
function formatDate() {
    const now = new Date();
    return now.toLocaleDateString() + ' ' + now.toLocaleTimeString();
}

function log(msg) {
    console.log(`[${formatDate()}] ${msg}`);
}
