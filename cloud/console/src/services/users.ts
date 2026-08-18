import { request } from './api';
import { User, PaginatedResponse, UserQueryParams, LoginResponse } from '@/types';
import { encryptPassword } from '@/utils/crypto';

export const userService = {
  // 用户注册
  register: (data: { username: string; password: string }): Promise<User> => {
    // 对密码进行MD5加密
    const encryptedPassword = encryptPassword(data.password);
    return request.post<User>('/v1/users/register', {
      username: data.username,
      password: encryptedPassword
    });
  },

  // 用户登录
  login: (data: { username: string; password: string }): Promise<LoginResponse> => {
    // 对密码进行MD5加密
    const encryptedPassword = encryptPassword(data.password);
    return request.post<LoginResponse>('/v1/users/login', {
      username: data.username,
      password: encryptedPassword
    });
  },

  // 获取用户列表
  getUsers: (params: UserQueryParams = {}): Promise<PaginatedResponse<User>> => {
    const searchParams = new URLSearchParams();
    
    if (params.offset !== undefined) {
      searchParams.append('offset', params.offset.toString());
    }
    if (params.limit !== undefined) {
      searchParams.append('limit', params.limit.toString());
    }
    if (params.username) {
      searchParams.append('username', params.username);
    }
    
    const queryString = searchParams.toString();
    const url = `/v1/users${queryString ? `?${queryString}` : ''}`;
    
    return request.get<PaginatedResponse<User>>(url);
  },

  // 获取用户详情
  getUser: (id: number): Promise<User> => {
    return request.get<User>(`/v1/users/${id}`);
  },

  // 更新用户信息
  updateUser: (id: number, data: Partial<User>): Promise<User> => {
    return request.put<User>(`/v1/users/${id}`, data);
  },

  // 修改密码
  changePassword: (id: number, data: { old_password: string; new_password: string }): Promise<{ message: string }> => {
    // 对密码进行MD5加密
    const encryptedOldPassword = encryptPassword(data.old_password);
    const encryptedNewPassword = encryptPassword(data.new_password);
    return request.put<{ message: string }>(`/v1/users/${id}/password`, {
      old_password: encryptedOldPassword,
      new_password: encryptedNewPassword
    });
  },

  // 删除用户
  deleteUser: (id: number): Promise<{ message: string }> => {
    return request.delete<{ message: string }>(`/v1/users/${id}`);
  },

  // 批量删除用户
  batchDeleteUsers: (ids: number[]): Promise<{ message: string }> => {
    return request.post<{ message: string }>('/v1/users/batch-delete', { ids });
  },

  // 检查用户是否已登录
  checkAuth: (): Promise<User> => {
    return request.get<User>('/v1/users/me');
  },

  // 用户登出
  logout: (): Promise<{ message: string }> => {
    return request.post<{ message: string }>('/v1/users/logout');
  },
};
