import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Statistic, Progress, List, Avatar, Tag, Spin, message, Tooltip, Select } from 'antd';
import { Column } from '@ant-design/charts';
import {
  SoundOutlined,
  ShopOutlined,
  EnvironmentOutlined,
  UserOutlined,
  RiseOutlined,
  ClockCircleOutlined,
  FireOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { dashboardService, RecentRecording } from '@/services/dashboard';
import { DashboardStats, HourlyDistributionItem, KeywordMatchRecord } from '@/types';
import { deviceService } from '@/services/devices';
import { recordingService } from '@/services/recordings';
import { storeService } from '@/services/stores';
import { keywordService } from '@/services/keywords';
import { formatSmartTime, formatDate } from '@/utils/dateUtils';

const DashboardPage: React.FC = () => {
  const navigate = useNavigate();
  
  // 动态数据状态
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState<DashboardStats>({
    total_records: 0,
    total_devices: 0,
    total_stores: 0,
    total_users: 0,
    today_active_devices: 0,
    weekly_record_trend: [],
    today_hourly_distribution: [],
    today_keyword_triggers: 0,
    today_keyword_matches: [],
  });
  // 标记是否已经从 keyword-matches API 获取了数据
  const [hasKeywordStatsFromAPI, setHasKeywordStatsFromAPI] = useState(false);
  const [recentRecordings, setRecentRecordings] = useState<RecentRecording[]>([]);
  const [deviceNameMap, setDeviceNameMap] = useState<Map<string, string>>(new Map());

  // 关键词匹配记录相关状态
  const [keywordMatches, setKeywordMatches] = useState<KeywordMatchRecord[]>([]);
  const [selectedKeyword, setSelectedKeyword] = useState<{id: number, name: string, color: string} | null>({
    id: -1,
    name: '全部',
    color: 'blue'
  });
  const [keywordMatchesLoading, setKeywordMatchesLoading] = useState(false);



  // 门店选择状态
  const [selectedStore, setSelectedStore] = useState<number | 'all'>('all');
  const [storeOptions, setStoreOptions] = useState<{value: number | 'all', label: string}[]>([
    { value: 'all', label: '全部门店' }
  ]);
  const [loadingStores, setLoadingStores] = useState(false);

  // 处理每小时分布数据，确保24小时都有数据
  const processHourlyDistribution = (data: HourlyDistributionItem[]) => {    
    // 如果数据为空或不是数组，返回默认数据
    if (!data || !Array.isArray(data)) {
      const defaultData = [];
      for (let hour = 0; hour < 24; hour++) {
        defaultData.push({
          hour: `${hour.toString().padStart(2, '0')}:00`,
          count: 0,
          hourNum: hour,
          // 添加额外字段以确保tooltip正常工作
          time: `${hour.toString().padStart(2, '0')}:00`,
          value: 0
        });
      }
      return defaultData;
    }
    
    const processedData = [];
    for (let hour = 0; hour < 24; hour++) {
      const existingData = data.find(item => {
        // 确保item存在且hour字段存在
        return item && typeof item.hour === 'number' && item.hour === hour;
      });
      
      // 确保count字段存在且为数字
      const count = existingData && typeof existingData.count === 'number' 
        ? existingData.count 
        : 0;
        
      
      processedData.push({
        hour: `${hour.toString().padStart(2, '0')}:00`,
        count: count,
        hourNum: hour,
        // 添加额外字段以确保tooltip正常工作
        time: `${hour.toString().padStart(2, '0')}:00`,
        value: count
      });
    }
    return processedData;
  };

  // 生成测试数据（如果API数据为空）
  const generateTestData = () => {
    const testData = [];
    for (let hour = 0; hour < 24; hour++) {
      let count = 0;
      if (hour >= 9 && hour <= 18) {
        count = Math.floor(Math.random() * 8) + 2; // 2-9之间
      } else if (hour >= 7 && hour <= 8) {
        count = Math.floor(Math.random() * 3) + 1; // 1-3之间
      } else if (hour >= 19 && hour <= 22) {
        count = Math.floor(Math.random() * 2) + 1; // 1-2之间
      }
      
      testData.push({
        hour: `${hour.toString().padStart(2, '0')}:00`,
        count: count,
        hourNum: hour,
        // 添加额外字段以确保tooltip正常工作
        time: `${hour.toString().padStart(2, '0')}:00`,
        value: count
      });
    }
    return testData;
  };

  // 图表配置 - 使用API数据
  const chartData = stats.today_hourly_distribution && stats.today_hourly_distribution.length > 0 
    ? processHourlyDistribution(stats.today_hourly_distribution)
    : generateTestData();
    
  const collectionTrendConfig = {
    data: chartData,
    xField: 'hour',
    yField: 'count',
    colorField: 'count',
    color: '#8CC020',
    // 添加调试信息
    meta: {
      hour: {
        alias: '时间',
      },
      count: {
        alias: '记录数',
        formatter: (value: any) => {
          return value ? `${value}` : '0';
        }
      }
    },
    columnStyle: {
      radius: [4, 4, 0, 0],
    },
    xAxis: {
      title: {
        text: '小时',
        style: {
          fontSize: 12,
          fill: '#666',
        },
      },
      label: {
        style: {
          fontSize: 10,
          fill: '#666',
        },
        formatter: (text: string) => {
          const hour = parseInt(text.split(':')[0]);
          return hour % 2 === 0 ? hour.toString().padStart(2, '0') : '';
        },
        rotate: 0,
      },
    },
    yAxis: {
      title: {
        text: '记录数',
        style: {
          fontSize: 12,
          fill: '#666',
        },
      },
      label: {
        style: {
          fontSize: 10,
          fill: '#666',
        },
      },
    },
    tooltip: {
      showTitle: true,
      title: (title: string) => title,
      formatter: (datum: any) => {
        return {
          name: '记录数',
          value: datum.count !== undefined && datum.count !== null ? `${datum.count} 条` : '0 条',
        };
      },
    },
    animation: {
      appear: {
        animation: 'scale-in-y',
        duration: 1000,
      },
    },
  };

  // 获取门店列表
  const fetchStores = async () => {
    try {
      setLoadingStores(true);
      const response = await storeService.getStores({ limit: 100 }); // 获取最多100个门店
      const options: {value: number | 'all', label: string}[] = [
        { value: 'all', label: '全部门店' },
        ...response.items.map(store => ({
          value: store.id as number,
          label: store.name
        }))
      ];
      setStoreOptions(options);
    } catch (error) {
      message.error('获取门店列表失败');
    } finally {
      setLoadingStores(false);
    }
  };

  // 获取仪表盘数据
  const fetchDashboardData = async (storeId?: number) => {
    try {
      setLoading(true);
      
      // 并行获取所有数据
      const [statsData, recordingsData] = await Promise.all([
        dashboardService.getStats(storeId),
        dashboardService.getRecentRecordings(5, storeId),
      ]);

      // 获取设备名称映射
      await fetchDeviceNameMap();

      // 更新统计数据，但不覆盖关键词数据（这些数据由 fetchKeywordStatistics 管理）
      setStats(prev => ({
        ...statsData,
        // 如果已经从 keyword-matches API 获取了数据，保留它
        today_keyword_triggers: hasKeywordStatsFromAPI ? prev.today_keyword_triggers : statsData.today_keyword_triggers,
        today_keyword_matches: hasKeywordStatsFromAPI ? prev.today_keyword_matches : statsData.today_keyword_matches
      }));
      setRecentRecordings(Array.isArray(recordingsData) ? recordingsData : []);
    } catch (error) {
      message.error('获取仪表盘数据失败');
      
      // 设置默认数据，防止页面崩溃
      setStats({
        total_records: 0,
        total_devices: 0,
        total_stores: 0,
        total_users: 0,
        today_active_devices: 0,
        weekly_record_trend: [],
        today_hourly_distribution: [],
        today_keyword_triggers: 0,
        today_keyword_matches: [],
      });
      setRecentRecordings([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const loadInitialData = async () => {
      fetchStores(); // 获取门店列表
      await fetchDashboardData();
      // 在dashboard数据加载完成后，再获取关键词统计数据
      await fetchKeywordStatistics();
    };
    loadInitialData();
  }, []);

  // 门店筛选变化时重新获取数据
  useEffect(() => {
    const loadStoreData = async () => {
      if (selectedStore !== 'all') {
        await fetchDashboardData(selectedStore);
      } else {
        await fetchDashboardData();
      }
      // 门店变化时，重新获取关键词统计
      await fetchKeywordStatistics();
    };
    loadStoreData();
  }, [selectedStore]);

  // 定时刷新：每10秒刷新一次最近记录
  useEffect(() => {
    let isCancelled = false;

    const fetchRecent = async () => {
      try {
        const storeId = selectedStore !== 'all' ? selectedStore : undefined;
        const data = await dashboardService.getRecentRecordings(5, storeId);
        if (!isCancelled) {
          setRecentRecordings(Array.isArray(data) ? data : []);
        }
      } catch (err) {
        // 避免频繁打扰用户，轮询错误不显示
      }
    };

    const timer = setInterval(fetchRecent, 10000);
    // 立即执行一次，避免等待首个10s
    fetchRecent();

    return () => {
      isCancelled = true;
      clearInterval(timer);
    };
  }, [selectedStore]);

  // 获取所有设备信息用于MAC地址到名称的映射
  const fetchDeviceNameMap = async () => {
    try {
      const nameMap = new Map<string, string>();
      let offset = 0;
      const limit = 100;
      let hasMore = true;

      while (hasMore) {
        const response = await deviceService.getDevices({ offset, limit });
        
        response.items.forEach(device => {
          nameMap.set(device.mac_address, device.name);
        });

        // 检查是否还有更多数据
        hasMore = response.items.length === limit;
        offset += limit;
      }

      setDeviceNameMap(nameMap);
    } catch (error) {
      // 获取设备映射失败，静默处理
    }
  };

  // 根据MAC地址获取设备名称
  const getDeviceNameByMac = (macAddress: string): string => {
    return deviceNameMap.get(macAddress) || macAddress;
  };

  // 获取并统计关键词数据
  const fetchKeywordStatistics = async () => {
    try {
      setKeywordMatchesLoading(true);
      const storeId = selectedStore !== 'all' ? selectedStore : undefined;
      
      // 获取当天的时间范围（00:00:00 ~ 23:59:59）
      const today = new Date();
      const startOfDay = new Date(today.getFullYear(), today.getMonth(), today.getDate(), 0, 0, 0, 0);
      const endOfDay = new Date(today.getFullYear(), today.getMonth(), today.getDate(), 23, 59, 59, 999);
      
      // 构建查询参数
      const params: any = {
        start_time: startOfDay.getTime(),
        end_time: endOfDay.getTime(),
        limit: 1000, // 获取足够多的记录用于统计
        offset: 0
      };
      
      // 如果选择了特定门店，添加 store_id 参数
      if (storeId) {
        params.store_id = storeId;
      }
      
      const response = await recordingService.getKeywordMatches(params);
      
      // 获取所有关键词信息（用于获取颜色）
      const keywordsResponse = await keywordService.getKeywords({ limit: 100 });
      
      const keywordColorMap = new Map<number, string>();
      keywordsResponse.items.forEach(kw => {
        keywordColorMap.set(kw.id, kw.mark_color);
      });
      
      // 更新总数
      setStats(prev => ({
        ...prev,
        today_keyword_triggers: response.total
      }));
      
      // 按 keyword_id 分组统计
      const keywordMap = new Map<number, any>();
      
      response.items.forEach(item => {
        if (!keywordMap.has(item.keyword_id)) {
          keywordMap.set(item.keyword_id, {
            keyword_id: item.keyword_id,
            keyword: item.keyword,
            mark_color: keywordColorMap.get(item.keyword_id) || '#8CC020', // 使用关键词的颜色
            match_count: 0
          });
        }
        keywordMap.get(item.keyword_id)!.match_count++;
      });
      
      // 获取有颜色但没有匹配的关键词（count=0）
      keywordsResponse.items.forEach(kw => {
        if (!keywordMap.has(kw.id)) {
          keywordMap.set(kw.id, {
            keyword_id: kw.id,
            keyword: kw.keyword,
            mark_color: kw.mark_color,
            match_count: 0
          });
        } else {
          // 如果已经存在，确保颜色是正确的
          const existing = keywordMap.get(kw.id);
          existing.mark_color = kw.mark_color;
          keywordMap.set(kw.id, existing);
        }
      });
      
      // 转换为数组并更新状态
      const keywordStats = Array.from(keywordMap.values());
      
      setStats(prev => ({
        ...prev,
        today_keyword_matches: keywordStats
      }));
      
      // 标记已经从 API 获取了关键词统计数据
      setHasKeywordStatsFromAPI(true);
      
      // 更新匹配记录列表（默认显示全部）
      setKeywordMatches(response.items || []);
      
    } catch (error) {
      message.error('获取关键词统计失败');
    } finally {
      setKeywordMatchesLoading(false);
    }
  };


  // 获取关键词匹配记录
  const fetchKeywordMatches = async (keywordId: number) => {
    try {
      setKeywordMatchesLoading(true);
      const storeId = selectedStore !== 'all' ? selectedStore : undefined;
      
      // 获取当天的时间范围（00:00:00 ~ 23:59:59）
      const today = new Date();
      const startOfDay = new Date(today.getFullYear(), today.getMonth(), today.getDate(), 0, 0, 0, 0);
      const endOfDay = new Date(today.getFullYear(), today.getMonth(), today.getDate(), 23, 59, 59, 999);
      
      // 构建查询参数
      const params: any = {
        keyword_id: keywordId,
        start_time: startOfDay.getTime(), // 当天00:00:00的时间戳
        end_time: endOfDay.getTime(),     // 当天23:59:59的时间戳
        limit: 100,
        offset: 0
      };
      
      // 如果选择了特定门店，需要根据门店获取设备列表，然后按设备筛选
      if (storeId) {
        try {
          // 获取门店下的所有设备
          const devicesResponse = await storeService.getStoreDevices(storeId, { limit: 100 });
          const macAddresses = devicesResponse.items.map(device => device.mac_address);
          
          if (macAddresses.length > 0) {
            // 使用第一个设备的MAC地址作为示例，实际应该支持多个MAC地址查询
            params.mac_address = macAddresses[0];
          }
        } catch (deviceError) {
          // 获取门店设备失败，静默处理
        }
      }
      
      const response = await recordingService.getKeywordMatches(params);
      setKeywordMatches(response.items || []);
    } catch (error) {
      message.error('获取关键词匹配记录失败');
      setKeywordMatches([]);
    } finally {
      setKeywordMatchesLoading(false);
    }
  };


  // 处理MAC地址点击，跳转到门店管理页面
  const handleMacAddressClick = (macAddress: string) => {
    navigate('/stores', { 
      state: { 
        macAddress: macAddress,
        activeTab: 'recordings' 
      } 
    });
  };

  // 处理关键词点击，显示匹配记录详情
  const handleKeywordClick = async (keywordMatch: any) => {
    setSelectedKeyword({
      id: keywordMatch.keyword_id,
      name: keywordMatch.keyword,
      color: keywordMatch.mark_color
    });
    await fetchKeywordMatches(keywordMatch.keyword_id);
  };

  // 截断文本到指定长度，超过则用...表示
  const truncateText = (text: string, maxLength: number = 5): string => {
    if (text.length <= maxLength) {
      return text;
    }
    return text.substring(0, maxLength) + '...';
  };



  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-96">
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* 页面标题和门店选择 */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold text-primary-700">
            仪表盘
          </h1>
          <div className="flex items-center gap-4">
            <span className="text-gray-600">选择门店：</span>
            <Select
              value={selectedStore}
              onChange={(value) => setSelectedStore(value)}
              options={storeOptions}
              style={{ width: 200 }}
              placeholder="请选择门店"
              loading={loadingStores}
              notFoundContent={loadingStores ? <Spin size="small" /> : '暂无数据'}
            />
          </div>
        </div>
      </div>

      {/* 统计卡片 */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} md={6}>
          <Card className="card-primary">
            <Statistic
              title="总记录数"
              value={stats.total_records}
              valueStyle={{ color: '#8CC020' }}
              prefix={<SoundOutlined className="text-2xl" />}
            />
            <div className="mt-2 text-sm text-gray-500">
              本周趋势 <span className="text-green-500">+{stats.weekly_record_trend && Array.isArray(stats.weekly_record_trend) && stats.weekly_record_trend.length > 0 ? stats.weekly_record_trend[stats.weekly_record_trend.length - 1].count : 0}</span>
            </div>
          </Card>
        </Col>
        
        <Col xs={24} sm={12} md={6}>
          <Card className="card-primary">
            <Statistic
              title="点位数量"
              value={stats.total_stores}
              valueStyle={{ color: '#8CC020' }}
              prefix={<ShopOutlined className="text-2xl" />}
            />
            <div className="mt-2 text-sm text-gray-500">
              门店下所有点位
            </div>
          </Card>
        </Col>
        
        <Col xs={24} sm={12} md={6}>
          <Card className="card-primary">
            <Statistic
              title="设备在线率"
              value={`${stats.total_devices > 0 ? Math.round((stats.today_active_devices / stats.total_devices) * 100) : 0}%`}
              valueStyle={{ color: '#8CC020' }}
              prefix={<EnvironmentOutlined className="text-2xl" />}
            />
            <div className="mt-2 text-sm text-gray-500">
              {stats.today_active_devices}/{stats.total_devices} 台设备在线
            </div>
          </Card>
        </Col>
        
        <Col xs={24} sm={12} md={6}>
          <Card className="card-primary">
            <Statistic
              title="关键词触发"
              value={stats.today_keyword_triggers}
              valueStyle={{ color: '#8CC020' }}
              prefix={<FireOutlined className="text-2xl" />}
            />
            <div className="mt-2 text-sm text-gray-500">
              今日触发总次数
            </div>
          </Card>
        </Col>
      </Row>

      {/* 主要内容区域 - 左右分栏布局 */}
      <Row gutter={[16, 16]}>
        {/* 左侧主区域 - 占5/10 */}
        <Col xs={24} xl={12}>
          <div className="space-y-6">
            {/* 今日采集趋势图表 */}
            <Card 
              title="今日采集趋势" 
              className="card-primary"
              extra={<RiseOutlined className="text-primary-500" />}
            >
              <div className="h-64">
                {chartData && chartData.length > 0 ? (
                  <Column {...collectionTrendConfig} />
                ) : (
                  <div className="flex items-center justify-center h-full text-gray-500">
                    暂无数据
                  </div>
                )}
              </div>
            </Card>

            {/* 关键词热度分析 */}
            <Card 
              title="关键词热度分析" 
              className="card-primary"
              extra={<FireOutlined className="text-primary-500" />}
            >
              <div className="space-y-4">
                <div className="flex gap-2 sm:gap-3 flex-wrap">
                  <div
                    className={`
                      px-3 sm:px-4 py-1.5 rounded-lg cursor-pointer
                      text-sm sm:text-base font-medium
                      transition-all duration-300 ease-in-out
                      transform hover:scale-105
                      ${selectedKeyword?.id === -1 
                        ? 'bg-primary-500 text-white shadow-lg scale-105' 
                        : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                      }
                    `}
                    onClick={async () => {
                      // 点击"全部"时，获取所有关键词的匹配记录
                      setSelectedKeyword({
                        id: -1,
                        name: '全部',
                        color: '#8CC020'
                      });
                      
                      // 直接从已经获取的统计数据中取出所有记录
                      try {
                        setKeywordMatchesLoading(true);
                        const storeId = selectedStore !== 'all' ? selectedStore : undefined;
                        
                        // 获取当天的时间范围（00:00:00 ~ 23:59:59）
                        const today = new Date();
                        const startOfDay = new Date(today.getFullYear(), today.getMonth(), today.getDate(), 0, 0, 0, 0);
                        const endOfDay = new Date(today.getFullYear(), today.getMonth(), today.getDate(), 23, 59, 59, 999);
                        
                        const params: any = {
                          start_time: startOfDay.getTime(),
                          end_time: endOfDay.getTime(),
                          limit: 1000,
                          offset: 0
                        };
                        
                        // 如果选择了特定门店，添加 store_id 参数
                        if (storeId) {
                          params.store_id = storeId;
                        }
                        
                        const response = await recordingService.getKeywordMatches(params);
                        setKeywordMatches(response.items || []);
                      } catch (error) {
                        message.error('获取关键词匹配记录失败');
                        setKeywordMatches([]);
                      } finally {
                        setKeywordMatchesLoading(false);
                      }
                    }}
                  >
                    全部 ({stats.today_keyword_triggers})
                  </div>
                  {stats.today_keyword_matches && Array.isArray(stats.today_keyword_matches) && stats.today_keyword_matches.map((keywordMatch) => (
                    <Tooltip 
                      key={keywordMatch.keyword_id}
                      title={`${keywordMatch.keyword} (${keywordMatch.match_count})`}
                      placement="top"
                    >
                      <div
                        className={`
                          px-3 sm:px-4 py-1.5 rounded-lg cursor-pointer
                          text-sm sm:text-base font-medium text-white
                          transition-all duration-300 ease-in-out
                          transform hover:scale-105
                          ${selectedKeyword?.id === keywordMatch.keyword_id
                            ? 'shadow-lg scale-105'
                            : 'opacity-80 hover:opacity-100 hover:shadow-md'
                          }
                        `}
                        style={{
                          backgroundColor: keywordMatch.mark_color,
                        }}
                        onClick={() => handleKeywordClick(keywordMatch)}
                      >
                        {truncateText(keywordMatch.keyword)} ({keywordMatch.match_count})
                      </div>
                    </Tooltip>
                  ))}
                </div>
                
                {/* 关键词匹配记录展示区域 */}
                {selectedKeyword && (
                  <div 
                    className="mt-4 border-t pt-4 animate-fade-in"
                    style={{
                      animation: 'fadeIn 0.3s ease-in-out'
                    }}
                  >
                    <div className="flex items-center gap-2 mb-3">
                      <div 
                        className="w-3 h-3 rounded-full transition-all duration-300"
                        style={{ 
                          backgroundColor: selectedKeyword.id === -1 ? '#8CC020' : selectedKeyword.color,
                          boxShadow: '0 0 0 4px ' + (selectedKeyword.id === -1 ? '#8CC02020' : selectedKeyword.color + '20')
                        }}
                      />
                      <span className="font-medium text-gray-700">
                        {selectedKeyword.name === '全部' ? '今日所有关键词相关对话' : '相关对话内容'}
                      </span>
                      {keywordMatchesLoading && <Spin size="small" />}
                    </div>
                    
                    {/* 限制高度并显示滚动条，超过3条时显示滚动条 */}
                    <div 
                      className={keywordMatches.length > 3 ? 'custom-scrollbar' : ''}
                      style={{ 
                        maxHeight: keywordMatches.length > 3 ? '240px' : 'auto',
                        overflowY: keywordMatches.length > 3 ? 'auto' : 'visible',
                        overflowX: 'hidden',
                        paddingRight: keywordMatches.length > 3 ? '2px' : '0'
                      }}
                    >
                      <List
                        dataSource={keywordMatches}
                        loading={keywordMatchesLoading}
                        locale={{ emptyText: '暂无匹配记录' }}
                        size="small"
                        renderItem={(match) => (
                          <List.Item className="px-0 py-2">
                            <List.Item.Meta
                              avatar={
                                <Avatar 
                                  icon={<UserOutlined />} 
                                  className="bg-primary-100 text-primary-600"
                                  size="small"
                                />
                              }
                              title={
                                <div className="flex items-center justify-between">
                                  <span className="font-medium text-primary-700 text-sm">
                                    {match.speaker_name}
                                  </span>
                                  <Tag 
                                    color="blue" 
                                    className="cursor-pointer hover:opacity-80 transition-opacity text-xs"
                                    onClick={() => handleMacAddressClick(match.mac_address)}
                                  >
                                    {getDeviceNameByMac(match.mac_address)}
                                  </Tag>
                                </div>
                              }
                              description={
                                <div className="space-y-1">
                                  <div className="text-gray-600 text-sm line-clamp-2">{match.text}</div>
                                  <div className="flex items-center justify-between">
                                    <div className="text-xs text-gray-400">
                                      <Tooltip title={formatDate(match.created_at)}>
                                        {formatSmartTime(match.created_at)}
                                      </Tooltip>
                                    </div>
                                    <div className="text-xs text-gray-500">
                                      匹配: "{match.matched_text}" 
                                      ({match.match_type === 'exact' ? '精确' : '近义词'})
                                    </div>
                                  </div>
                                </div>
                              }
                            />
                          </List.Item>
                        )}
                      />
                    </div>
                  </div>
                )}
              </div>
            </Card>
          </div>
        </Col>

        {/* 右侧区域 - 占5/10 */}
        <Col xs={24} xl={12}>
          <div className="space-y-6">
            {/* 今日活跃设备 */}
            <Card 
              title="今日活跃设备" 
              className="card-primary"
              extra={<ClockCircleOutlined className="text-primary-500" />}
            >
              <div className="space-y-4">
                <div className="flex justify-between items-center">
                  <span className="text-gray-600">活跃设备数</span>
                  <span className="text-lg font-semibold text-primary-600">
                    {stats.today_active_devices}
                  </span>
                </div>
                <Progress 
                  percent={stats.total_devices > 0 ? Math.min(100, (stats.today_active_devices / stats.total_devices) * 100) : 0} 
                  strokeColor="#8CC020"
                  showInfo={false}
                  strokeWidth={8}
                />
                <div className="text-sm text-gray-500">
                  今日有 {stats.today_active_devices} 台设备产生记录
                </div>
              </div>
            </Card>

            {/* 最近记录 */}
            <Card 
              title="最近记录" 
              className="card-primary"
              extra={<SoundOutlined className="text-primary-500" />}
            >
              <List
                dataSource={Array.isArray(recentRecordings) ? recentRecordings.slice(0, 5) : []}
                renderItem={(item) => (
                  <List.Item className="px-0">
                    <List.Item.Meta
                      avatar={
                        <Avatar 
                          icon={<UserOutlined />} 
                          className="bg-primary-100 text-primary-600"
                        />
                      }
                      title={
                        <div className="flex items-center justify-between">
                          <span className="font-medium text-primary-700 text-sm">
                            {item.speaker_name}
                          </span>
                          <Tag 
                            color="blue" 
                            className="cursor-pointer hover:opacity-80 transition-opacity text-xs"
                            onClick={() => handleMacAddressClick(item.mac_address)}
                          >
                            {getDeviceNameByMac(item.mac_address)}
                          </Tag>
                        </div>
                      }
                      description={
                        <div className="space-y-1">
                          <div className="text-gray-600 text-sm">{item.text}</div>
                          <div className="text-xs text-gray-400">
                            <Tooltip title={formatDate(item.created_at)}>
                              {formatSmartTime(item.created_at)}
                            </Tooltip>
                          </div>
                        </div>
                      }
                    />
                  </List.Item>
                )}
              />
            </Card>
          </div>
        </Col>
      </Row>

    </div>
  );
};

export default DashboardPage;
