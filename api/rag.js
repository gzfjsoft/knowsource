(function() {
    console.log("rag.js: 开始执行 RAG 检索...");
    
    try {
        // RAG 检索参数
        const query = {{query}};
        const keys = {{keys}};
        const documentCode = {{documentCode}};
        const tags = {{tags}};
        const collectionPrefix = {{collectionPrefix}};
        const RAGURL = {{RAGURL}};
        keysArray = keys.split(",");

        const k = 10;
        let newResponseData = [];
        collectionName  = collectionPrefix+ documentCode;

        //{"query":"无人机","top_k":20,"use_rerank":true,"collection_name":" 新能技术","rerank_top_k":5}

        const requestData = {
            query: query,
            top_k: k,
            use_rerank: true,
            collection_name: collectionName,
            rerank_top_k: k,
            tags:tags
        };
        console.log("请求参数:", JSON.stringify(requestData, null, 2));
        
        const response = httpPost(
            RAGURL, 
            JSON.stringify(requestData)
        );
        console.log("响应:", response);
        responseData = JSON.parse(response);
        // 解析响应并过滤结果
        let maxRerankScore = 0;
        for (let i = 0; i < responseData.data.list.length; i++) {
           if (responseData.data.list[i].rerank_score > maxRerankScore) {
              maxRerankScore = responseData.data.list[i].rerank_score;
              console.log("maxRerankScore:", maxRerankScore);
           }
           let minScore = 0.05;
         
           if (maxRerankScore >= 0.9) {
              minScore = 0.3;
           }
           
           if (responseData.data.list[i].rerank_score > minScore) {
              newResponseData.push(
                  {
                      content: responseData.data.list[i].content,
                      rerank_score: responseData.data.list[i].rerank_score
                  }
              );
           }
        }
         
        
        console.log("===================newResponseData===========", newResponseData);

        try {
            
            
            markdownText = ``
            let jsonfileinfo = [];
           
            // 将搜索结果转换为markdown格式
            if (newResponseData ) {
                markdownText = newResponseData.map((result, index) => {

                    //只保留 ### DocInfo: 位置之前的字符
                    const docInfoIndex = result.content.indexOf('### DocInfo:');
                    documentWithoutDocInfo = docInfoIndex !== -1 ? result.content.substring(0, docInfoIndex) : result.content;


                    return `# 资料库第${index + 1}条信息（重排分数: ${result.rerank_score}）:\n\n${documentWithoutDocInfo}\n\n`;
                }).join('\n');
                
                // 提取每个文档的DocInfo信息
                newResponseData.forEach((result, index) => {
                    
                    const docInfoMatch = result.content.match(/### DocInfo:\s*(\{[\s\S]*?\})/);
                    if (docInfoMatch) {
                        try {
                            
                            jsonfileinfo.push(docInfoMatch[1]);
                                
                        } catch (parseError) {
                            console.error(`解析文档 ${index + 1} 的DocInfo失败:`, parseError.message);
                        }
                    }
                    
                });
            }
            
            console.log("合并的jsonfileinfo:", jsonfileinfo);

            return {
                success: true,
                response: markdownText,
                jsonfileinfo: jsonfileinfo,
                error: "成功"
            };
            
        } catch (parseError) {
            console.error("响应解析失败:", parseError.message);
            return {
                success: false,
                error: "响应解析失败: " + parseError.message,
                rawResponse: response,
                jsonfileinfo:""
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