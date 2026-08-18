import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import { message } from 'antd';
import { envConfig } from '@/config/env';

// 创建axios实例（直连模式：使用环境变量中的后端地址 + /api 前缀）
const api: AxiosInstance = axios.create({
  baseURL: `${envConfig.API_BASE_URL.replace(/\/$/, '')}/api`,
  timeout: envConfig.TIMEOUT,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    // 从localStorage获取token
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
api.interceptors.response.use(
  (response: AxiosResponse) => {
    const { data } = response;
    console.log('API Response interceptor - raw response:', response);
    console.log('API Response interceptor - data:', data);

    // 兼容后端返回 code 为 0 或 200 作为成功
    const code = data?.code;
    console.log('API Response interceptor - code:', code);
    
    if (code !== 0 && code !== 200) {
      console.error('API Response interceptor - code check failed:', code);
      message.error(data?.message || '请求失败');
      return Promise.reject(new Error(data?.message || '请求失败'));
    }

    // 如果响应有result字段，返回result；否则返回整个data
    const result = data.result !== undefined ? data.result : data;
    console.log('API Response interceptor - returning:', result);
    return result;
  },
  (error) => {
    console.error('API Response interceptor - error:', error);
    console.error('API Response interceptor - error.response:', error.response);
    console.error('API Response interceptor - error.request:', error.request);
    
    if (error.response) {
      const { status, data } = error.response;
      console.error('API Response interceptor - error status:', status);
      console.error('API Response interceptor - error data:', data);
      
      switch (status) {
        case 401:
          message.error('登录已过期，请重新登录');
          localStorage.removeItem('token');
          window.location.href = '/login';
          break;
        case 403:
          message.error('没有权限访问');
          break;
        case 404:
          message.error('请求的资源不存在');
          break;
        case 500:
          message.error('服务器内部错误');
          break;
        default:
          message.error(data?.message || '请求失败');
      }
    } else if (error.request) {
      console.error('API Response interceptor - network error:', error.request);
      message.error('网络连接失败，请检查网络');
    } else {
      console.error('API Response interceptor - config error:', error.message);
      message.error('请求配置错误');
    }
    
    return Promise.reject(error);
  }
);

// 通用请求方法
export const request = {
  get: <T = any>(url: string, config?: AxiosRequestConfig): Promise<T> =>
    api.get(url, config).then((res: any) => {
      // 如果响应拦截器已经处理了 result，直接返回
      // 否则返回 res.result
      return res.result !== undefined ? res.result : res;
    }),
    
  post: <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> =>
    api.post(url, data, config).then((res: any) => {
      return res.result !== undefined ? res.result : res;
    }),
    
  put: <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> =>
    api.put(url, data, config).then((res: any) => {
      return res.result !== undefined ? res.result : res;
    }),
    
  delete: <T = any>(url: string, config?: AxiosRequestConfig): Promise<T> =>
    api.delete(url, config).then((res: any) => {
      return res.result !== undefined ? res.result : res;
    }),
    
  patch: <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> =>
    api.patch(url, data, config).then((res: any) => {
      return res.result !== undefined ? res.result : res;
    }),
};

export default api;
