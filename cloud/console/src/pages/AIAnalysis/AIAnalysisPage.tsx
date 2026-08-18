import React, { useState, useRef, useEffect, useCallback } from 'react';
import { 
  Card, 
  Input, 
  Button, 
  Space, 
  Avatar, 
  message,
  Spin,
  Empty,
  Tooltip,
  Tag,
  Select,
  Row,
  Col,
  Pagination,
  List,
  DatePicker
} from 'antd';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeHighlight from 'rehype-highlight';
import 'highlight.js/styles/github.css';
import {
  SendOutlined,
  UserOutlined,
  RobotOutlined,
  PlusOutlined,
  SoundOutlined,
  SearchOutlined,
  UploadOutlined,
  FileTextOutlined,
  ExclamationCircleOutlined,
  InfoCircleOutlined,
  RedoOutlined,
} from '@ant-design/icons';
import { chatService, StreamingChatResponse } from '@/services/chat';
import { recordingService } from '@/services/recordings';
import { storeService } from '@/services/stores';
import { locationService } from '@/services/locations';
import { deviceService } from '@/services/devices';
import { systemPromptService } from '@/services/systemPrompts';
import { Recording, RecordingQueryParams, Store, Location, Device, SystemPrompt } from '@/types';
import SystemPromptSelector from '@/components/SystemPromptManager/SystemPromptSelector';
import SystemPromptManagerModal from '@/components/SystemPromptManager/SystemPromptManagerModal';
import { formatFullDateTime } from '@/utils/dateUtils';
import dayjs from 'dayjs';

const { TextArea } = Input;

interface ChatMessage {
  id: string;
  type: 'user' | 'ai';
  content: string;
  timestamp: number;
  loading?: boolean;
  streaming?: boolean;
  error?: boolean;
  errorDetails?: string;
  originalQuery?: string; // 保存原始查询，用于重试
}

const AIAnalysisPage: React.FC = () => {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [inputValue, setInputValue] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [currentStreamingResponse, setCurrentStreamingResponse] = useState<StreamingChatResponse | null>(null);
  const [currentSessionId, setCurrentSessionId] = useState<string | null>(null);
  const [expandedMessages, setExpandedMessages] = useState<Set<string>>(new Set());
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // 录音管理相关状态
  const [recordings, setRecordings] = useState<Recording[]>([]);
  const [recordingsLoading, setRecordingsLoading] = useState(false);
  const [recordingsTotal, setRecordingsTotal] = useState(0);
  const [recordingsPage, setRecordingsPage] = useState(1);
  const [recordingsPageSize, setRecordingsPageSize] = useState(2000);


  const [selectedRecordings, setSelectedRecordings] = useState<Recording[]>([]);
  
  // 新增筛选状态
  const [stores, setStores] = useState<Store[]>([]);
  const [locations, setLocations] = useState<Location[]>([]);
  const [devices, setDevices] = useState<Device[]>([]);
  const [selectedStoreId, setSelectedStoreId] = useState<number | undefined>();
  const [selectedLocationId, setSelectedLocationId] = useState<number | undefined>();
  const [selectedDeviceIds, setSelectedDeviceIds] = useState<number[]>([]);
  const [startTime, setStartTime] = useState<dayjs.Dayjs | null>(null);
  const [endTime, setEndTime] = useState<dayjs.Dayjs | null>(null);
  const [showCustomTime, setShowCustomTime] = useState(false);
  
  // 设备MAC地址到名称的映射
  const [deviceNameMap, setDeviceNameMap] = useState<Map<string, string>>(new Map());

  // 系统提示词相关状态
  const [selectedSystemPrompt, setSelectedSystemPrompt] = useState<SystemPrompt | null>(null);
  const [systemPromptModalVisible, setSystemPromptModalVisible] = useState(false);

  // 自动滚动到底部
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // 获取录音数据
  useEffect(() => {
    fetchRecordings();
  }, [recordingsPage, recordingsPageSize, selectedStoreId, selectedLocationId, selectedDeviceIds, startTime, endTime]);

  // 获取所有设备信息用于名称匹配
  useEffect(() => {
    fetchAllDevicesForMapping();
  }, []);

  // 获取默认系统提示词
  useEffect(() => {
    const fetchDefaultSystemPrompt = async () => {
      try {
        const defaultPrompt = await systemPromptService.getDefaultSystemPrompt();
        if (defaultPrompt) {
          setSelectedSystemPrompt(defaultPrompt);
        }
      } catch (error) {
        console.error('获取默认系统提示词失败:', error);
      }
    };
    fetchDefaultSystemPrompt();
  }, []);

  // 获取门店数据
  useEffect(() => {
    fetchStores();
  }, []);

  // 当选择门店时，获取对应的点位
  useEffect(() => {
    if (selectedStoreId) {
      fetchLocationsByStore(selectedStoreId);
    } else {
      setLocations([]);
      setSelectedLocationId(undefined);
      setDevices([]);
      setSelectedDeviceIds([]);
    }
  }, [selectedStoreId]);

  // 当选择点位时，获取对应的设备
  useEffect(() => {
    if (selectedLocationId) {
      fetchDevicesByLocation(selectedLocationId);
    } else {
      setDevices([]);
      setSelectedDeviceIds([]);
    }
  }, [selectedLocationId]);



  // 发送消息
  const handleSendMessage = async () => {
    if (!inputValue.trim()) {
      message.warning('请输入消息内容');
      return;
    }

    // 检查字符数限制
    if (inputValue.length > 5000000) {
      message.error('输入内容超过500万字符限制，请减少内容后重试');
      return;
    }

    const userMessage: ChatMessage = {
      id: Date.now().toString(),
      type: 'user',
      content: inputValue.trim(),
      timestamp: Date.now(),
    };

    setMessages(prev => [...prev, userMessage]);
    const query = inputValue.trim();
    setInputValue('');
    setIsLoading(true);

    // 添加AI响应消息
    const aiMessageId = (Date.now() + 1).toString();
    const aiMessage: ChatMessage = {
      id: aiMessageId,
      type: 'ai',
      content: '',
      timestamp: Date.now(),
      streaming: true,
      originalQuery: query, // 保存原始查询用于重试
    };
    setMessages(prev => [...prev, aiMessage]);

    try {
      // 直接调用真实的流式API
      console.log('发送消息到AI服务:', query);
      await handleRealStreamingResponse(query, aiMessageId);
    } catch (error) {
      console.error('AI服务调用失败:', error);
      message.error('AI服务暂时不可用，请稍后重试');
      // 移除AI消息
      setMessages(prev => prev.filter(msg => msg.id !== aiMessageId));
    } finally {
      setIsLoading(false);
    }
  };

  // 处理真实的流式响应 (v2 API)
  const handleRealStreamingResponse = async (query: string, messageId: string) => {
    try {
      // 获取用户ID（可以从auth store获取，这里先用默认值）
      const userId = 'web-user'; // TODO: 从auth store获取真实用户ID
      
      // 使用v2 API，直接传递系统提示词ID
      const streamResponse = await chatService.sendStreamingMessage(
        query,
        userId,
        selectedSystemPrompt?.id,
        undefined // 不使用自定义系统提示词内容
      );
      setCurrentStreamingResponse(streamResponse);

      // 处理流式消息
      streamResponse.onMessage((content: string) => {
        // 更新会话ID（如果还没有的话）
        const sessionId = chatService.getCurrentSessionId();
        if (sessionId && !currentSessionId) {
          setCurrentSessionId(sessionId);
        }
        
        setMessages(prev => prev.map(msg => 
          msg.id === messageId 
            ? { ...msg, content: msg.content + content, streaming: true }
            : msg
        ));
      });

      // 处理完成
      streamResponse.onComplete(() => {
        setMessages(prev => prev.map(msg => 
          msg.id === messageId 
            ? { ...msg, streaming: false }
            : msg
        ));
        setCurrentStreamingResponse(null);
      });

      // 处理错误
      streamResponse.onError((error: Error) => {
        console.error('流式响应错误:', error);
        
        // 根据错误类型显示不同的提示信息
        let errorMessage = 'AI响应出错，请稍后重试';
        
        if (error.message.includes('token too long')) {
          errorMessage = '输入内容过长，请减少输入内容后重试';
        } else if (error.message.includes('bufio.Scanner')) {
          errorMessage = '数据传输异常，请检查网络连接或减少输入内容';
        } else if (error.message.includes('timeout')) {
          errorMessage = '请求超时，请检查网络连接后重试';
        } else if (error.message.includes('rate limit')) {
          errorMessage = '请求过于频繁，请稍后再试';
        }
        
        message.error(errorMessage);
        
        // 更新消息状态为错误状态，而不是直接删除
        setMessages(prev => prev.map(msg => 
          msg.id === messageId 
            ? { 
                ...msg, 
                content: errorMessage,
                streaming: false,
                error: true,
                errorDetails: error.message
              }
            : msg
        ));
        
        setCurrentStreamingResponse(null);
      });

    } catch (error) {
      console.error('启动流式响应失败:', error);
      throw error;
    }
  };




  // 新建会话
  const handleRestartChat = () => {
    // 如果有正在进行的流式响应，先取消它
    if (currentStreamingResponse) {
      currentStreamingResponse.cancel();
      setCurrentStreamingResponse(null);
    }
    
    // 重置会话
    chatService.resetSession();
    setCurrentSessionId(null);
    
    setMessages([]);
    setInputValue('');
    setIsLoading(false);
    message.success('已新建会话');
  };

  // 重试失败的消息
  const handleRetryMessage = async (messageId: string) => {
    const message = messages.find(msg => msg.id === messageId);
    if (!message || !message.originalQuery) {
      return;
    }

    // 重置消息状态
    setMessages(prev => prev.map(msg => 
      msg.id === messageId 
        ? { 
            ...msg, 
            content: '',
            streaming: true,
            error: false,
            errorDetails: undefined
          }
        : msg
    ));

    setIsLoading(true);

    try {
      // 重新发送消息
      console.log('重试发送消息到AI服务:', message.originalQuery);
      await handleRealStreamingResponse(message.originalQuery, messageId);
    } catch (error) {
      console.error('重试AI服务调用失败:', error);
      // 恢复错误状态
      setMessages(prev => prev.map(msg => 
        msg.id === messageId 
          ? { 
              ...msg, 
              content: '重试失败，请稍后再试',
              streaming: false,
              error: true,
              errorDetails: (error as Error).message
            }
          : msg
      ));
    } finally {
      setIsLoading(false);
    }
  };



  // 按Enter发送消息
  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      // 检查是否超过字符限制
      if (inputValue.length <= 5000000) {
        handleSendMessage();
      } else {
        message.error('输入内容超过500万字符限制，请减少内容后重试');
      }
    }
  };

  // 录音管理相关函数
  
  // 获取录音列表
  const fetchRecordings = useCallback(async () => {
    try {
      setRecordingsLoading(true);
      
      const params: RecordingQueryParams = {
        offset: (recordingsPage - 1) * recordingsPageSize,
        limit: recordingsPageSize,
      };
      
      if (selectedStoreId !== undefined) {
        params.store_id = selectedStoreId;
      }

      if (selectedLocationId !== undefined) {
        params.location_id = selectedLocationId;
      }

      // 如果选择了多个设备，将设备ID转换为MAC地址进行查询
      if (selectedDeviceIds.length > 0) {
        const selectedMacAddresses = devices
          .filter(device => selectedDeviceIds.includes(device.id))
          .map(device => device.mac_address);
        if (selectedMacAddresses.length > 0) {
          // 由于当前API不支持多个device_id，我们使用mac_address数组
          // 这里需要修改录音服务以支持多个MAC地址
          params.mac_addresses = selectedMacAddresses;
        }
      }

      if (startTime) {
        params.start_time = startTime.valueOf();
      }

      if (endTime) {
        params.end_time = endTime.valueOf();
      }
      
      const response = await recordingService.getRecordings(params);
      setRecordings(response.items);
      setRecordingsTotal(response.total);
      // 默认全选所有录音
      setSelectedRecordings(response.items);
    } catch (error) {
      message.error('获取录音记录失败');
    } finally {
      setRecordingsLoading(false);
    }
  }, [recordingsPage, recordingsPageSize, selectedStoreId, selectedLocationId, selectedDeviceIds, startTime, endTime, devices]);

  // 获取门店列表
  const fetchStores = useCallback(async () => {
    try {
      const response = await storeService.getStores({ limit: 100 });
      setStores(response.items);
    } catch (error) {
      message.error('获取门店列表失败');
    }
  }, []);

  // 根据门店获取点位列表
  const fetchLocationsByStore = useCallback(async (storeId: number) => {
    try {
      const response = await locationService.getLocations({ store_id: storeId, limit: 100 });
      setLocations(response.items);
    } catch (error) {
      message.error('获取点位列表失败');
    }
  }, []);

  // 根据点位获取设备列表
  const fetchDevicesByLocation = useCallback(async (locationId: number) => {
    try {
      const response = await deviceService.getDevices({ location_id: locationId, limit: 100 });
      setDevices(response.items);
    } catch (error) {
      message.error('获取设备列表失败');
    }
  }, []);

  // 获取所有设备信息用于MAC地址到名称的映射
  const fetchAllDevicesForMapping = useCallback(async () => {
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
      console.error('获取设备映射失败:', error);
    }
  }, []);

  // 根据MAC地址获取设备名称
  const getDeviceNameByMac = (macAddress: string): string => {
    return deviceNameMap.get(macAddress) || macAddress;
  };

  // 压缩用户名显示（用户_01 → 1）
  const compressUserName = (speakerName: string | null): string => {
    if (!speakerName) return '未知';
    
    // 匹配 "用户_XX" 格式
    const match = speakerName.match(/^用户_(\d+)$/);
    if (match) {
      return match[1]; // 返回数字部分
    }
    
    return speakerName; // 其他格式保持原样
  };

  // 处理录音分页
  const handleRecordingPageChange = (page: number, size: number) => {
    const limitedSize = Math.min(size, 100);
    setRecordingsPage(page);
    setRecordingsPageSize(limitedSize);
  };

  // 处理录音分页大小变化
  const handleRecordingPageSizeChange = (current: number, size: number) => {
    setRecordingsPage(current);
    setRecordingsPageSize(size);
  };

  // 搜索录音
  const handleRecordingSearch = () => {
    setRecordingsPage(1);
    fetchRecordings();
  };

  // 快速时间选择函数
  const handleQuickTimeSelect = (type: string) => {
    const now = dayjs();
    let start: dayjs.Dayjs;
    let end: dayjs.Dayjs = now;

    switch (type) {
      case '1hour':
        start = now.subtract(1, 'hour');
        break;
      case '4hours':
        start = now.subtract(4, 'hour');
        break;
      case 'today':
        start = now.startOf('day');
        break;
      case 'yesterday':
        start = now.subtract(1, 'day').startOf('day');
        end = now.subtract(1, 'day').endOf('day');
        break;
      case '7days':
        start = now.subtract(7, 'day').startOf('day');
        break;
      default:
        return;
    }

    setStartTime(start);
    setEndTime(end);
    setShowCustomTime(false);
  };

  // 显示自定义时间选择
  const handleShowCustomTime = () => {
    setShowCustomTime(true);
  };

  // 清除时间选择
  const handleClearTime = () => {
    setStartTime(null);
    setEndTime(null);
    setShowCustomTime(false);
  };



  // 发送录音内容给AI分析
  const handleSendRecordingToAI = (recording: Recording) => {
    const recordingInfo = `请分析以下录音内容：${recording.speaker_name || '未知'}:${recording.text}

请提供关键信息和见解分析。`;
    
    setInputValue(recordingInfo);
    message.success('录音内容已添加到输入框，您可以编辑后发送给AI分析');
  };

  // 批量发送选中录音给AI
  const handleSendSelectedRecordingsToAI = () => {
    if (selectedRecordings.length === 0) {
      message.warning('请先选择要分析的录音记录');
      return;
    }

    const recordingsInfo = selectedRecordings.map((recording) => 
      `${compressUserName(recording.speaker_name)}:${recording.text}`
    ).join(',');

    // 获取时间范围
    const timestamps = selectedRecordings.map(r => r.created_at).sort((a, b) => a - b);
    const startTime = formatFullDateTime(timestamps[0]);
    const endTime = formatFullDateTime(timestamps[timestamps.length - 1]);
    
    const batchAnalysisRequest = `请分析${startTime}～${endTime}时间的数据：
${recordingsInfo}

请提供关键信息、主要话题和重要见解的分析。`;

    setInputValue(batchAnalysisRequest);
    setSelectedRecordings([]);
    message.success(`已将${selectedRecordings.length}条录音添加到输入框，您可以编辑后发送给AI分析`);
  };

  // 切换消息展开状态
  const toggleMessageExpansion = (messageId: string) => {
    setExpandedMessages(prev => {
      const newSet = new Set(prev);
      if (newSet.has(messageId)) {
        newSet.delete(messageId);
      } else {
        newSet.add(messageId);
      }
      return newSet;
    });
  };

  // 处理系统提示词选择
  const handleSystemPromptChange = (prompt: SystemPrompt | null) => {
    setSelectedSystemPrompt(prompt);
    if (prompt) {
      message.success(`已切换到角色: ${prompt.name}`);
    } else {
      message.info('已清除AI角色选择');
    }
  };

  // 处理系统提示词管理弹窗
  const handleSystemPromptManage = () => {
    setSystemPromptModalVisible(true);
  };

  const handleSystemPromptModalClose = () => {
    setSystemPromptModalVisible(false);
  };

  return (
    <div className="h-full flex flex-col space-y-6" style={{ maxHeight: '100vh', overflow: 'hidden' }}>
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-primary-700 mb-2">AI智能分析 & 录音管理</h1>
          <p className="text-gray-600">与AI助手对话分析录音内容，获取数据洞察</p>
          {currentSessionId && (
            <p className="text-xs text-gray-500 mt-1">
              会话ID: {currentSessionId}
            </p>
          )}
        </div>
        
        <Space>
          <Tooltip title="新建会话">
            <Button 
              icon={<PlusOutlined />} 
              onClick={handleRestartChat}
              type="primary"
              className="btn-primary"
              disabled={isLoading}
            >
              新建会话
            </Button>
          </Tooltip>
        </Space>
      </div>

      {/* 系统提示词选择器 - 紧凑版本 */}
      <div className="mb-2">
        <SystemPromptSelector
          selectedPrompt={selectedSystemPrompt}
          onPromptChange={handleSystemPromptChange}
          onManageClick={handleSystemPromptManage}
        />
      </div>

      {/* 主内容区域：左侧录音管理 + 右侧聊天 */}
      <div className="flex space-x-6 flex-1" style={{ minHeight: '600px', maxHeight: 'calc(100vh - 120px)' }}>
        {/* 左侧录音管理区域 - 进一步增加高度 */}
        <Card className="w-96 flex flex-col card-primary" styles={{ body: { padding: 0, height: '100%', display: 'flex', flexDirection: 'column', maxHeight: '900px' } }}>
          <div className="p-3 border-b border-gray-200">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-base font-semibold text-primary-700 flex items-center">
                <SoundOutlined className="mr-2" />
                录音记录
              </h3>
              <div className="flex items-center space-x-2">
                <Button
                  type="link"
                  size="small"
                  onClick={() => {
                    const allSelected = selectedRecordings.length === recordings.length;
                    setSelectedRecordings(allSelected ? [] : recordings);
                  }}
                  className="text-xs p-0"
                >
                  {selectedRecordings.length === recordings.length ? '取消全选' : '全选'}
                </Button>
                <div className="text-xs text-gray-500">
                  共 {recordingsTotal} 条
                </div>
              </div>
            </div>

            {/* 搜索和筛选 - 紧凑版本 */}
            <div className="space-y-2">
              <Row gutter={8}>
                <Col span={12}>
                  <Select
                    placeholder="选择门店"
                    value={selectedStoreId}
                    onChange={setSelectedStoreId}
                    allowClear
                    size="small"
                    className="w-full"
                  >
                    {stores.map(store => (
                      <Select.Option key={store.id} value={store.id}>
                        {store.name}
                      </Select.Option>
                    ))}
                  </Select>
                </Col>
                <Col span={12}>
                  <Select
                    placeholder="选择点位"
                    value={selectedLocationId}
                    onChange={setSelectedLocationId}
                    allowClear
                    size="small"
                    className="w-full"
                    disabled={!selectedStoreId}
                  >
                    {locations.map(location => (
                      <Select.Option key={location.id} value={location.id}>
                        {location.name}
                      </Select.Option>
                    ))}
                  </Select>
                </Col>
              </Row>

              <Row gutter={8}>
                <Col span={24}>
                  <Select
                    mode="multiple"
                    placeholder="选择设备（可多选）"
                    value={selectedDeviceIds}
                    onChange={setSelectedDeviceIds}
                    allowClear
                    size="small"
                    className="w-full"
                    disabled={!selectedLocationId}
                    showSearch
                    filterOption={(input, option) =>
                      String(option?.children || '').toLowerCase().includes(input.toLowerCase())
                    }
                  >
                    {devices.map(device => (
                      <Select.Option key={device.id} value={device.id}>
                        {device.name ? `${device.name} (${device.mac_address.slice(-5)})` : device.mac_address}
                      </Select.Option>
                    ))}
                  </Select>
                </Col>
              </Row>

              {/* 快速设备选择按钮 */}
              {devices.length > 0 && (
                <Row gutter={8}>
                  <Col span={24}>
                    <div className="flex flex-wrap gap-1">
                      <Button
                        size="small"
                        onClick={() => setSelectedDeviceIds(devices.map(d => d.id))}
                        className="text-xs"
                        disabled={!selectedLocationId}
                      >
                        选择全部设备
                      </Button>
                      <Button
                        size="small"
                        onClick={() => setSelectedDeviceIds([])}
                        className="text-xs"
                        disabled={selectedDeviceIds.length === 0}
                      >
                        清空设备选择
                      </Button>
                      {selectedDeviceIds.length > 0 && (
                        <span className="text-xs text-gray-500 self-center">
                          已选择 {selectedDeviceIds.length} 个设备
                        </span>
                      )}
                    </div>
                  </Col>
                </Row>
              )}

              {/* 时间筛选 - 紧凑版本 */}
              <div className="space-y-1">
                {/* 快速时间选择按钮 */}
                <div className="flex flex-wrap gap-1">
                  <Button
                    size="small"
                    type={startTime && endTime && startTime.isSame(dayjs().subtract(1, 'hour'), 'minute') ? 'primary' : 'default'}
                    onClick={() => handleQuickTimeSelect('1hour')}
                    className="text-xs"
                  >
                    最近1小时
                  </Button>
                  <Button
                    size="small"
                    type={startTime && endTime && startTime.isSame(dayjs().subtract(4, 'hour'), 'minute') ? 'primary' : 'default'}
                    onClick={() => handleQuickTimeSelect('4hours')}
                    className="text-xs"
                  >
                    最近4小时
                  </Button>
                  <Button
                    size="small"
                    type={startTime && startTime.isSame(dayjs().startOf('day'), 'minute') ? 'primary' : 'default'}
                    onClick={() => handleQuickTimeSelect('today')}
                    className="text-xs"
                  >
                    今天
                  </Button>
                  <Button
                    size="small"
                    type={startTime && startTime.isSame(dayjs().subtract(1, 'day').startOf('day'), 'minute') ? 'primary' : 'default'}
                    onClick={() => handleQuickTimeSelect('yesterday')}
                    className="text-xs"
                  >
                    昨天
                  </Button>
                  <Button
                    size="small"
                    type={startTime && startTime.isSame(dayjs().subtract(7, 'day').startOf('day'), 'minute') ? 'primary' : 'default'}
                    onClick={() => handleQuickTimeSelect('7days')}
                    className="text-xs"
                  >
                    最近7天
                  </Button>
                  <Button
                    size="small"
                    type={showCustomTime ? 'primary' : 'default'}
                    onClick={handleShowCustomTime}
                    className="text-xs"
                  >
                    自定义
                  </Button>
                  {(startTime || endTime) && (
                    <Button
                      size="small"
                      onClick={handleClearTime}
                      className="text-xs"
                      danger
                    >
                      清除
                    </Button>
                  )}
                </div>

                {/* 当前选择的时间范围显示 */}
                {(startTime || endTime) && (
                  <div className="text-xs text-gray-500 bg-gray-50 p-2 rounded">
                    时间范围: {startTime ? startTime.format('YYYY-MM-DD HH:mm') : '开始'} 至 {endTime ? endTime.format('YYYY-MM-DD HH:mm') : '现在'}
                  </div>
                )}

                {/* 自定义时间选择（展开时显示） */}
                {showCustomTime && (
                  <Row gutter={8}>
                    <Col span={12}>
                      <DatePicker
                        placeholder="开始时间"
                        value={startTime}
                        onChange={setStartTime}
                        size="small"
                        className="w-full"
                        showTime
                      />
                    </Col>
                    <Col span={12}>
                      <DatePicker
                        placeholder="结束时间"
                        value={endTime}
                        onChange={setEndTime}
                        size="small"
                        className="w-full"
                        showTime
                      />
                    </Col>
                  </Row>
                )}
              </div>

              {/* 搜索按钮 */}
              <Button
                type="primary"
                icon={<SearchOutlined />}
                onClick={handleRecordingSearch}
                size="small"
                className="w-full"
              >
                搜索
              </Button>

              {/* 批量操作 */}
              {selectedRecordings.length > 0 && (
                <div className="bg-primary-50 p-2 rounded">
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-primary-700">
                      已选择 {selectedRecordings.length} 条录音
                    </span>
                    <Space size="small">
                      <Button
                        type="link"
                        size="small"
                        onClick={() => setSelectedRecordings([])}
                      >
                        清空
                      </Button>
                      <Button
                        type="primary"
                        size="small"
                        icon={<UploadOutlined />}
                        onClick={handleSendSelectedRecordingsToAI}
                      >
                        批量分析
                      </Button>
                    </Space>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* 录音列表 */}
          <div className="flex-1 overflow-y-auto">
            <List
              loading={recordingsLoading}
              dataSource={recordings}
              renderItem={(recording) => (
                <List.Item
                  className="px-4 py-3 hover:bg-gray-50 cursor-pointer border-b border-gray-100"
                  onClick={() => {
                    const isSelected = selectedRecordings.some(r => r.id === recording.id);
                    if (isSelected) {
                      setSelectedRecordings(prev => prev.filter(r => r.id !== recording.id));
                    } else {
                      setSelectedRecordings(prev => [...prev, recording]);
                    }
                  }}
                >
                  <div className="w-full">
                    <div className="flex items-start justify-between mb-2">
                      <div className="flex items-center space-x-2">
                        <input
                          type="checkbox"
                          checked={selectedRecordings.some(r => r.id === recording.id)}
                          onChange={() => {}} // 由父级onClick处理
                          className="rounded border-gray-300"
                        />
                        <Tag 
                          color={recording.status === 1 ? 'orange' : 'green'}
                        >
                          {recording.status === 1 ? '离线' : '实时'}
                        </Tag>
                      </div>
                      
                      <Tooltip title="发送给AI分析">
                        <Button
                          type="text"
                          size="small"
                          icon={<FileTextOutlined />}
                          onClick={(e) => {
                            e.stopPropagation();
                            handleSendRecordingToAI(recording);
                          }}
                          className="text-primary-600 hover:text-primary-700"
                        />
                      </Tooltip>
                    </div>

                    <div className="text-xs text-gray-500 mb-1">
                      设备: {getDeviceNameByMac(recording.mac_address)}
                    </div>
                    
                    <div className="text-xs text-gray-500 mb-2">
                      {recording.speaker_name || '未知用户'} • {formatFullDateTime(recording.created_at)}
                    </div>
                    
                    <div className="text-sm text-gray-800 line-clamp-3">
                      {recording.text}
                    </div>
                  </div>
                </List.Item>
              )}
              locale={{
                emptyText: (
                  <Empty
                    image={<SoundOutlined className="text-4xl text-gray-300" />}
                    styles={{ image: { height: 60 } }}
                    description="暂无录音记录"
                  />
                )
              }}
            />
          </div>

          {/* 分页 */}
          {recordingsTotal > 0 && (
            <div className="p-4 border-t border-gray-200">
              <Pagination
                current={recordingsPage}
                pageSize={recordingsPageSize}
                total={recordingsTotal}
                showSizeChanger={true}
                showQuickJumper={false}
                pageSizeOptions={['50', '100', '500', '1000', '2000', '5000']}
                onChange={handleRecordingPageChange}
                onShowSizeChange={handleRecordingPageSizeChange}
                size="small"
              />
              
              {/* 设备统计信息 */}
              {recordings.length > 0 && (
                <div className="mt-3 pt-3 border-t border-gray-100">
                  <div className="text-xs text-gray-500 mb-2">当前页面涉及设备：</div>
                  <div className="flex flex-wrap gap-1">
                    {Array.from(new Set(recordings.map(r => r.mac_address))).map(mac => (
                      <Tag key={mac} color="blue">
                        {getDeviceNameByMac(mac)}
                      </Tag>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}
        </Card>

        {/* 右侧聊天窗口 */}
        <Card className="flex-1 flex flex-col card-primary" styles={{ body: { padding: 0, height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' } }}>
        {/* 聊天消息区域 */}
        <div className="flex-1 p-6 overflow-y-auto hide-scrollbar" style={{ scrollbarWidth: 'none', msOverflowStyle: 'none' }}>
          {messages.length === 0 ? (
            <div className="h-80 flex items-center justify-center">
              <Empty
                image={<RobotOutlined className="text-6xl text-primary-300" />}
                styles={{ image: { height: 80 } }}
                description={
                  <div className="text-center">
                    <p className="text-lg text-gray-600 mb-2">AI助手准备就绪</p>
                    <p className="text-sm text-gray-400">
                      您可以询问关于数据分析、业务洞察、系统优化等问题
                    </p>
                  </div>
                }
              />
            </div>
          ) : (
            <div className="space-y-4">
              {messages.map((message) => (
                <div
                  key={message.id}
                  className={`flex ${message.type === 'user' ? 'justify-end' : 'justify-start'}`}
                >
                  <div className={`flex items-start space-x-3 max-w-[70%] ${message.type === 'user' ? 'flex-row-reverse space-x-reverse' : ''}`}>
                    <Avatar
                      icon={
                        message.type === 'user' 
                          ? <UserOutlined /> 
                          : message.error 
                          ? <ExclamationCircleOutlined />
                          : <RobotOutlined />
                      }
                      className={
                        message.type === 'user' 
                          ? 'bg-primary-500' 
                          : message.error 
                          ? 'bg-red-500'
                          : 'bg-green-500'
                      }
                    />
                    <div
                      className={`px-4 py-3 rounded-lg ${
                        message.type === 'user'
                          ? 'bg-primary-500 text-white'
                          : message.error
                          ? 'bg-red-50 text-red-800 border border-red-200'
                          : 'bg-gray-100 text-gray-800'
                      }`}
                    >
                      {message.loading ? (
                        <div className="flex items-center space-x-2">
                          <Spin size="small" />
                          <span className="text-gray-500">AI正在思考中...</span>
                        </div>
                      ) : message.error ? (
                        <div>
                          <div className="flex items-start space-x-2 mb-2">
                            <ExclamationCircleOutlined className="text-red-500 mt-0.5" />
                            <div className="flex-1">
                              <div className="font-medium text-red-800 mb-1">AI响应错误</div>
                              <div className="text-red-700">{message.content}</div>
                            </div>
                          </div>
                          <div className="mt-3 pt-3 border-t border-red-200 flex items-center justify-between">
                            <div className="flex-1">
                              {message.errorDetails && (
                                <div className="flex items-start space-x-2 mb-2">
                                  <InfoCircleOutlined className="text-red-400 mt-0.5 text-sm" />
                                  <div className="text-xs text-red-600">
                                    <div className="font-medium mb-1">详细错误信息:</div>
                                    <div className="bg-red-100 p-2 rounded text-red-700 font-mono text-xs">
                                      {message.errorDetails}
                                    </div>
                                  </div>
                                </div>
                              )}
                            </div>
                            {message.originalQuery && (
                              <Tooltip title="重试发送消息">
                                <Button
                                  type="text"
                                  size="small"
                                  icon={<RedoOutlined />}
                                  onClick={() => handleRetryMessage(message.id)}
                                  className="text-red-600 hover:text-red-700 hover:bg-red-100"
                                  disabled={isLoading}
                                >
                                  重试
                                </Button>
                              </Tooltip>
                            )}
                          </div>
                        </div>
                      ) : (
                        <div className="prose prose-sm max-w-none">
                          {message.type === 'user' && message.content.length > 200 ? (
                            <div>
                              <div className="whitespace-pre-wrap">
                                {expandedMessages.has(message.id) 
                                  ? message.content 
                                  : message.content.substring(0, 200) + '...'
                                }
                              </div>
                              <button
                                onClick={() => toggleMessageExpansion(message.id)}
                                className="text-xs text-black hover:text-gray-700 mt-2 underline bg-transparent border-none outline-none cursor-pointer"
                              >
                                {expandedMessages.has(message.id) ? '收起' : '展开'}
                              </button>
                            </div>
                          ) : (
                            <ReactMarkdown
                              remarkPlugins={[remarkGfm]}
                              rehypePlugins={[rehypeHighlight]}
                              components={{
                                // 自定义组件样式
                                h1: ({children}) => <h1 className="text-lg font-bold mb-2 text-gray-800">{children}</h1>,
                                h2: ({children}) => <h2 className="text-base font-bold mb-2 text-gray-800">{children}</h2>,
                                h3: ({children}) => <h3 className="text-sm font-bold mb-1 text-gray-800">{children}</h3>,
                                p: ({children}) => <p className="mb-2 text-gray-800">{children}</p>,
                                ul: ({children}) => <ul className="mb-2 ml-4 list-disc text-gray-800">{children}</ul>,
                                ol: ({children}) => <ol className="mb-2 ml-4 list-decimal text-gray-800">{children}</ol>,
                                li: ({children}) => <li className="mb-1">{children}</li>,
                                code: ({children, className}) => {
                                  const isInline = !className;
                                  return isInline ? (
                                    <code className="bg-gray-100 px-1 py-0.5 rounded text-sm font-mono text-red-600">
                                      {children}
                                    </code>
                                  ) : (
                                    <code className={className}>{children}</code>
                                  );
                                },
                                pre: ({children}) => (
                                  <pre className="bg-gray-100 p-3 rounded mb-2 overflow-x-auto text-sm">
                                    {children}
                                  </pre>
                                ),
                                blockquote: ({children}) => (
                                  <blockquote className="border-l-4 border-gray-300 pl-4 italic text-gray-600 mb-2">
                                    {children}
                                  </blockquote>
                                ),
                                table: ({children}) => (
                                  <table className="border-collapse border border-gray-300 mb-2 w-full">
                                    {children}
                                  </table>
                                ),
                                th: ({children}) => (
                                  <th className="border border-gray-300 px-2 py-1 bg-gray-100 font-bold text-left">
                                    {children}
                                  </th>
                                ),
                                td: ({children}) => (
                                  <td className="border border-gray-300 px-2 py-1">
                                    {children}
                                  </td>
                                ),
                              }}
                            >
                              {message.content}
                            </ReactMarkdown>
                          )}
                          {message.streaming && (
                            <span className="inline-block w-2 h-4 bg-primary-500 ml-1 animate-pulse" />
                          )}
                        </div>
                      )}
                      <div className={`text-xs mt-2 flex items-center justify-between ${message.type === 'user' ? 'text-primary-100' : 'text-gray-400'}`}>
                        <span>{new Date(message.timestamp).toLocaleTimeString()}</span>
                        {message.streaming && (
                          <span className="flex items-center space-x-1">
                            <Spin size="small" />
                            <span>AI分析中...</span>
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              ))}
              <div ref={messagesEndRef} />
            </div>
          )}
        </div>

        {/* 输入区域 */}
        <div className="border-t border-gray-200 p-4">
          <div className="flex space-x-3">
            <TextArea
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onKeyPress={handleKeyPress}
              placeholder="输入您的问题，按Enter发送，Shift+Enter换行..."
              autoSize={{ minRows: 1, maxRows: 4 }}
              className="flex-1"
              disabled={isLoading}
              maxLength={5000000}
              showCount
            />
            <Button
              type="primary"
              icon={<SendOutlined />}
              onClick={handleSendMessage}
              loading={isLoading}
              disabled={!inputValue.trim() || inputValue.length > 5000000}
              className="btn-primary"
            >
              发送
            </Button>
          </div>
          
          <div className="mt-2">
            <div className="text-xs text-gray-400">
              提示：您可以询问数据分析、异常检测、业务优化等相关问题
            </div>
          </div>
        </div>
        </Card>
      </div>

      {/* 系统提示词管理弹窗 */}
      <SystemPromptManagerModal
        visible={systemPromptModalVisible}
        onClose={handleSystemPromptModalClose}
        onSelect={handleSystemPromptChange}
        selectedPromptId={selectedSystemPrompt?.id}
      />
    </div>
  );
};

export default AIAnalysisPage;
