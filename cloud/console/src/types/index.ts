// 通用响应类型
export interface ApiResponse<T = any> {
  code: number;
  result: T;
  message?: string;
}

// 分页响应类型
export interface PaginatedResponse<T> {
  total: number;
  items: T[];
  limit: number;
  offset: number;
}

// 录音记录类型
export interface Recording {
  id: number;
  session_id: string;
  audio_id: string;
  mac_address: string;
  speaker_id: string;
  speaker_name: string;
  text: string;
  status: number; // 0=中间结果，1=最终结果
  created_at: number;
  device_time: number;
  play_url: string;
}

// 门店类型
export interface Store {
  id: number;
  name: string;
  code: string;
  address: string;
  contact: string;
  phone: string;
  status: number;
  created_at?: number;
  updated_at?: number;
}

// 点位类型
export interface Location {
  id: number;
  store_id: number;
  name: string;
  code: string;
  description: string;
  status: number;
  created_at?: number;
  updated_at?: number;
}

// 设备类型 - 根据API文档更新
export interface Device {
  id: number;
  mac_address: string;
  name: string;
  version: string;
  ip_address?: string; // IP地址
  cpu_usage_percent: number;
  memory_used_bytes: number;
  memory_total_bytes: number; // 内存总量
  disk_used_bytes: number;
  disk_total_bytes: number; // 磁盘总量
  swap_used_bytes: number;
  swap_total_bytes: number; // 交换分区总量
  location_id?: number;
  store_id?: number;
  status?: number; // 在线状态，可能不在API响应中
  created_at: number;
  updated_at: number;
}

// 用户类型
export interface User {
  id: number;
  username: string;
  created_at: number;
  updated_at: number;
}

// 登录响应类型
export interface LoginResponse {
  user: User;
  token: string;
}

// 分页查询参数
export interface PaginationParams {
  offset?: number;
  limit?: number;
}

// 录音查询参数
export interface RecordingQueryParams extends PaginationParams {
  mac_address?: string;
  mac_addresses?: string[]; // 多个MAC地址
  status?: number;
  start_time?: number; // 开始时间戳
  end_time?: number;   // 结束时间戳
  store_id?: number;   // 门店ID
  location_id?: number; // 点位ID
  device_id?: number;  // 设备ID
  type?: number;       // 录音类型
}

// 门店查询参数
export interface StoreQueryParams extends PaginationParams {
  name?: string;
  code?: string;
}

// 点位查询参数
export interface LocationQueryParams extends PaginationParams {
  store_id?: number;
  name?: string;
  code?: string;
}

// 设备查询参数
export interface DeviceQueryParams extends PaginationParams {
  name?: string;
  mac_address?: string;
  location_id?: number;
  store_id?: number;
  status?: 'online' | 'offline'; // 在线状态筛选
}

// 用户查询参数
export interface UserQueryParams extends PaginationParams {
  username?: string;
}

// 系统提示词类型
export interface SystemPrompt {
  id: number;
  name: string;
  role: string;
  content: string;
  tags: string;
  is_active: boolean;
  version: number;
  created_at: number;
  updated_at: number;
}

// API响应中的系统提示词类型（字段名大写）
export interface SystemPromptAPI {
  ID: number;
  Name: string;
  Role: string;
  Content: string;
  Tags: string;
  IsActive: boolean;
  Version: number;
  CreatedAt: number;
  UpdatedAt: number;
}

// 创建系统提示词请求
export interface CreateSystemPromptRequest {
  name: string;
  role?: string;
  content: string;
  tags?: string;
  is_active?: boolean;
}

// 更新系统提示词请求
export interface UpdateSystemPromptRequest {
  name?: string;
  role?: string;
  content?: string;
  tags?: string;
  is_active?: boolean;
}

// 系统提示词查询参数
export interface SystemPromptQueryParams extends PaginationParams {
  name?: string;
  role?: string;
  active?: boolean;
}

// 关键词类型
export interface Keyword {
  id: number;
  keyword: string;
  synonyms: string;
  mark_color: string;
  created_at: number;
  updated_at: number;
}

// 关键词查询参数
export interface KeywordQueryParams extends PaginationParams {
  keyword?: string;
  mark_color?: string;
}

// 创建关键词请求
export interface CreateKeywordRequest {
  keyword: string;
  synonyms: string;
  mark_color: string;
}

// 更新关键词请求
export interface UpdateKeywordRequest {
  keyword?: string;
  synonyms?: string;
  mark_color?: string;
}

// 仪表板统计数据类型 - 根据新API文档更新
export interface DashboardStats {
  total_records: number;        // 录音记录总数
  total_devices: number;        // 设备总数
  total_stores: number;         // 门店总数
  total_users: number;          // 用户总数
  weekly_record_trend: WeeklyTrendItem[]; // 本周录音记录增长趋势
  today_active_devices: number; // 今日活跃设备数
  today_hourly_distribution: HourlyDistributionItem[]; // 今日每小时采集分布
  today_keyword_triggers: number; // 今日关键词触发总次数
  today_keyword_matches: KeywordMatchItem[]; // 今日关键词匹配详情
}

// 周趋势数据项
export interface WeeklyTrendItem {
  date: string;    // 日期，格式 YYYY-MM-DD
  count: number;   // 当日录音记录数量
}

// 每小时分布数据项
export interface HourlyDistributionItem {
  hour: number;    // 小时，0-23
  count: number;   // 该小时的记录数量
}

// 关键词匹配数据项
export interface KeywordMatchItem {
  keyword_id: number;    // 关键词ID
  keyword: string;       // 关键词内容
  mark_color: string;    // 关键词标记颜色
  match_count: number;   // 匹配次数
  record_count: number;  // 涉及到的记录数
}

// 关键词匹配记录类型
export interface KeywordMatchRecord {
  id: number;
  recording_id: number;
  mac_address: string;
  keyword_id: number;
  keyword: string;
  matched_text: string;
  match_type: 'exact' | 'synonym';
  confidence: number;
  position: number;
  length: number;
  created_at: number;
  // 录音信息字段
  session_id: string;
  audio_id: string;
  speaker_id: string;
  speaker_name: string;
  text: string;
  device_time: number;
  status: number;
}

// 关键词匹配查询参数
export interface KeywordMatchQueryParams {
  offset?: number;
  limit?: number;
  mac_address?: string;
  keyword_id?: number;
  recording_id?: number;
  start_time?: number;
  end_time?: number;
  store_id?: number;  // 门店ID筛选
}
