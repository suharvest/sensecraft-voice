import envConfig from "@/config/env";
import { 
  SystemPrompt, 
  SystemPromptAPI,
  CreateSystemPromptRequest, 
  UpdateSystemPromptRequest, 
  SystemPromptQueryParams,
  ApiResponse,
  PaginatedResponse 
} from "@/types";

class SystemPromptService {
  private baseURL = envConfig.API_BASE_URL;

  /**
   * 将API响应格式转换为内部格式
   */
  private mapSystemPrompt(apiPrompt: SystemPromptAPI): SystemPrompt {
    return {
      id: apiPrompt.ID,
      name: apiPrompt.Name,
      role: apiPrompt.Role,
      content: apiPrompt.Content,
      tags: apiPrompt.Tags,
      is_active: apiPrompt.IsActive,
      version: apiPrompt.Version,
      created_at: apiPrompt.CreatedAt,
      updated_at: apiPrompt.UpdatedAt
    };
  }

  /**
   * 将内部格式转换为API请求格式
   */
  private mapToAPIFormat(prompt: CreateSystemPromptRequest | UpdateSystemPromptRequest): any {
    const result: any = {};
    if ('name' in prompt && prompt.name !== undefined) result.Name = prompt.name;
    if ('role' in prompt && prompt.role !== undefined) result.Role = prompt.role;
    if ('content' in prompt && prompt.content !== undefined) result.Content = prompt.content;
    if ('tags' in prompt && prompt.tags !== undefined) result.Tags = prompt.tags;
    if ('is_active' in prompt && prompt.is_active !== undefined) result.IsActive = prompt.is_active;
    return result;
  }

  /**
   * 获取系统提示词列表
   */
  async getSystemPrompts(params?: SystemPromptQueryParams): Promise<PaginatedResponse<SystemPrompt>> {
    const searchParams = new URLSearchParams();
    
    if (params?.name) searchParams.append('name', params.name);
    if (params?.role) searchParams.append('role', params.role);
    if (params?.active !== undefined) searchParams.append('active', params.active.toString());
    if (params?.offset) searchParams.append('offset', params.offset.toString());
    if (params?.limit) searchParams.append('limit', params.limit.toString());

    const response = await fetch(`${this.baseURL}/api/v2/openai/system-prompts?${searchParams}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`获取系统提示词列表失败: ${response.status}`);
    }

    const result: ApiResponse<PaginatedResponse<SystemPromptAPI>> = await response.json();
    
    // 检查API响应码
    if (result.code !== 200) {
      throw new Error(`API错误: ${result.message || '未知错误'}`);
    }
    
    const apiResult = result.result || { items: [], total: 0, offset: 0, limit: 10 };
    
    // 转换字段格式
    return {
      items: apiResult.items.map(item => this.mapSystemPrompt(item)),
      total: apiResult.total,
      offset: apiResult.offset,
      limit: apiResult.limit
    };
  }

  /**
   * 获取单个系统提示词
   */
  async getSystemPrompt(id: number): Promise<SystemPrompt> {
    const response = await fetch(`${this.baseURL}/api/v2/openai/system-prompts/${id}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`获取系统提示词失败: ${response.status}`);
    }

    const result: ApiResponse<SystemPromptAPI> = await response.json();
    
    if (result.code !== 200) {
      throw new Error(`API错误: ${result.message || '未知错误'}`);
    }
    
    return this.mapSystemPrompt(result.result);
  }

  /**
   * 创建系统提示词
   */
  async createSystemPrompt(data: CreateSystemPromptRequest): Promise<SystemPrompt> {
    const apiData = this.mapToAPIFormat(data);
    
    const response = await fetch(`${this.baseURL}/api/v2/openai/system-prompts`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(apiData),
    });

    if (!response.ok) {
      throw new Error(`创建系统提示词失败: ${response.status}`);
    }

    const result: ApiResponse<SystemPromptAPI> = await response.json();
    
    if (result.code !== 200) {
      throw new Error(result.message || '未知错误');
    }
    
    return this.mapSystemPrompt(result.result);
  }

  /**
   * 更新系统提示词
   */
  async updateSystemPrompt(id: number, data: UpdateSystemPromptRequest): Promise<SystemPrompt> {
    const apiData = this.mapToAPIFormat(data);
    
    const response = await fetch(`${this.baseURL}/api/v2/openai/system-prompts/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(apiData),
    });

    if (!response.ok) {
      throw new Error(`更新系统提示词失败: ${response.status}`);
    }

    const result: ApiResponse<SystemPromptAPI> = await response.json();
    
    if (result.code !== 200) {
      throw new Error(result.message || '未知错误');
    }
    
    return this.mapSystemPrompt(result.result);
  }

  /**
   * 删除系统提示词
   */
  async deleteSystemPrompt(id: number): Promise<void> {
    const response = await fetch(`${this.baseURL}/api/v2/openai/system-prompts/${id}`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`删除系统提示词失败: ${response.status}`);
    }

    const result: ApiResponse<any> = await response.json();
    
    if (result.code !== 200) {
      throw new Error(`API错误: ${result.message || '未知错误'}`);
    }
  }

  /**
   * 批量删除系统提示词
   */
  async batchDeleteSystemPrompts(ids: number[]): Promise<void> {
    const response = await fetch(`${this.baseURL}/api/v2/openai/system-prompts`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ ids }),
    });

    if (!response.ok) {
      throw new Error(`批量删除系统提示词失败: ${response.status}`);
    }

    const result: ApiResponse<any> = await response.json();
    
    if (result.code !== 200) {
      throw new Error(`API错误: ${result.message || '未知错误'}`);
    }
  }

  /**
   * 更新系统提示词状态
   */
  async updateSystemPromptStatus(id: number, isActive: boolean): Promise<void> {
    const response = await fetch(`${this.baseURL}/api/v2/openai/system-prompts/${id}/status`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ is_active: isActive }),
    });

    if (!response.ok) {
      throw new Error(`更新系统提示词状态失败: ${response.status}`);
    }

    const result: ApiResponse<any> = await response.json();
    
    if (result.code !== 200) {
      throw new Error(result.message || '未知错误');
    }
  }

  /**
   * 按名称模糊搜索系统提示词
   */
  async searchSystemPrompts(name: string, limit: number = 20): Promise<SystemPrompt[]> {
    const searchParams = new URLSearchParams();
    searchParams.append('name', name);
    searchParams.append('limit', Math.min(limit, 100).toString());

    const response = await fetch(`${this.baseURL}/api/v2/openai/system-prompts/search?${searchParams}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`搜索系统提示词失败: ${response.status}`);
    }

    const result: ApiResponse<{
      count: number;
      items: SystemPromptAPI[];
      limit: number;
      name: string;
    }> = await response.json();
    
    if (result.code !== 200) {
      throw new Error(result.message || '搜索失败');
    }
    
    return (result.result?.items || []).map(item => this.mapSystemPrompt(item));
  }

  /**
   * 获取激活的系统提示词列表
   */
  async getActiveSystemPrompts(): Promise<SystemPrompt[]> {
    const response = await this.getSystemPrompts({ active: true, limit: 100 });
    return response.items;
  }

  /**
   * 获取默认系统提示词
   */
  async getDefaultSystemPrompt(): Promise<SystemPrompt | null> {
    try {
      const response = await this.getSystemPrompts({ active: true, limit: 1 });
      return response.items.length > 0 ? response.items[0] : null;
    } catch (error) {
      console.error('获取默认系统提示词失败:', error);
      return null;
    }
  }
}

export const systemPromptService = new SystemPromptService();
