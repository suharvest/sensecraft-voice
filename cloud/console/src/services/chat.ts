import envConfig from "@/config/env";

// v2 API 请求参数
export interface ChatV2Request {
  session_id?: string;              // 可选，会话ID
  message: string;                  // 必填，用户消息内容
  user_id: string;                  // 必填，用户ID
  system_prompt_id?: number;        // 可选，系统提示词ID
  system_prompt_content?: string;   // 可选，系统提示词内容
}

// v2 API 响应事件数据
export interface ChatV2EventData {
  event: 'message' | 'completed' | 'error';
  data: {
    session_id: string;
    content?: string;               // message事件时包含
    timestamp: number;
    error?: string;                 // error事件时包含
  };
}

// 流式聊天响应接口
export interface StreamingChatResponse {
  onMessage: (callback: (content: string) => void) => void;
  onComplete: (callback: () => void) => void;
  onError: (callback: (error: Error) => void) => void;
  cancel: () => void;
}

class ChatService {
  private baseURL = envConfig.API_BASE_URL;
  private currentSessionId: string | null = null;
  
  
  /**
   * 发送流式聊天消息 (v2 API)
   */
  async sendStreamingMessage(
    message: string, 
    userId: string = 'web-user',
    systemPromptId?: number,
    systemPromptContent?: string
  ): Promise<StreamingChatResponse> {
    const controller = new AbortController();
    
    return new Promise((resolve, reject) => {
      const callbacks = {
        onMessage: (callback: (content: string) => void) => {
          callbacks.messageCallback = callback;
        },
        onComplete: (callback: () => void) => {
          callbacks.completeCallback = callback;
        },
        onError: (callback: (error: Error) => void) => {
          callbacks.errorCallback = callback;
        },
        cancel: () => {
          controller.abort();
        },
        messageCallback: null as ((content: string) => void) | null,
        completeCallback: null as (() => void) | null,
        errorCallback: null as ((error: Error) => void) | null,
      };

      // 立即返回响应对象
      resolve(callbacks);

      // 发起流式请求
      this.startStreamingV2(message, userId, systemPromptId, systemPromptContent, controller, callbacks).catch(reject);
    });
  }

  private async startStreamingV2(
    message: string,
    userId: string,
    systemPromptId: number | undefined,
    systemPromptContent: string | undefined,
    controller: AbortController,
    callbacks: any
  ) {
    try {
      const requestData: ChatV2Request = {
        session_id: this.currentSessionId || undefined, // 第一次为空，后续使用现有会话ID
        message,
        user_id: userId,
      };

      // 添加系统提示词支持
      if (systemPromptId) {
        requestData.system_prompt_id = systemPromptId;
      } else if (systemPromptContent) {
        requestData.system_prompt_content = systemPromptContent;
      }

      const response = await fetch(`${this.baseURL}/api/v2/openai/chat/stream`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'text/event-stream',
        },
        body: JSON.stringify(requestData),
        signal: controller.signal,
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const reader = response.body?.getReader();
      if (!reader) {
        throw new Error('无法获取响应流');
      }

      const decoder = new TextDecoder();
      let buffer = '';

      while (true) {
        const { done, value } = await reader.read();
        
        if (done) {
          break;
        }

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || ''; // 保留最后一行（可能不完整）

        for (const line of lines) {
          if (line.trim() === '') continue;
          
          try {
            // 解析 Server-Sent Events 格式
            if (line.startsWith('data: ')) {
              const dataStr = line.substring(6); // 移除 'data: ' 前缀
              
              if (dataStr.trim() === '[DONE]') {
                // 流式响应结束
                callbacks.completeCallback?.();
                return;
              }

              const eventData: ChatV2EventData = JSON.parse(dataStr);
              this.handleStreamEventV2(eventData, callbacks);
            }
          } catch (error) {
            console.warn('解析流式数据失败:', line, error);
          }
        }
      }

      callbacks.completeCallback?.();
    } catch (error) {
      if (error instanceof Error && error.name === 'AbortError') {
        // 用户取消请求
        return;
      }
      
      console.error('流式聊天错误:', error);
      callbacks.errorCallback?.(error as Error);
    }
  }

  private handleStreamEventV2(eventData: ChatV2EventData, callbacks: any) {
    try {
      console.log('处理v2流式事件:', eventData.event, eventData);
      
      // 保存会话ID（从任何事件中获取）
      if (eventData.data.session_id && !this.currentSessionId) {
        this.currentSessionId = eventData.data.session_id;
        console.log('保存会话ID:', this.currentSessionId);
      }
      
      // v2 API 简化的事件处理
      switch (eventData.event) {
        case 'message':
          // 处理消息内容
          if (eventData.data.content) {
            console.log('收到消息片段:', eventData.data.content);
            callbacks.messageCallback?.(eventData.data.content);
          }
          break;
          
        case 'completed':
          console.log('对话完成:', eventData);
          callbacks.completeCallback?.();
          break;
          
        case 'error':
          console.error('AI服务返回错误:', eventData);
          // 创建错误对象并调用错误回调
          const errorMessage = eventData.data.error || '未知错误';
          const error = new Error(errorMessage);
          callbacks.errorCallback?.(error);
          break;
          
        default:
          console.log('未知事件类型:', eventData.event, eventData);
          break;
      }
    } catch (error) {
      console.error('处理v2流式事件失败:', error);
    }
  }

  /**
   * 测试API连接 (v2)
   */
  async testConnection(): Promise<boolean> {
    try {
      console.log('正在测试AI API连接:', `${this.baseURL}/api/v2/openai/chat/stream`);
      
      const response = await fetch(`${this.baseURL}/api/v2/openai/chat/stream`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'text/event-stream',
        },
        body: JSON.stringify({
          message: 'test connection',
          user_id: 'test-user',
        } as ChatV2Request),
      });
      
      console.log('API连接测试响应:', response.status, response.statusText);
      return response.ok;
    } catch (error) {
      console.error('API连接测试失败:', error);
      // 如果是CORS错误，可能服务是运行的，但浏览器阻止了请求
      if (error instanceof TypeError && error.message.includes('Failed to fetch')) {
        console.warn('可能是CORS问题或服务未启动');
      }
      return false;
    }
  }

  /**
   * 重置会话
   */
  resetSession(): void {
    this.currentSessionId = null;
  }

  /**
   * 获取当前会话ID
   */
  getCurrentSessionId(): string | null {
    return this.currentSessionId;
  }
}

export const chatService = new ChatService();
