import { request } from './api';
import { Device, PaginatedResponse, DeviceQueryParams } from '@/types';

export const deviceService = {
  // 获取设备列表
  getDevices: (params: DeviceQueryParams = {}): Promise<PaginatedResponse<Device>> => {
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
    if (params.mac_address) {
      searchParams.append('mac_address', params.mac_address);
    }
    if (params.location_id !== undefined) {
      searchParams.append('location_id', params.location_id.toString());
    }
    if (params.store_id !== undefined) {
      searchParams.append('store_id', params.store_id.toString());
    }
    
    const queryString = searchParams.toString();
    const url = `/v1/devices${queryString ? `?${queryString}` : ''}`;
    
    return request.get<PaginatedResponse<Device>>(url);
  },

  // 获取门店的所有设备
  getStoreDevices: (storeId: number, params: DeviceQueryParams = {}): Promise<PaginatedResponse<Device>> => {
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
    if (params.mac_address) {
      searchParams.append('mac_address', params.mac_address);
    }
    
    const queryString = searchParams.toString();
    const url = `/v1/stores/${storeId}/devices${queryString ? `?${queryString}` : ''}`;
    
    return request.get<PaginatedResponse<Device>>(url);
  },

  // 获取点位的设备
  getLocationDevices: (locationId: number, params: DeviceQueryParams = {}): Promise<PaginatedResponse<Device>> => {
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
    if (params.mac_address) {
      searchParams.append('mac_address', params.mac_address);
    }
    
    const queryString = searchParams.toString();
    const url = `/v1/locations/${locationId}/devices${queryString ? `?${queryString}` : ''}`;
    
    return request.get<PaginatedResponse<Device>>(url);
  },

  // 获取设备详情 - 根据MAC地址查询
  getDeviceByMac: (mac: string): Promise<Device> => {
    return request.get<Device>(`/v1/devices/${mac}`);
  },

  // 获取设备详情 - 根据ID查询（如果后端支持）
  getDevice: (id: number): Promise<Device> => {
    return request.get<Device>(`/v1/devices/${id}`);
  },

  // 更新设备名称
  updateDeviceName: (id: number, name: string): Promise<Device> => {
    return request.put<Device>(`/v1/devices/${id}/name`, { name });
  },

  // 更新设备信息
  updateDevice: (id: number, data: Partial<Device>): Promise<Device> => {
    return request.put<Device>(`/v1/devices/${id}`, data);
  },

  // 设备分配到点位
  assignDevice: (id: number, locationId: number): Promise<{ message: string }> => {
    return request.put<{ message: string }>(`/v1/devices/${id}/assign`, {
      location_id: locationId
    });
  },

  // 取消设备分配
  unassignDevice: (id: number): Promise<{ message: string }> => {
    return request.put<{ message: string }>(`/v1/devices/${id}/assign`, {
      location_id: null
    });
  },

  // 批量分配设备
  batchAssignDevices: (deviceIds: number[], locationId: number): Promise<{ message: string }> => {
    return request.post<{ message: string }>('/v1/devices/batch-assign', {
      device_ids: deviceIds,
      location_id: locationId
    });
  },
};
