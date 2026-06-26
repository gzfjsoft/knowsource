(function() {
    console.log("尝试访问外网...");
    try {
        // 这里可以替换为任何你想访问的外网URL
        const response = httpGet("https://gzfjsoft.com");
        console.log("JS：请求成功，响应长度:", response.length);
        console.log("JS：响应前100字符:", response.substring(0, 100));
        
        // 返回结果给Go代码
        return {
            success: true,
            length: response.length,
            content: response.substring(0, 100),
            message: "请求成功"
        };
    } catch (e) {
        console.error("请求失败:", e.message);
        
        // 返回错误信息给Go代码
        return {
            success: false,
            error: e.message
        };
    }
})()