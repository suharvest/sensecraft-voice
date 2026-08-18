import { request } from './api';
import { Location, PaginatedResponse, LocationQueryParams } from '@/types';

export const locationService = {
  // 获取点位列表
  getLocations: (params: LocationQueryParams = {}): Promise<PaginatedResponse<Location>> => {
    const searchParams = new URLSearchParams();
    
    if (params.offset !== undefined) {
      searchParams.append('offset', params.offset.toString());
    }
    if (params.limit !== undefined) {
      searchParams.append('limit', params.limit.toString());
    }
    if (params.store_id !== undefined) {
      searchParams.append('store_id', params.store_id.toString());
    }
    if (params.name) {
      searchParams.append('name', params.name);
    }
    if (params.code) {
      searchParams.append('code', params.code);
    }
    
    const queryString = searchParams.toString();
    const url = `/v1/locations${queryString ? `?${queryString}` : ''}`;
    
    return request.get<PaginatedResponse<Location>>(url);
  },

  // 获取点位详情
  getLocation: (id: number): Promise<Location> => {
    return request.get<Location>(`/v1/locations/${id}`);
  },

  // 创建点位
  createLocation: (data: Omit<Location, 'id' | 'created_at' | 'updated_at'>): Promise<Location> => {
    return request.post<Location>('/v1/locations', data);
  },

  // 更新点位信息
  updateLocation: (id: number, data: Partial<Location>): Promise<Location> => {
    return request.put<Location>(`/v1/locations/${id}`, data);
  },

  // 删除点位
  deleteLocation: (id: number): Promise<{ message: string }> => {
    return request.delete<{ message: string }>(`/v1/locations/${id}`);
  },

  // 批量删除点位
  batchDeleteLocations: (ids: number[]): Promise<{ message: string }> => {
    return request.post<{ message: string }>('/v1/locations/batch-delete', { ids });
  },
};
