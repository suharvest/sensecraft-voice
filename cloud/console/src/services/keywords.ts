import { request } from './api';
import { 
  Keyword, 
  KeywordQueryParams, 
  CreateKeywordRequest, 
  UpdateKeywordRequest,
  PaginatedResponse 
} from '@/types';

export const keywordService = {
  // 获取关键词列表
  async getKeywords(params: KeywordQueryParams = {}): Promise<PaginatedResponse<Keyword>> {
    return request.get<PaginatedResponse<Keyword>>('/v1/keywords', { params });
  },

  // 获取单个关键词
  async getKeyword(id: number): Promise<Keyword> {
    return request.get<Keyword>(`/v1/keywords/${id}`);
  },

  // 创建关键词
  async createKeyword(data: CreateKeywordRequest): Promise<Keyword> {
    return request.post<Keyword>('/v1/keywords', data);
  },

  // 更新关键词
  async updateKeyword(id: number, data: UpdateKeywordRequest): Promise<Keyword> {
    return request.put<Keyword>(`/v1/keywords/${id}`, data);
  },

  // 删除关键词
  async deleteKeyword(id: number): Promise<void> {
    return request.delete<void>(`/v1/keywords/${id}`);
  },

  // 批量删除关键词
  async batchDeleteKeywords(ids: number[]): Promise<{ deleted_count: number; deleted_ids: number[] }> {
    return request.delete<{ deleted_count: number; deleted_ids: number[] }>('/v1/keywords/batch', { data: { ids } });
  },
};
