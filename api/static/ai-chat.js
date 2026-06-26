$(document).ready(function() {
    // 配置
    const config = {
        baseUrl: window.location.origin,
        apiEndpoint: '/api/ai/session/chat',
        token: localStorage.getItem('token') || '',
        defaultModel: 'qwen3'
    };

    // 状态管理
    let currentSession = '';
    let isTyping = false;
    let messageHistory = [];
    let currentDocumentCode = '';
    let currentDocumentTypeName = '';
    let availableDocumentTypes = [];

    // DOM元素
    const $chatMessages = $('#chatMessages');
    const $messageInput = $('#messageInput');
    const $sendBtn = $('#sendBtn');
    const $sessionId = $('#sessionId');
    const $newSessionBtn = $('#newSessionBtn');
    const $typingIndicator = $('#typingIndicator');
    const $modelSelect = $('#modelSelect');
    const $errorMessage = $('#errorMessage');
    const $successMessage = $('#successMessage');
    const $documentTypeModal = $('#documentTypeModal');
    const $documentTypeSelect = $('#documentTypeSelect');
    const $confirmBtn = $('#confirmBtn');
    const $cancelBtn = $('#cancelBtn');
    const $documentTypeInfo = $('#documentTypeInfo');
    const $documentTypeName = $('#documentTypeName');

    // 初始化
    function init() {
        setWelcomeTime();
        setupEventListeners();
        loadSessionFromStorage();
        checkAuth();
        loadDocumentTypes();
    }

    // 设置欢迎消息时间
    function setWelcomeTime() {
        const now = new Date();
        const timeString = now.toLocaleTimeString('zh-CN', { 
            hour: '2-digit', 
            minute: '2-digit' 
        });
        $('#welcomeTime').text(timeString);
    }

    // 设置事件监听器
    function setupEventListeners() {
        // 发送消息
        $sendBtn.on('click', sendMessage);
        $messageInput.on('keydown', function(e) {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                sendMessage();
            }
        });

        // 新会话
        $newSessionBtn.on('click', handleNewSession);

        // 文档类型对话框
        $confirmBtn.on('click', confirmDocumentType);
        $cancelBtn.on('click', cancelDocumentType);
        $documentTypeSelect.on('change', function() {
            $confirmBtn.prop('disabled', !$(this).val());
        });

        // 对话框回车键支持
        $documentTypeSelect.on('keydown', function(e) {
            if (e.key === 'Enter' && !$confirmBtn.prop('disabled')) {
                e.preventDefault();
                confirmDocumentType();
            }
        });

        // 确定按钮回车键支持
        $confirmBtn.on('keydown', function(e) {
            if (e.key === 'Enter') {
                e.preventDefault();
                confirmDocumentType();
            }
        });

        // 点击对话框外部关闭
        $documentTypeModal.on('click', function(e) {
            if ($(e.target).is('.modal')) {
                cancelDocumentType();
            }
        });



        // 模型选择
        $modelSelect.on('change', function() {
            // 模型选择只在当前会话中有效
        });

        // 自动调整输入框高度
        $messageInput.on('input', function() {
            this.style.height = 'auto';
            this.style.height = Math.min(this.scrollHeight, 100) + 'px';
        });
    }

    // 检查认证状态
    function checkAuth() {
        if (!config.token) {
            showError('请先登录获取认证 token');
            return false;
        }
        return true;
    }

    // 从localStorage加载会话
    function loadSessionFromStorage() {
        // 由于token已写死，session管理改为内存存储
        // 页面刷新后会重置为新会话
    }

    // 保存会话到localStorage
    function saveSessionToStorage() {
        // 由于token已写死，session管理改为内存存储
        // 只在当前会话中保持session状态
    }

    // 加载文档类型列表
    function loadDocumentTypes() {
        fetch(config.baseUrl + '/api/knowsource/my-document-type/list', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + config.token
            }
        })
        .then(response => response.json())
        .then(data => {
            if (data.code === 200 && data.data && data.data.list) {
                availableDocumentTypes = data.data.list;
                // 填充下拉框
                $documentTypeSelect.empty();
                $documentTypeSelect.append('<option value="">请选择文档类型</option>');
                data.data.list.forEach(docType => {
                    $documentTypeSelect.append(
                        `<option value="${docType.documentTypeCode}">${docType.documentTypeName}</option>`
                    );
                });
            }
        })
        .catch(error => {
            console.error('加载文档类型失败:', error);
        });
    }

    // 处理新建会话
    function handleNewSession() {
        if (availableDocumentTypes.length === 0) {
            // 如果没有文档类型，直接创建新会话
            startNewSession('');
            return;
        }
        
        // 显示文档类型选择对话框
        $documentTypeSelect.val('');
        $confirmBtn.prop('disabled', true);
        $documentTypeModal.show();
        
        // 默认选择第一个文档类型
        if (availableDocumentTypes.length > 0) {
            const firstDocType = availableDocumentTypes[0];
            $documentTypeSelect.val(firstDocType.documentTypeCode);
            $confirmBtn.prop('disabled', false);
        }
        
        // 设置焦点到确定按钮
        setTimeout(() => {
            $confirmBtn.focus();
        }, 100);
    }

    // 确认文档类型
    function confirmDocumentType() {
        const selectedCode = $documentTypeSelect.val();
        if (!selectedCode) {
            showError('请选择文档类型');
            return;
        }
        
        const selectedType = availableDocumentTypes.find(dt => dt.documentTypeCode === selectedCode);
        currentDocumentCode = selectedCode;
        currentDocumentTypeName = selectedType ? selectedType.documentTypeName : selectedCode;
        
        $documentTypeModal.hide();
        startNewSession(selectedCode);
        
        // 设置焦点到输入框
        setTimeout(() => {
            $messageInput.focus();
        }, 100);
    }

    // 取消文档类型选择
    function cancelDocumentType() {
        $documentTypeModal.hide();
    }

    // 开始新会话
    function startNewSession(documentCode) {
        currentSession = '';
        currentDocumentCode = documentCode || '';
        $sessionId.text('新会话');
        
        // 更新文档类型显示
        if (currentDocumentCode && currentDocumentTypeName) {
            $documentTypeName.text(currentDocumentTypeName);
            $documentTypeInfo.show();
        } else {
            $documentTypeInfo.hide();
        }
        
        // 清空聊天记录
        $chatMessages.empty();
        
        // 添加欢迎消息
        addMessage('assistant', '你好！我是AI咨询助手，有什么可以帮到您的？');
        
        showSuccess('已开始新会话');
    }

    // 发送消息
    function sendMessage() {
        const message = $messageInput.val().trim();
        if (!message) return;

        if (!checkAuth()) return;

        // 禁用输入和发送按钮
        setInputState(false);
        
        // 添加用户消息
        addMessage('user', message);
        
        // 清空输入框
        $messageInput.val('').trigger('input');
        
        // 显示打字指示器
        showTypingIndicator();

        // 准备请求数据
        const requestData = {
            message: message,
            session: currentSession,
            prompt: '',
            model: $modelSelect.val(),
            think: false,
            stream: true
        };
        
        // 如果有文档类型，添加到请求中
        if (currentDocumentCode) {
            requestData.documentCode = currentDocumentCode;
        }

        // 发送请求
        sendChatRequest(requestData);
    }

    // 解析多个JSON对象的函数
    function parseJsonObjects(text) {
        const objects = [];
        const lines = text.split('\n');
        
        for (const line of lines) {
            const trimmedLine = line.trim();
            if (trimmedLine === '') continue;
            
            try {
                const obj = JSON.parse(trimmedLine);
                objects.push(obj);
            } catch (e) {
                // 忽略解析失败的行，可能是部分数据
                console.log('跳过无效JSON行:', trimmedLine);
            }
        }
        
        return objects;
    }

    // 发送聊天请求
    function sendChatRequest(data) {
        // 使用fetch API处理流式响应
        fetch(config.baseUrl + config.apiEndpoint, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + config.token
            },
            body: JSON.stringify(data)
        })
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const reader = response.body.getReader();
            const decoder = new TextDecoder();
            let assistantMessage = '';
            let isFirstChunk = true;
            let buffer = ''; // 用于存储跨数据块的JSON片段
            
            function readStream() {
                return reader.read().then(({ done, value }) => {
                    if (done) {
                        hideTypingIndicator();
                        setInputState(true);
                        return;
                    }
                    
                    const chunk = decoder.decode(value, { stream: true });
                    console.log('收到数据块:', chunk); // 调试日志
                    
                    // 将新数据添加到缓冲区
                    buffer += chunk;
                    
                    // 按行分割并处理每个JSON对象
                    const lines = buffer.split('\n');
                    // 保留最后一行（可能不完整）
                    buffer = lines.pop() || '';
                    
                    lines.forEach(line => {
                        const trimmedLine = line.trim();
                        if (trimmedLine === '') return;
                        
                        try {
                            const data = JSON.parse(trimmedLine);
                            console.log('解析到JSON对象:', data); // 调试日志
                            
                            if (data.message && data.message.content !== undefined) {
                                assistantMessage += data.message.content;
                                console.log('当前消息内容:', assistantMessage); // 调试日志
                                
                                if (isFirstChunk) {
                                    // 隐藏打字指示器并添加助手消息
                                    hideTypingIndicator();
                                    addMessage('assistant', assistantMessage, true);
                                    isFirstChunk = false;
                                } else {
                                    // 更新最后一条助手消息
                                    updateLastAssistantMessage(assistantMessage);
                                }
                            }
                            
                            // 检查是否完成
                            if (data.done) {
                                console.log('对话完成'); // 调试日志
                                setInputState(true);
                                if (data.session) {
                                    currentSession = data.session;
                                    $sessionId.text(currentSession);
                                    saveSessionToStorage();
                                }
                            }
                        } catch (e) {
                            console.log('跳过无效JSON行:', trimmedLine);
                        }
                    });
                    
                    return readStream();
                });
            }
            
            return readStream();
        })
        .catch(error => {
            hideTypingIndicator();
            setInputState(true);
            
            console.error('请求失败:', error);
            let errorMsg = '发送消息失败';
            
            if (error.message.includes('401')) {
                errorMsg = '认证失败，请检查token';
            } else if (error.message.includes('403')) {
                errorMsg = '权限不足';
            }
            
            showError(errorMsg);
            addMessage('assistant', '抱歉，我遇到了一些问题，请稍后再试。');
        });
    }



    // 添加消息到聊天界面
    function addMessage(role, content, isStreaming = false) {
        const messageId = 'msg_' + Date.now();
        const now = new Date();
        const timeString = now.toLocaleTimeString('zh-CN', { 
            hour: '2-digit', 
            minute: '2-digit' 
        });

        const avatarIcon = role === 'user' ? 'fas fa-user' : 'fas fa-robot';
        const avatarClass = role === 'user' ? 'user' : 'assistant';

        const messageHtml = `
            <div class="message ${role}" id="${messageId}">
                <div class="message-avatar ${avatarClass}">
                    <i class="${avatarIcon}"></i>
                </div>
                <div class="message-content">
                    ${content}
                    <div class="message-time">${timeString}</div>
                </div>
            </div>
        `;

        $chatMessages.append(messageHtml);
        scrollToBottom();

        // 保存到历史记录
        messageHistory.push({
            id: messageId,
            role: role,
            content: content,
            timestamp: now
        });
    }

    // 更新最后一条助手消息（用于流式响应）
    function updateLastAssistantMessage(content) {
        const $lastAssistantMessage = $chatMessages.find('.message.assistant:last .message-content');
        if ($lastAssistantMessage.length > 0) {
            $lastAssistantMessage.html(content + '<div class="message-time">' + 
                new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }) + '</div>');
        }
    }

    // 显示打字指示器
    function showTypingIndicator() {
        $typingIndicator.show();
        scrollToBottom();
    }

    // 隐藏打字指示器
    function hideTypingIndicator() {
        $typingIndicator.hide();
    }

    // 设置输入状态
    function setInputState(enabled) {
        $messageInput.prop('disabled', !enabled);
        $sendBtn.prop('disabled', !enabled);
        isTyping = !enabled;
    }

    // 滚动到底部
    function scrollToBottom() {
        $chatMessages.scrollTop($chatMessages[0].scrollHeight);
    }

    // 显示错误消息
    function showError(message) {
        $errorMessage.text(message).show();
        setTimeout(() => {
            $errorMessage.hide();
        }, 5000);
    }

    // 显示成功消息
    function showSuccess(message) {
        $successMessage.text(message).show();
        setTimeout(() => {
            $successMessage.hide();
        }, 3000);
    }

    // 导出聊天记录
    function exportChatHistory() {
        const history = messageHistory.map(msg => ({
            role: msg.role,
            content: msg.content,
            timestamp: msg.timestamp.toISOString()
        }));

        const dataStr = JSON.stringify(history, null, 2);
        const dataBlob = new Blob([dataStr], { type: 'application/json' });
        
        const link = document.createElement('a');
        link.href = URL.createObjectURL(dataBlob);
        link.download = `chat_history_${new Date().toISOString().split('T')[0]}.json`;
        link.click();
    }

    // 清空聊天记录
    function clearChatHistory() {
        if (confirm('确定要清空聊天记录吗？')) {
            messageHistory = [];
            $chatMessages.empty();
            addMessage('assistant', '聊天记录已清空。有什么可以帮助您的吗？');
        }
    }

    // 键盘快捷键
    $(document).on('keydown', function(e) {
        // Ctrl+Enter 发送消息
        if (e.ctrlKey && e.key === 'Enter') {
            e.preventDefault();
            sendMessage();
        }
        
        // Ctrl+N 新会话
        if (e.ctrlKey && e.key === 'n') {
            e.preventDefault();
            handleNewSession();
        }
        
        // Ctrl+E 导出聊天记录
        if (e.ctrlKey && e.key === 'e') {
            e.preventDefault();
            exportChatHistory();
        }
    });

    // 初始化应用
    init();

    // 全局函数（用于调试）
    window.AIChat = {
        exportHistory: exportChatHistory,
        clearHistory: clearChatHistory,
        startNewSession: handleNewSession,
        getSession: () => currentSession,
        getHistory: () => messageHistory,
        getDocumentCode: () => currentDocumentCode
    };
});
