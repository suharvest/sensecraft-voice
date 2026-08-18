import { request } from './api';
import { 
  DashboardStats, 
  WeeklyTrendItem
} from '@/types';

// 最近录音类型（从录音列表API获取）
export interface RecentRecording {
  id: number;
  mac_address: string;
  speaker_id: string;
  speaker_name: string;
  text: string;
  status: number; // 0=中间结果，1=最终结果
  created_at: number;
  device_time: number;
  device_name?: string; // 设备名称，可选字段
}

// 系统活动类型（模拟数据，实际项目中可能需要单独的API）
export interface SystemActivity {
  id: number;
  type: 'recording' | 'store' | 'user' | 'location';
  message: string;
  created_at: string;
}

export const dashboardService = {
  // 获取仪表盘统计数据 - 使用新的API端点，支持门店筛选
  getStats: (storeId?: number): Promise<DashboardStats> => {
    const params = storeId ? { store_id: storeId } : {};
    return request.get<DashboardStats>('/v1/stats/dashboard', { params });
  },

  // 获取最近录音记录 - 从录音列表API获取，支持门店筛选
  getRecentRecordings: (limit: number = 5, storeId?: number): Promise<RecentRecording[]> => {
    const params = new URLSearchParams({
      limit: limit.toString(),
      offset: '0'
    });
    
    if (storeId !== undefined) {
      params.append('store_id', storeId.toString());
    }
    
    return request.get<{ items: RecentRecording[] }>(`/v1/recordings?${params.toString()}`)
      .then(response => {
        // response 已经是 result 部分，直接访问 items
        return response?.items || [];
      });
  },

  // 获取系统活动记录 - 暂时使用模拟数据
  getSystemActivities: (_limit: number = 5): Promise<SystemActivity[]> => {
    // 这里可以后续添加真实的系统活动API
    return Promise.resolve([
      {
        id: 1,
        type: 'recording',
        message: '系统正在处理录音数据',
        created_at: new Date().toISOString(),
      },
      {
        id: 2,
        type: 'store',
        message: '门店数据同步完成',
        created_at: new Date(Date.now() - 3600000).toISOString(),
      },
      {
        id: 3,
        type: 'user',
        message: '用户登录系统',
        created_at: new Date(Date.now() - 7200000).toISOString(),
      },
    ]);
  },

  // 获取录音增长趋势数据 - 从统计数据中提取
  getRecordingTrends: (): Promise<WeeklyTrendItem[]> => {
    return dashboardService.getStats().then(stats => stats.weekly_record_trend);
  },

  // 获取设备活跃度数据 - 从统计数据中提取
  getDeviceActivity: (): Promise<{ active: number; total: number }> => {
    return dashboardService.getStats().then(stats => ({
      active: stats.today_active_devices,
      total: stats.total_devices,
    }));
  },
};
