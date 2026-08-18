import { request } from './api';
import { Recording, PaginatedResponse, RecordingQueryParams, KeywordMatchRecord, KeywordMatchQueryParams } from '@/types';

export const recordingService = {
  // 获取录音记录列表
  getRecordings: (params: RecordingQueryParams = {}): Promise<PaginatedResponse<Recording>> => {
    const searchParams = new URLSearchParams();
    
    // 默认添加 type=1 参数
    searchParams.append('type', '1');
    
    if (params.offset !== undefined) {
      searchParams.append('offset', params.offset.toString());
    }
    if (params.limit !== undefined) {
      searchParams.append('limit', params.limit.toString());
    }
    if (params.mac_address) {
      searchParams.append('mac_address', params.mac_address);
    }
    if (params.mac_addresses && params.mac_addresses.length > 0) {
      params.mac_addresses.forEach(mac => {
        searchParams.append('mac_address', mac);
      });
    }
    if (params.status !== undefined) {
      searchParams.append('status', params.status.toString());
    }
    if (params.start_time !== undefined) {
      searchParams.append('start_time', params.start_time.toString());
    }
    if (params.end_time !== undefined) {
      searchParams.append('end_time', params.end_time.toString());
    }
    if (params.store_id !== undefined) {
      searchParams.append('store_id', params.store_id.toString());
    }
    if (params.location_id !== undefined) {
      searchParams.append('location_id', params.location_id.toString());
    }
    if (params.device_id !== undefined) {
      searchParams.append('device_id', params.device_id.toString());
    }
    if (params.type !== undefined) {
      // 如果明确指定了 type 参数，覆盖默认值
      searchParams.set('type', params.type.toString());
    }
    
    const queryString = searchParams.toString();
    const url = `/v1/recordings${queryString ? `?${queryString}` : ''}`;
    
    return request.get<PaginatedResponse<Recording>>(url);
  },

  // 获取录音记录详情
  getRecording: (id: number): Promise<Recording> => {
    return request.get<Recording>(`/v1/recordings/${id}`);
  },

  // 删除录音记录
  deleteRecording: (id: number): Promise<{ message: string }> => {
    return request.delete<{ message: string }>(`/v1/recordings/${id}`);
  },

  // 批量删除录音记录
  batchDeleteRecordings: (ids: number[]): Promise<{ message: string }> => {
    return request.post<{ message: string }>('/v1/recordings/batch-delete', { ids });
  },

  // 导出录音记录
  exportRecordings: (params: RecordingQueryParams = {}): Promise<Blob> => {
    const searchParams = new URLSearchParams();
    
    // 默认添加 type=1 参数
    searchParams.append('type', '1');
    
    if (params.offset !== undefined) {
      searchParams.append('offset', params.offset.toString());
    }
    if (params.limit !== undefined) {
      searchParams.append('limit', params.limit.toString());
    }
    if (params.mac_address) {
      searchParams.append('mac_address', params.mac_address);
    }
    if (params.mac_addresses && params.mac_addresses.length > 0) {
      params.mac_addresses.forEach(mac => {
        searchParams.append('mac_address', mac);
      });
    }
    if (params.status !== undefined) {
      searchParams.append('status', params.status.toString());
    }
    if (params.start_time !== undefined) {
      searchParams.append('start_time', params.start_time.toString());
    }
    if (params.end_time !== undefined) {
      searchParams.append('end_time', params.end_time.toString());
    }
    if (params.store_id !== undefined) {
      searchParams.append('store_id', params.store_id.toString());
    }
    if (params.location_id !== undefined) {
      searchParams.append('location_id', params.location_id.toString());
    }
    if (params.device_id !== undefined) {
      searchParams.append('device_id', params.device_id.toString());
    }
    if (params.type !== undefined) {
      // 如果明确指定了 type 参数，覆盖默认值
      searchParams.set('type', params.type.toString());
    }
    
    const queryString = searchParams.toString();
    const url = `/v1/recordings/export${queryString ? `?${queryString}` : ''}`;
    
    return request.get<Blob>(url, {
      responseType: 'blob',
    });
  },

  // 获取关键词匹配记录
  getKeywordMatches: (params: KeywordMatchQueryParams = {}): Promise<PaginatedResponse<KeywordMatchRecord>> => {
    const searchParams = new URLSearchParams();
    
    if (params.offset !== undefined) {
      searchParams.append('offset', params.offset.toString());
    }
    if (params.limit !== undefined) {
      searchParams.append('limit', params.limit.toString());
    }
    if (params.mac_address) {
      searchParams.append('mac_address', params.mac_address);
    }
    if (params.keyword_id !== undefined) {
      searchParams.append('keyword_id', params.keyword_id.toString());
    }
    if (params.recording_id !== undefined) {
      searchParams.append('recording_id', params.recording_id.toString());
    }
    if (params.start_time !== undefined) {
      searchParams.append('start_time', params.start_time.toString());
    }
    if (params.end_time !== undefined) {
      searchParams.append('end_time', params.end_time.toString());
    }
    if (params.store_id !== undefined) {
      searchParams.append('store_id', params.store_id.toString());
    }
    
    const queryString = searchParams.toString();
    const url = `/v1/recordings/keyword-matches${queryString ? `?${queryString}` : ''}`;
    
    return request.get<PaginatedResponse<KeywordMatchRecord>>(url);
  },
};
