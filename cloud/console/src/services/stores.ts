import { request } from './api';
import { Store, PaginatedResponse, StoreQueryParams } from '@/types';

export const storeService = {
  // 获取门店列表
  getStores: (params: StoreQueryParams = {}): Promise<PaginatedResponse<Store>> => {
    const searchParams = new URLSearchParams();
    
    if (params.offset !== undefined) {
      searchParams.append('offset', params.offset.toString());
    }
    if (params.limit !== undefined) {
      searchParams.append('limit', params.limit.toString());
    }
    if (params.name) {
      searchParams.append('name', params.name);
    }
    if (params.code) {
      searchParams.append('code', params.code);
    }
    
    const queryString = searchParams.toString();
    const url = `/v1/stores${queryString ? `?${queryString}` : ''}`;
    
    return request.get<PaginatedResponse<Store>>(url);
  },

  // 获取门店详情
  getStore: (id: number): Promise<Store> => {
    return request.get<Store>(`/v1/stores/${id}`);
  },

  // 创建门店
  createStore: (data: Omit<Store, 'id' | 'created_at' | 'updated_at'>): Promise<Store> => {
    return request.post<Store>('/v1/stores', data);
  },

  // 更新门店
  updateStore: (id: number, data: Partial<Store>): Promise<Store> => {
    return request.put<Store>(`/v1/stores/${id}`, data);
  },

  // 删除门店
  deleteStore: (id: number): Promise<{ message: string }> => {
    return request.delete<{ message: string }>(`/v1/stores/${id}`);
  },

  // 批量删除门店
  batchDeleteStores: (ids: number[]): Promise<{ message: string }> => {
    return request.post<{ message: string }>('/v1/stores/batch-delete', { ids });
  },

  // 获取门店的所有点位
  getStoreLocations: (storeId: number, params: { offset?: number; limit?: number } = {}): Promise<PaginatedResponse<any>> => {
    const searchParams = new URLSearchParams();
    
    if (params.offset !== undefined) {
      searchParams.append('offset', params.offset.toString());
    }
    if (params.limit !== undefined) {
      searchParams.append('limit', params.limit.toString());
    }
    
    const queryString = searchParams.toString();
    const url = `/v1/stores/${storeId}/locations${queryString ? `?${queryString}` : ''}`;
    
    return request.get<PaginatedResponse<any>>(url);
  },

  // 获取门店的所有设备
  getStoreDevices: (storeId: number, params: { offset?: number; limit?: number } = {}): Promise<PaginatedResponse<any>> => {
    const searchParams = new URLSearchParams();
    
    if (params.offset !== undefined) {
      searchParams.append('offset', params.offset.toString());
    }
    if (params.limit !== undefined) {
      searchParams.append('limit', params.limit.toString());
    }
    
    const queryString = searchParams.toString();
    const url = `/v1/stores/${storeId}/devices${queryString ? `?${queryString}` : ''}`;
    
    return request.get<PaginatedResponse<any>>(url);
  },
};
