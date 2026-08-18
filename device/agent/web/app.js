// 语音识别系统前端应用逻辑
let ws = null;
let reconnectAttempts = 0;
const maxReconnectAttempts = 5;
let conversationHistory = [];
const maxConversations = 20;
let isRecording = false;

// 获取当前访问的IP地址和端口
function getCurrentHost() {
    return window.location.protocol + '//' + window.location.host;
}

// 显示指定标签页
function showTab(tabName) {
    // 隐藏所有标签页
    const tabContents = document.querySelectorAll('.tab-content');
    tabContents.forEach(tab => tab.classList.remove('active'));
    
    // 移除所有导航链接的active类
    const navLinks = document.querySelectorAll('.nav-link');
    navLinks.forEach(link => link.classList.remove('active'));
    
    // 显示选中的标签页
    const targetTab = document.getElementById(tabName);
    if (targetTab) {
        targetTab.classList.add('active');
    } else {
        console.warn(`Tab with id '${tabName}' not found`);
    }
    
    // 添加active类到对应的导航链接
    if (event && event.target) {
        event.target.classList.add('active');
    }
}

// 检查录音状态
async function checkRecordingStatus() {
    try {
        const response = await fetch(getCurrentHost() + '/v1/voice/status');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        if (data.code === 200) {
            updateRecordingStatus(data.result.running);
        } else {
            throw new Error(data.message || '获取录音状态失败');
        }
    } catch (error) {
        console.error('获取录音状态失败:', error);
        updateRecordingStatus(false);
    }
}

// 更新录音状态显示
function updateRecordingStatus(running) {
    const micButton = document.getElementById('mic-button');
    const recordingStatus = document.getElementById('recording-status');
    
    if (!micButton || !recordingStatus) {
        console.warn('Recording status elements not found');
        return;
    }
    
    isRecording = running;
    
    if (running) {
        // 正在录音
        micButton.className = 'mic-button recording';
        micButton.innerHTML = '⏹️';
        micButton.title = '点击停止录音';
        recordingStatus.className = 'recording-status recording';
        recordingStatus.textContent = '正在录音...';
    } else {
        // 未在录音
        micButton.className = 'mic-button not-recording';
        micButton.innerHTML = '🎤';
        micButton.title = '点击开始录音';
        recordingStatus.className = 'recording-status not-recording';
        recordingStatus.textContent = '未在录音';
    }
    
    // 启用按钮
    micButton.disabled = false;
}

// 切换录音状态
async function toggleRecording() {
    const action = isRecording ? 'stop' : 'start';
    const micButton = document.getElementById('mic-button');
    
    if (!micButton) {
        console.warn('Microphone button not found');
        return;
    }
    
    // 禁用按钮防止重复点击
    micButton.disabled = true;
    
    try {
        const response = await fetch(getCurrentHost() + '/v1/voice/record', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                action: action
            })
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        if (data.code === 200) {
            // 操作成功后重新检查状态
            setTimeout(checkRecordingStatus, 500);
        } else {
            throw new Error(data.message || `${action === 'start' ? '开始' : '停止'}录音失败`);
        }
    } catch (error) {
        console.error(`${action === 'start' ? '开始' : '停止'}录音失败:`, error);
        alert(`${action === 'start' ? '开始' : '停止'}录音失败: ` + error.message);
        // 恢复按钮状态
        micButton.disabled = false;
    }
}

// 自动连接WebSocket
function autoConnectWebSocket() {
    try {
        // 如果已经有连接，先关闭
        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.close();
        }
        
        // 使用当前访问的IP地址建立WebSocket连接
        const wsUrl = 'ws://' + window.location.host + '/v1/voice/asr-ws';
        console.log('正在连接WebSocket:', wsUrl);
        ws = new WebSocket(wsUrl);
        
        ws.onopen = function(event) {
            console.log('WebSocket连接已建立');
            updateConnectionStatus(true, '已连接');
            reconnectAttempts = 0;
        };
        
        ws.onmessage = function(event) {
            try {
                const data = JSON.parse(event.data);
                addConversation(data);
            } catch (error) {
                console.error('解析WebSocket消息失败:', error);
            }
        };
        
        ws.onclose = function(event) {
            console.log('WebSocket连接已关闭, code:', event.code, 'reason:', event.reason);
            updateConnectionStatus(false, '连接已断开');
            
            // 只有在非正常关闭时才重连
            if (event.code !== 1000 && event.code !== 1001) {
                // 自动重连逻辑
                if (reconnectAttempts < maxReconnectAttempts) {
                    reconnectAttempts++;
                    console.log(`尝试重连... (${reconnectAttempts}/${maxReconnectAttempts})`);
                    setTimeout(autoConnectWebSocket, 5000); // 增加到5秒
                } else {
                    console.log('达到最大重连次数，停止重连');
                    updateConnectionStatus(false, '重连失败，请刷新页面');
                }
            } else {
                console.log('正常关闭连接，不进行重连');
            }
        };
        
        ws.onerror = function(error) {
            console.error('WebSocket错误:', error);
            updateConnectionStatus(false, '连接错误');
        };
        
    } catch (error) {
        console.error('创建WebSocket连接失败:', error);
        updateConnectionStatus(false, '连接失败');
    }
}

// 手动重连WebSocket
function reconnectWebSocket() {
    console.log('手动重连WebSocket');
    reconnectAttempts = 0;
    autoConnectWebSocket();
}

// 关闭WebSocket连接
function closeWebSocket() {
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.close(1000, '用户主动关闭');
    }
}

// 更新连接状态显示
function updateConnectionStatus(connected, text) {
    const statusElement = document.getElementById('ws-status');
    const statusTextElement = document.getElementById('ws-status-text');
    
    if (!statusElement || !statusTextElement) {
        console.warn('WebSocket status elements not found');
        return;
    }
    
    if (connected) {
        statusElement.className = 'status-indicator connected';
    } else {
        statusElement.className = 'status-indicator disconnected';
    }
    
    statusTextElement.textContent = text;
}

// 添加新的对话记录
function addConversation(data) {
    // 创建对话对象
    const conversation = {
        id: Date.now() + Math.random(), // 唯一ID
        speaker: data.speaker || {},
        text: data.text || '',
        timestamp: data.timestamp || Date.now(),
        type: data.type || 'unknown'
    };

    // 添加到历史记录开头
    conversationHistory.unshift(conversation);
    
    // 保持最多20条记录
    if (conversationHistory.length > maxConversations) {
        conversationHistory = conversationHistory.slice(0, maxConversations);
    }
    
    // 更新显示
    updateConversationDisplay();
}

// 更新对话显示
function updateConversationDisplay() {
    const conversationList = document.getElementById('conversation-list');
    const conversationCount = document.getElementById('conversation-count');
    
    if (!conversationList || !conversationCount) {
        console.warn('Conversation display elements not found');
        return;
    }
    
    // 更新计数
    conversationCount.textContent = `${conversationHistory.length} 条`;
    
    if (conversationHistory.length === 0) {
        // 显示空状态
        conversationList.innerHTML = `
            <div class="empty-state">
                <div class="empty-state-icon">💬</div>
                <div class="empty-state-text">暂无对话记录</div>
                <div class="empty-state-subtext">正在连接WebSocket，稍后开始接收语音识别数据</div>
            </div>
        `;
        return;
    }
    
    // 生成对话HTML
    const conversationHTML = conversationHistory.map(conv => {
        const speaker = conv.speaker;
        const speakerName = speaker.speaker_name || '未知用户';
        const speakerId = speaker.speaker_id || 'unknown';
        const timestamp = new Date(conv.timestamp).toLocaleString('zh-CN');
        
        // 从speaker_id中提取数字作为头像
        let avatarText = '?';
        if (speakerId && speakerId.includes('speaker_')) {
            const match = speakerId.match(/speaker_(\d+)/);
            if (match) {
                avatarText = match[1];
            }
        }
        
        return `
            <div class="conversation-item">
                <div class="speaker-avatar">${avatarText}</div>
                <div class="conversation-content">
                    <div class="speaker-info">
                        <span class="speaker-name">${speakerName}</span>
                        <span class="speaker-id">${speakerId}</span>
                    </div>
                    <div class="conversation-text">${conv.text}</div>
                    <div class="conversation-meta">
                        <span class="timestamp">${timestamp}</span>
                    </div>
                </div>
            </div>
        `;
    }).join('');
    
    conversationList.innerHTML = conversationHTML;
}

// 清空对话记录
function clearConversation() {
    if (confirm('确定要清空所有对话记录吗？')) {
        conversationHistory = [];
        updateConversationDisplay();
    }
}

// 麦克风测试
async function testMicrophone() {
    const testBtn = document.getElementById('test-mic-btn');
    if (!testBtn) {
        console.warn('Test microphone button not found');
        return;
    }

    const originalText = testBtn.textContent;
    
    try {
        // 禁用按钮防止重复点击
        testBtn.disabled = true;
        testBtn.textContent = '🎤 录音中...';

        // 调用快速录音接口，录制3秒
        const response = await fetch(getCurrentHost() + '/v1/voice/quick', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                seconds: 3,
                sampleRate: 16000,
                channels: 1
            })
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        if (data.code === 200 && data.result && data.result.path) {
            // 播放录制的音频
            await playAudioFile(data.result.path);
            testBtn.textContent = '✅ 测试完成';
        } else {
            throw new Error(data.message || '录音失败');
        }
    } catch (error) {
        console.error('麦克风测试失败:', error);
        testBtn.textContent = '❌ 测试失败';
        alert('麦克风测试失败: ' + error.message);
    } finally {
        // 2秒后恢复按钮状态
        setTimeout(() => {
            testBtn.disabled = false;
            testBtn.textContent = originalText;
        }, 2000);
    }
}

// 播放音频文件
async function playAudioFile(audioPath) {
    return new Promise((resolve, reject) => {
        // 构造完整的音频URL
        const audioUrl = getCurrentHost() + '/' + audioPath;
        console.log('尝试播放音频:', audioUrl);
        
        // 创建音频元素
        const audio = new Audio(audioUrl);
        
        // 设置音频事件监听
        audio.onloadeddata = () => {
            console.log('音频文件加载完成，开始播放');
            audio.play().catch(err => {
                console.error('播放音频失败:', err);
                // 不直接reject，而是显示提示信息
                alert('音频播放失败，但录音已成功完成。文件路径: ' + audioPath);
                resolve(); // 仍然resolve，因为录音成功了
            });
        };
        
        audio.onended = () => {
            console.log('音频播放完成');
            resolve();
        };
        
        audio.onerror = (error) => {
            console.error('音频加载失败:', error);
            // 不直接reject，而是显示提示信息
            alert('音频文件无法加载，但录音已成功完成。文件路径: ' + audioPath);
            resolve(); // 仍然resolve，因为录音成功了
        };
        
        // 设置超时
        setTimeout(() => {
            if (!audio.ended) {
                audio.pause();
                console.log('音频播放超时，但录音已成功完成');
                resolve(); // 超时也resolve，因为录音成功了
            }
        }, 10000); // 10秒超时
        
        // 开始加载音频
        audio.load();
    });
}

// 刷新设备状态
async function refreshDeviceStatus() {
    try {
        const response = await fetch(getCurrentHost() + '/v1/voice/device/status');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        if (data.code === 200) {
            updateDeviceStatus(data.result);
        } else {
            throw new Error(data.message || '获取设备状态失败');
        }
    } catch (error) {
        console.error('获取设备状态失败:', error);
        alert('获取设备状态失败: ' + error.message);
    }
}

// 更新设备状态显示
function updateDeviceStatus(status) {
    // 系统资源
    const cpuUsageElement = document.getElementById('cpu-usage');
    if (cpuUsageElement) {
        cpuUsageElement.textContent = status.cpu_usage_percent.toFixed(2) + '%';
    }
    
    const memoryUsageElement = document.getElementById('memory-usage');
    if (memoryUsageElement) {
        memoryUsageElement.textContent = formatBytes(status.memory_used_bytes);
    }
    
    const recordingTimeLeftElement = document.getElementById('recording-time-left');
    if (recordingTimeLeftElement) {
        recordingTimeLeftElement.textContent = formatTimeLeft(status.recording_time_left_seconds);
    }
    
    // 设备信息
    const macAddressElement = document.getElementById('mac-address');
    if (macAddressElement) {
        macAddressElement.textContent = status.mac_address;
    }
    
    const ipAddressElement = document.getElementById('ip-address');
    if (ipAddressElement) {
        ipAddressElement.textContent = status.ip_address;
    }
    
    const versionElement = document.getElementById('version');
    if (versionElement) {
        versionElement.textContent = status.version;
    }
    
    // 远程连接 - 从设备状态中获取，但输入框的值由refreshRemoteConfig单独管理
    if (status.remote_connection_status === '已连接') {
        setConnectionStatus('connection-success', '已连接');
    } else {
        setConnectionStatus('connection-failed', '连接失败');
    }
    
    // 最后更新时间
    if (status.last_update) {
        const lastUpdateElement = document.getElementById('last-update');
        if (lastUpdateElement) {
            const date = new Date(status.last_update * 1000);
            lastUpdateElement.textContent = date.toLocaleString('zh-CN');
        }
    }
}

// 格式化字节数
function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// 格式化剩余时间
function formatTimeLeft(seconds) {
    if (seconds === null || seconds === undefined) {
        return 'N/A';
    }
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = Math.floor(seconds % 60);
    return `${h}小时${m}分${s}秒`;
}

// 验证远程URL格式
function validateRemoteUrl() {
    const input = document.getElementById('remote-base-url-input');
    const updateBtn = document.getElementById('update-remote-btn');
    const messageDiv = document.getElementById('url-validation-message');
    
    if (!input || !updateBtn || !messageDiv) {
        console.warn('页面元素未找到，跳过URL验证');
        return;
    }
    
    const url = input.value.trim();
    
    // 清除之前的验证状态
    input.style.borderColor = '#e1e5e9';
    messageDiv.style.display = 'none';
    messageDiv.className = 'url-validation-message';
    updateBtn.disabled = true;
    
    if (!url) {
        return;
    }
    
    // 基本URL格式验证
    const urlPattern = /^https?:\/\/[^\s/$.?#].[^\s]*$/i;
    if (!urlPattern.test(url)) {
        input.style.borderColor = '#dc3545';
        messageDiv.textContent = '请输入有效的URL格式 (如: http://example.com:3008)';
        messageDiv.className = 'url-validation-message error';
        messageDiv.style.display = 'block';
        return;
    }
    
    // 检查协议
    if (!url.startsWith('http://') && !url.startsWith('https://')) {
        input.style.borderColor = '#dc3545';
        messageDiv.textContent = 'URL必须以 http:// 或 https:// 开头';
        messageDiv.className = 'url-validation-message error';
        messageDiv.style.display = 'block';
        return;
    }
    
    // 检查主机名
    try {
        const urlObj = new URL(url);
        if (!urlObj.hostname) {
            input.style.borderColor = '#dc3545';
            messageDiv.textContent = 'URL必须包含有效的主机名';
            messageDiv.className = 'url-validation-message error';
            messageDiv.style.display = 'block';
            return;
        }
    } catch (e) {
        input.style.borderColor = '#dc3545';
        messageDiv.textContent = 'URL格式不正确';
        messageDiv.className = 'url-validation-message error';
        messageDiv.style.display = 'block';
        return;
    }
    
    // 验证通过
    input.style.borderColor = '#28a745';
    messageDiv.textContent = 'URL格式正确';
    messageDiv.className = 'url-validation-message success';
    messageDiv.style.display = 'block';
    updateBtn.disabled = false;
}

// 处理回车键
function handleRemoteUrlKeyPress(event) {
    if (event.key === 'Enter') {
        const updateBtn = document.getElementById('update-remote-btn');
        if (updateBtn && !updateBtn.disabled) {
            updateRemoteConfig();
        } else if (!updateBtn) {
            console.warn('Update remote button not found');
        }
    }
}

// 获取远程配置
async function getRemoteConfig() {
    try {
        const response = await fetch(getCurrentHost() + '/v1/voice/config/remote');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        if (data.code === 200) {
            return data.result;
        } else {
            throw new Error(data.message || '获取远程配置失败');
        }
    } catch (error) {
        console.error('获取远程配置失败:', error);
        throw error;
    }
}

// 更新远程配置
async function updateRemoteConfig() {
    const inputElement = document.getElementById('remote-base-url-input');
    const updateBtn = document.getElementById('update-remote-btn');
    
    if (!inputElement || !updateBtn) {
        console.warn('页面元素未找到！');
        alert('页面元素未找到！');
        return;
    }
    
    const newBaseUrl = inputElement.value.trim();
    if (!newBaseUrl) {
        alert('请输入远程服务地址！');
        return;
    }
    
    if (updateBtn.disabled) {
        alert('请先输入有效的URL格式！');
        return;
    }
    
    const originalText = updateBtn.textContent;
    
    try {
        updateBtn.textContent = '更新中...';
        updateBtn.disabled = true;
        
        const response = await fetch(getCurrentHost() + '/v1/voice/config/remote', {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                base_url: newBaseUrl
            })
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        if (data.code === 200) {
            // 显示成功消息
            const messageDiv = document.getElementById('url-validation-message');
            messageDiv.textContent = '配置更新成功！';
            messageDiv.className = 'url-validation-message success';
            messageDiv.style.display = 'block';
            
            // 刷新设备状态以显示新配置
            await refreshDeviceStatus();
            
            // 3秒后隐藏成功消息
            setTimeout(() => {
                messageDiv.style.display = 'none';
            }, 3000);
        } else {
            throw new Error(data.message || '更新远程配置失败');
        }
    } catch (error) {
        console.error('更新远程配置失败:', error);
        const messageDiv = document.getElementById('url-validation-message');
        messageDiv.textContent = '更新失败: ' + error.message;
        messageDiv.className = 'url-validation-message error';
        messageDiv.style.display = 'block';
    } finally {
        updateBtn.textContent = originalText;
        updateBtn.disabled = false;
    }
}

// 设置连接状态显示
function setConnectionStatus(status, text) {
    const statusElement = document.getElementById('remote-connection-status');
    if (!statusElement) {
        console.warn('remote-connection-status element not found');
        return;
    }
    statusElement.textContent = text;
    statusElement.className = `status-value ${status}`;
}

// 显示连接测试消息
function showConnectionMessage(message, type) {
    const messageDiv = document.getElementById('url-validation-message');
    if (messageDiv) {
        // 格式化消息，使其更易读
        const formattedMessage = message.replace(/\|/g, '\n• ');
        messageDiv.innerHTML = formattedMessage.replace(/\n/g, '<br>');
        messageDiv.className = 'url-validation-message ' + (type === 'success' ? 'success' : 'error');
        messageDiv.style.display = 'block';
        
        // 6秒后自动隐藏消息
        setTimeout(() => {
            messageDiv.style.display = 'none';
        }, 6000);
    }
}

// 刷新远程配置
async function refreshRemoteConfig() {
    try {
        const config = await getRemoteConfig();
        const inputElement = document.getElementById('remote-base-url-input');
        if (inputElement) {
            inputElement.value = config.base_url || '';
        } else {
            console.warn('Remote base URL input element not found');
        }
        
        // 设置连接状态
        if (config.enabled) {
            setConnectionStatus('connection-success', '已连接');
        } else {
            setConnectionStatus('connection-failed', '连接失败');
        }
    } catch (error) {
        console.error('刷新远程配置失败:', error);
        setConnectionStatus('connection-failed', '获取失败');
    }
}

// 刷新远程配置并自动测试连接
async function refreshRemoteConfigAndTest() {
    try {
        const config = await getRemoteConfig();
        const inputElement = document.getElementById('remote-base-url-input');
        if (inputElement) {
            inputElement.value = config.base_url || '';
        } else {
            console.warn('Remote base URL input element not found');
        }
        
        // 如果有配置的远程URL，自动测试连接
        if (config.base_url && config.base_url.trim()) {
            await testRemoteConnectionSilent();
        } else {
            setConnectionStatus('connection-failed', '未配置远程地址');
        }
    } catch (error) {
        console.error('刷新远程配置失败:', error);
        setConnectionStatus('connection-failed', '获取失败');
    }
}

// 静默测试远程连接（不显示alert）
async function testRemoteConnectionSilent() {
    const inputElement = document.getElementById('remote-base-url-input');
    if (!inputElement) {
        console.warn('页面元素未找到！');
        setConnectionStatus('connection-failed', '页面元素未找到');
        return;
    }
    
    const remoteBaseUrl = inputElement.value.trim();
    if (!remoteBaseUrl) {
        setConnectionStatus('connection-failed', '未配置远程地址');
        return;
    }

    // 设置测试中状态
    setConnectionStatus('connection-checking', '测试中...');

    try {
        const response = await fetch(getCurrentHost() + '/v1/voice/config/remote/test', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                base_url: remoteBaseUrl
            })
        });
        
        if (!response.ok) {
            setConnectionStatus('connection-failed', '连接失败');
            return;
        }
        
        const data = await response.json();
        if (data.code === 200) {
            const result = data.result;
            if (result.reachable) {
                setConnectionStatus('connection-success', '已连接');
            } else {
                setConnectionStatus('connection-failed', '连接失败');
            }
        } else {
            setConnectionStatus('connection-failed', '连接失败');
        }
    } catch (error) {
        console.error('远程连接测试失败:', error);
        setConnectionStatus('connection-failed', '连接失败');
    }
}

// 测试远程连接
async function testRemoteConnection() {
    const inputElement = document.getElementById('remote-base-url-input');
    if (!inputElement) {
        console.warn('页面元素未找到！');
        showConnectionMessage('❌ 页面元素未找到！', 'error');
        return;
    }
    
    const remoteBaseUrl = inputElement.value.trim();
    if (!remoteBaseUrl) {
        showConnectionMessage('❌ 请先输入远程服务地址！', 'error');
        return;
    }

    // 设置测试中状态
    setConnectionStatus('connection-checking', '测试中...');

    try {
        const response = await fetch(getCurrentHost() + '/v1/voice/config/remote/test', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                base_url: remoteBaseUrl
            })
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        if (data.code === 200) {
            const result = data.result;
            if (result.reachable) {
                setConnectionStatus('connection-success', '连接成功');
                showConnectionMessage(
                    `✅ 连接成功！\n• 状态码: ${result.status_code}\n• 响应时间: ${result.response_time_ms}ms\n• 测试URL: ${result.test_url}`, 
                    'success'
                );
            } else {
                setConnectionStatus('connection-failed', '连接失败');
                showConnectionMessage(
                    `❌ 连接失败！\n• 状态码: ${result.status_code}\n• 响应时间: ${result.response_time_ms}ms\n• 错误信息: ${result.message}`, 
                    'error'
                );
            }
        } else {
            throw new Error(data.message || '测试连接失败');
        }
    } catch (error) {
        console.error('远程连接测试失败:', error);
        setConnectionStatus('connection-failed', '连接失败');
        showConnectionMessage('❌ 远程连接测试失败: ' + error.message, 'error');
    }
}

// 页面加载完成后自动连接WebSocket和获取设备状态
document.addEventListener('DOMContentLoaded', function() {
    // 延迟连接WebSocket，避免页面加载时的冲突
    setTimeout(() => {
        autoConnectWebSocket();
    }, 1000);
    
    // 检查录音状态
    checkRecordingStatus();
    
    // 自动获取设备状态
    refreshDeviceStatus();
    
    // 自动获取远程配置并测试连接
    refreshRemoteConfigAndTest();
    
    // 每30秒自动刷新一次设备状态
    setInterval(refreshDeviceStatus, 30000);
    
    // 每10秒检查一次录音状态
    setInterval(checkRecordingStatus, 10000);
});
