(function() {
    console.log("开始执行 RAG 检索...");
    
    try {
        // RAG 检索参数
        const query = "{{query}}";
        const keys = "{{keys}}";
        const tags = "{{tags}}";
        keysArray = keys ? keys.split(",") : [];
        tagsArray = tags ? tags.split(",") : [];


        const k = 10;
        const useReranker = true;
        const instruction = "";
        
        // 构建请求数据
        const requestData = {
            query: query,
            k: k,
            use_reranker: useReranker,
            instruction: instruction
        };
        
        // 如果有 tags，添加到请求数据中
        if (tagsArray.length > 0) {
            requestData.tags = tagsArray;
        }
        
        console.log("请求参数:", JSON.stringify(requestData, null, 2));
        
        // 发送 HTTP POST 请求到知识库 RAG 端点
        const response = httpPost(
            // "http://127.0.0.1:5001/knowledge-base/rag/retrieve",
            "http://xxxxx.com:6751/knowledge-base/rag/retrieve",
            JSON.stringify(requestData)
        );
        
        // 解析响应并过滤结果
        let responseData;
        try {
            responseData = JSON.parse(response);
            
            // 过滤掉 rerank_score 小于 0.8 的结果
            if (responseData.results && Array.isArray(responseData.results)) {
                const originalCount = responseData.results.length;
                
                // 显示所有结果的 rerank_score
                console.log("所有结果的 rerank_score:");
                responseData.results.forEach((result, index) => {
                    console.log(`结果 ${index + 1}: ${result.rerank_score}`);
                });
                
                let filteredResults = responseData.results.filter(result => {
                    return result.rerank_score >= 0.8;
                });

              

                console.log("keysArray: ", keysArray);
                console.log("tagsArray: ", tagsArray);
                
                // 如果有 tags，根据 metadata 中的 tag 字段过滤结果
                if (tagsArray.length > 0) {
                    filteredResults = filteredResults.filter(result => {
                        // 检查 metadata 中是否有 tag 字段
                        if (result.metadata && result.metadata.tag) {
                            const resultTag = result.metadata.tag;
                            // 检查 resultTag 是否在 tagsArray 中
                            return tagsArray.some(tag => tag.trim() === resultTag.trim());
                        }
                        return false;
                    });
                }
                
                // if (keysArray.length > 0) {
                //     //content 必须要包含 keysArray 中的全部字符串
                //     filteredResults = filteredResults.filter(result => {
                //         for (let i = 0; i < keysArray.length; i++) {
                //             if (result.document && result.document.toLowerCase().includes(keysArray[i].toLowerCase())) {
                //                 return true;
                //             }
                //         }
                //         return false;                      
                //     });
                // }
                 
                
                // 更新结果数组
                responseData.results = filteredResults;
                
                console.log(`过滤前结果数量: ${originalCount}`);
                console.log(`过滤后结果数量: ${filteredResults.length}`);
                console.log(`过滤掉 ${originalCount - filteredResults.length} 个低质量结果 (rerank_score < 0.8)`);
                
                // 如果过滤后没有符合条件的结果，返回空串
                if (filteredResults.length === 0) {
                    console.log("没有符合条件的结果，返回空串");
                    return {
                        success: true,
                        response: "",
                        error: ""
                    };
                }
            }
            
            markdownText = ``
           
            // 将搜索结果转换为markdown格式
            if (responseData.results && Array.isArray(responseData.results)) {
                markdownText = responseData.results.map((result, index) => {
                    return `# 资料库第${index + 1}条信息（重排分数: ${result.rerank_score}）:\n\n${result.document}\n\n`;
                }).join('\n');
            }
            

            return {
                success: true,
                response: markdownText,
                error: ""
            };
            
        } catch (parseError) {
            console.error("响应解析失败:", parseError.message);
            return {
                success: false,
                error: "响应解析失败: " + parseError.message,
                rawResponse: response
            };
        }


        
    } catch (e) {
        console.error("RAG 检索失败:", e.message);
        
        // 返回错误信息给Go代码
        return {
            success: false,
            error: e.message
        };
    }
})()