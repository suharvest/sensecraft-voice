import React, { useState, useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { 
  Card, 
  Table, 
  Button, 
  Space, 
  Tag, 
  Tooltip, 
  message,
  Popconfirm,
  Input,
  Row,
  Col,
  Statistic,
  Modal,
  Form,
  Select,
  Tabs,
  Badge,
  Pagination,
  DatePicker
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  SearchOutlined,
  ShopOutlined,
  EnvironmentOutlined,
  MobileOutlined,
  SoundOutlined,
  ReloadOutlined,
  DownloadOutlined,
  UserOutlined,
  CheckOutlined,
  CloseOutlined,
  MessageOutlined,
  ClockCircleOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  StepBackwardOutlined,
  StepForwardOutlined,
  StopOutlined,
} from '@ant-design/icons';
import { storeService } from '@/services/stores';
import { locationService } from '@/services/locations';
import { deviceService } from '@/services/devices';
import { recordingService } from '@/services/recordings';
import { Store, StoreQueryParams, Location, LocationQueryParams, Device, DeviceQueryParams, Recording, RecordingQueryParams } from '@/types';
import { isDeviceOnline, getDeviceStatusText, getDeviceStatusColor, formatDate } from '@/utils/dateUtils';

const { Option } = Select;
const { TabPane } = Tabs;
const { RangePicker } = DatePicker;

const StoresPage: React.FC = () => {
  const location = useLocation();
  
  // 导航状态
  const [fromDashboard, setFromDashboard] = useState(false);
  const [dashboardMacAddress, setDashboardMacAddress] = useState<string>('');
  
  // 门店管理状态
  const [stores, setStores] = useState<Store[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [searchParams, setSearchParams] = useState<StoreQueryParams>({
    offset: 0,
    limit: 20,
  });
  const [modalVisible, setModalVisible] = useState(false);
  const [editingStore, setEditingStore] = useState<Store | null>(null);
  const [form] = Form.useForm();

  // 点位管理状态
  const [locations, setLocations] = useState<Location[]>([]);
  const [locationLoading, setLocationLoading] = useState(false);
  const [locationTotal, setLocationTotal] = useState(0);
  const [locationCurrentPage, setLocationCurrentPage] = useState(1);
  const [locationPageSize, setLocationPageSize] = useState(20);
  const [locationSearchParams, setLocationSearchParams] = useState<LocationQueryParams>({
    offset: 0,
    limit: 20,
  });
  const [locationModalVisible, setLocationModalVisible] = useState(false);
  const [editingLocation, setEditingLocation] = useState<Location | null>(null);
  const [locationForm] = Form.useForm();
  const [selectedStore, setSelectedStore] = useState<Store | null>(null);
  const [activeTab, setActiveTab] = useState('stores');

  // 设备管理状态
  const [devices, setDevices] = useState<Device[]>([]);
  const [deviceLoading, setDeviceLoading] = useState(false);
  const [deviceTotal, setDeviceTotal] = useState(0);
  const [deviceCurrentPage, setDeviceCurrentPage] = useState(1);
  const [devicePageSize, setDevicePageSize] = useState(20);
  const [deviceSearchParams, setDeviceSearchParams] = useState<DeviceQueryParams>({
    offset: 0,
    limit: 20,
  });
  const [selectedLocation, setSelectedLocation] = useState<Location | null>(null);
  const [assignModalVisible, setAssignModalVisible] = useState(false);
  const [assignForm] = Form.useForm();
  
  // 设备名称编辑状态
  const [editingDeviceId, setEditingDeviceId] = useState<number | null>(null);
  const [editingDeviceName, setEditingDeviceName] = useState('');

  // 录音管理状态
  const [recordingLoading, setRecordingLoading] = useState(false);
  const [recordingTotal, setRecordingTotal] = useState(0);
  const [recordingSearchParams, setRecordingSearchParams] = useState<RecordingQueryParams>({
    offset: 0,
    limit: 20,
  });

  // 新的录音管理设计相关状态
  const [recordingDevices, setRecordingDevices] = useState<any[]>([]);
  const [selectedDeviceMac, setSelectedDeviceMac] = useState<string>('');
  const [currentDeviceRecordings, setCurrentDeviceRecordings] = useState<Recording[]>([]);
  const [currentDeviceRecordingsTotal, setCurrentDeviceRecordingsTotal] = useState(0);
  const [currentDeviceRecordingsPage, setCurrentDeviceRecordingsPage] = useState(1);
  const [currentDeviceRecordingsPageSize, setCurrentDeviceRecordingsPageSize] = useState(20);
  const [deviceSearchKeyword, setDeviceSearchKeyword] = useState('');
  const [recordingStatusFilter, setRecordingStatusFilter] = useState<number | undefined>();
  const [deviceStatusFilter, setDeviceStatusFilter] = useState<string | undefined>();
  const [dateRange, setDateRange] = useState<[any, any] | null>(null);

  // 时间轴播放器相关状态
  const [viewMode, setViewMode] = useState<'chat' | 'timeline'>('chat');
  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [totalDuration, setTotalDuration] = useState(0);
  const [playbackRate, setPlaybackRate] = useState(1.0);
  const [volume, setVolume] = useState(80);
  const [currentRecording, setCurrentRecording] = useState<Recording | null>(null);
  const [audioQueue, setAudioQueue] = useState<Recording[]>([]);
  const [currentAudioIndex, setCurrentAudioIndex] = useState(0);
  // 当前播放的音频元素引用
  const [currentAudioRef, setCurrentAudioRef] = useState<HTMLAudioElement | null>(null);

  useEffect(() => {
    fetchStores();
  }, [searchParams]);

  // 清理音频资源
  useEffect(() => {
    return () => {
      // 组件卸载时清理
      if (currentAudioRef) {
        currentAudioRef.pause();
        currentAudioRef.currentTime = 0;
        currentAudioRef.removeEventListener('ended', () => {});
        currentAudioRef.removeEventListener('error', () => {});
      }
    };
  }, [currentAudioRef]);

  // 处理从仪表盘跳转过来的路由状态
  useEffect(() => {
    const state = location.state as { macAddress?: string; activeTab?: string } | null;
    if (state?.macAddress && state?.activeTab) {
      // 设置导航状态
      setFromDashboard(true);
      setDashboardMacAddress(state.macAddress);
      
      // 设置活跃标签页
      setActiveTab(state.activeTab);
      
      // 设置录音筛选参数
      if (state.activeTab === 'recordings') {
        setRecordingSearchParams(prev => ({
          ...prev,
          mac_address: state.macAddress,
          offset: 0,
        }));
        setSelectedDeviceMac(state.macAddress);
      }
      
      // 清除路由状态，避免重复处理
      window.history.replaceState({}, document.title);
    }
  }, [location.state]);

  useEffect(() => {
    fetchLocations();
  }, [locationSearchParams, selectedStore]);

  useEffect(() => {
    fetchDevices();
  }, [deviceSearchParams, selectedStore, selectedLocation]);

  useEffect(() => {
    if (activeTab === 'recordings') {
      fetchRecordingDevices();
      fetchRecordingStats();
    }
  }, [selectedStore, selectedLocation, activeTab]);

  // 当录音状态筛选变化时，不再自动重新加载，改为手动筛选模式

  // 当统计数据变化时更新录音总数
  useEffect(() => {
    fetchRecordingStats();
  }, [selectedDeviceMac, currentDeviceRecordingsTotal]);

  // 当recordingDevices更新后，重新计算统计数据（仅在没有选中特定设备时）
  useEffect(() => {
    if (!selectedDeviceMac) {
      fetchRecordingStats();
    }
  }, [recordingDevices]);

  // 添加定时器，每30秒更新一次设备状态（用于实时显示在线/离线状态）
  useEffect(() => {
    if (activeTab === 'devices') {
      const interval = setInterval(() => {
        // 强制重新渲染设备表格，更新在线状态显示
        setDevices(prevDevices => [...prevDevices]);
      }, 30000); // 30秒更新一次

      return () => clearInterval(interval);
    }
  }, [activeTab]);

  const fetchStores = async () => {
    try {
      setLoading(true);
      const response = await storeService.getStores(searchParams);
      setStores(response.items);
      setTotal(response.total);
    } catch (error) {
      message.error('获取门店列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = () => {
    setSearchParams(prev => ({
      ...prev,
      offset: 0,
    }));
    setCurrentPage(1);
  };

  const handleReset = () => {
    setSearchParams({
      offset: 0,
      limit: 20,
    });
    setCurrentPage(1);
  };

  const handlePageChange = (page: number, size: number) => {
    // 限制页面大小最大为100
    const limitedSize = Math.min(size, 100);
    const offset = (page - 1) * limitedSize;
    setSearchParams(prev => ({
      ...prev,
      offset,
      limit: limitedSize,
    }));
    setCurrentPage(page);
    setPageSize(limitedSize);
  };

  const handleAdd = () => {
    setEditingStore(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (store: Store) => {
    setEditingStore(store);
    form.setFieldsValue(store);
    setModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    try {
      await storeService.deleteStore(id);
      message.success('删除成功');
      fetchStores();
    } catch (error) {
      message.error('删除失败');
    }
  };

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields();
      
      if (editingStore) {
        await storeService.updateStore(editingStore.id, values);
        message.success('更新成功');
      } else {
        await storeService.createStore(values);
        message.success('创建成功');
      }
      
      setModalVisible(false);
      fetchStores();
    } catch (error) {
      message.error('操作失败');
    }
  };

  // 点位管理相关函数
  const fetchLocations = async () => {
    try {
      setLocationLoading(true);
      const params: any = {
        ...locationSearchParams,
      };
      if (selectedStore) {
        params.store_id = selectedStore.id;
      }
      const response = await locationService.getLocations(params);
      setLocations(response.items);
      setLocationTotal(response.total);
    } catch (error) {
      message.error('获取点位列表失败');
    } finally {
      setLocationLoading(false);
    }
  };

  const handleLocationSearch = () => {
    setLocationSearchParams(prev => ({
      ...prev,
      offset: 0,
    }));
    setLocationCurrentPage(1);
  };

  const handleLocationReset = () => {
    setLocationSearchParams({
      offset: 0,
      limit: 20,
    });
    setLocationCurrentPage(1);
  };

  const handleLocationPageChange = (page: number, size: number) => {
    // 限制页面大小最大为100
    const limitedSize = Math.min(size, 100);
    const offset = (page - 1) * limitedSize;
    setLocationSearchParams(prev => ({
      ...prev,
      offset,
      limit: limitedSize,
    }));
    setLocationCurrentPage(page);
    setLocationPageSize(size);
  };

  const handleLocationAdd = () => {
    setEditingLocation(null);
    locationForm.resetFields();
    if (selectedStore) {
      locationForm.setFieldsValue({ store_id: selectedStore.id });
    }
    setLocationModalVisible(true);
  };

  const handleLocationEdit = (location: Location) => {
    setEditingLocation(location);
    locationForm.setFieldsValue(location);
    setLocationModalVisible(true);
  };

  const handleLocationDelete = async (id: number) => {
    try {
      await locationService.deleteLocation(id);
      message.success('删除成功');
      fetchLocations();
    } catch (error) {
      message.error('删除失败');
    }
  };

  const handleLocationModalOk = async () => {
    try {
      const values = await locationForm.validateFields();
      
      if (editingLocation) {
        await locationService.updateLocation(editingLocation.id, values);
        message.success('更新成功');
      } else {
        await locationService.createLocation(values);
        message.success('创建成功');
      }
      
      setLocationModalVisible(false);
      fetchLocations();
    } catch (error) {
      message.error('操作失败');
    }
  };

  const handleViewLocations = (store: Store) => {
    setSelectedStore(store);
    setActiveTab('locations'); // 自动跳转到点位管理标签页
  };

  // 设备管理相关函数
  const fetchDevices = async () => {
    try {
      setDeviceLoading(true);
      let response;
      
      if (selectedLocation) {
        // 获取指定点位的设备
        response = await deviceService.getLocationDevices(selectedLocation.id, deviceSearchParams);
      } else if (selectedStore) {
        // 获取指定门店的设备
        response = await deviceService.getStoreDevices(selectedStore.id, deviceSearchParams);
      } else {
        // 获取所有设备
        response = await deviceService.getDevices(deviceSearchParams);
      }
      
      setDevices(response.items);
      setDeviceTotal(response.total);
    } catch (error) {
      message.error('获取设备列表失败');
    } finally {
      setDeviceLoading(false);
    }
  };

  const handleDeviceSearch = () => {
    setDeviceSearchParams(prev => ({
      ...prev,
      offset: 0,
    }));
    setDeviceCurrentPage(1);
  };

  // 设备名称编辑相关函数
  const handleDeviceNameEdit = (device: Device) => {
    setEditingDeviceId(device.id);
    setEditingDeviceName(device.name);
  };

  const handleDeviceNameSave = async (deviceId: number) => {
    if (!editingDeviceName.trim()) {
      message.error('设备名称不能为空');
      return;
    }

    try {
      await deviceService.updateDeviceName(deviceId, editingDeviceName.trim());
      message.success('设备名称更新成功');
      setEditingDeviceId(null);
      setEditingDeviceName('');
      fetchDevices();
    } catch (error) {
      message.error('设备名称更新失败');
    }
  };

  const handleDeviceNameCancel = () => {
    setEditingDeviceId(null);
    setEditingDeviceName('');
  };

  // 根据状态筛选器过滤设备
  const getFilteredDevices = () => {
    let filteredDevices = devices;
    
    if (deviceSearchParams.status) {
      filteredDevices = devices.filter(device => {
        const isOnline = isDeviceOnline(device.updated_at);
        if (deviceSearchParams.status === 'online') {
          return isOnline;
        } else if (deviceSearchParams.status === 'offline') {
          return !isOnline;
        }
        return true;
      });
    }
    
    return filteredDevices;
  };

  const handleDeviceReset = () => {
    setDeviceSearchParams({
      offset: 0,
      limit: 20,
    });
    setDeviceCurrentPage(1);
  };

  const handleDevicePageChange = (page: number, size: number) => {
    // 限制页面大小最大为100
    const limitedSize = Math.min(size, 100);
    const offset = (page - 1) * limitedSize;
    setDeviceSearchParams(prev => ({
      ...prev,
      offset,
      limit: limitedSize,
    }));
    setDeviceCurrentPage(page);
    setDevicePageSize(size);
  };

  const handleViewStoreDevices = (store: Store) => {
    setSelectedStore(store);
    setSelectedLocation(null);
    setActiveTab('devices');
  };

  const handleViewLocationDevices = (location: Location) => {
    setSelectedLocation(location);
    setActiveTab('devices');
  };

  // 录音管理入口函数
  const handleViewStoreRecordings = (store: Store) => {
    setSelectedStore(store);
    setSelectedLocation(null);
    setRecordingSearchParams(prev => ({ ...prev, mac_address: undefined }));
    setSelectedDeviceMac(''); // 重置选中的设备
    setActiveTab('recordings');
  };

  const handleViewLocationRecordings = (location: Location) => {
    setSelectedLocation(location);
    setRecordingSearchParams(prev => ({ ...prev, mac_address: undefined }));
    setSelectedDeviceMac(''); // 重置选中的设备
    setActiveTab('recordings');
  };

  const handleViewDeviceRecordings = (device: Device) => {
    // 设置设备MAC地址作为筛选条件，保留门店和点位信息用于显示上下文
    setRecordingSearchParams(prev => ({
      ...prev,
      mac_address: device.mac_address,
      offset: 0,
    }));
    setSelectedDeviceMac(device.mac_address); // 预设选中的设备
    setActiveTab('recordings');
  };

  const handleAssignDevice = (device: Device) => {
    assignForm.setFieldsValue({
      device_id: device.id,
      location_id: device.location_id,
    });
    setAssignModalVisible(true);
  };

  const handleAssignModalOk = async () => {
    try {
      const values = await assignForm.validateFields();
      
      if (values.location_id) {
        await deviceService.assignDevice(values.device_id, values.location_id);
        message.success('设备分配成功');
      } else {
        await deviceService.unassignDevice(values.device_id);
        message.success('设备取消分配成功');
      }
      
      setAssignModalVisible(false);
      fetchDevices();
    } catch (error) {
      message.error('操作失败');
    }
  };



  const fetchRecordingStats = () => {
    // 更新录音总数统计 - 使用当前选中设备的录音总数，如果没有选中设备则使用所有设备的录音总数
    if (selectedDeviceMac) {
      setRecordingTotal(currentDeviceRecordingsTotal);
    } else {
      // 如果没有选中设备，可以计算所有设备的录音总数（这里先简化处理）
      setRecordingTotal(recordingDevices.reduce((total, device) => total + (device.total_recordings || 0), 0));
    }
  };



  const handleRecordingExport = async () => {
    try {
      // 根据当前筛选条件导出录音
      const exportParams: RecordingQueryParams = {
        offset: 0,
        limit: 100,
      };
      
      // 如果选中了特定设备，只导出该设备的录音
      if (selectedDeviceMac) {
        exportParams.mac_address = selectedDeviceMac;
      }
      
      // 应用录音状态筛选
      if (recordingStatusFilter !== undefined) {
        exportParams.status = recordingStatusFilter;
      }

      // 应用日期范围筛选
      if (dateRange && dateRange[0] && dateRange[1]) {
        exportParams.start_time = dateRange[0].unix();
        exportParams.end_time = dateRange[1].unix();
      }
      
      const blob = await recordingService.exportRecordings(exportParams);
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      const dateRangeStr = dateRange ? `_${dateRange[0].format('YYYY-MM-DD')}_到_${dateRange[1].format('YYYY-MM-DD')}` : '';
      a.download = `录音记录_${selectedDeviceMac || '全部设备'}${dateRangeStr}_${new Date().toISOString().split('T')[0]}.xlsx`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      message.success('导出成功');
    } catch (error) {
      message.error('导出失败');
    }
  };

  // 新的录音管理相关函数
  
  // 筛选后的设备列表
  const filteredDevices = recordingDevices.filter(device => {
    // 关键词筛选
    if (deviceSearchKeyword && !device.mac_address.toLowerCase().includes(deviceSearchKeyword.toLowerCase())) {
      return false;
    }
    
    // 设备状态筛选
    if (deviceStatusFilter) {
      if (deviceStatusFilter === 'online' && !device.online_status) return false;
      if (deviceStatusFilter === 'offline' && device.online_status) return false;
    }
    
    return true;
  });

  // 当前选中的设备信息
  const selectedDevice = recordingDevices.find(device => device.mac_address === selectedDeviceMac);

  // 获取设备列表和录音数据
  const fetchRecordingDevices = async () => {
    try {
      setRecordingLoading(true);
      
      // 根据入口条件获取设备列表
      let allDevices: any[] = [];
      
      if (recordingSearchParams.mac_address) {
        // 从设备进入：只获取指定MAC地址的设备
        try {
          const device = await deviceService.getDeviceByMac(recordingSearchParams.mac_address);
          allDevices = [device];
        } catch (error) {
          // 如果单个设备查询失败，尝试通过列表查询
          const deviceResponse = await deviceService.getDevices({ 
            mac_address: recordingSearchParams.mac_address,
            offset: 0,
            limit: 100
          });
          allDevices = deviceResponse.items;
        }
      } else if (selectedLocation) {
        // 从点位进入：获取该点位下的所有设备
        let offset = 0;
        const limit = 100;
        
        while (true) {
          const deviceResponse = await deviceService.getLocationDevices(selectedLocation.id, { offset, limit });
          allDevices = allDevices.concat(deviceResponse.items);
          
          if (deviceResponse.items.length < limit) break;
          offset += limit;
        }
      } else if (selectedStore) {
        // 从门店进入：获取该门店下的所有设备
        let offset = 0;
        const limit = 100;
        
        while (true) {
          const deviceResponse = await deviceService.getStoreDevices(selectedStore.id, { offset, limit });
          allDevices = allDevices.concat(deviceResponse.items);
          
          if (deviceResponse.items.length < limit) break;
          offset += limit;
        }
      } else {
        // 直接进入：获取所有设备
        let offset = 0;
        const limit = 100;
        
        while (true) {
          const deviceResponse = await deviceService.getDevices({ offset, limit });
          allDevices = allDevices.concat(deviceResponse.items);
          
          if (deviceResponse.items.length < limit) break;
          offset += limit;
        }
      }
      
      // 处理设备数据，添加额外信息
      const processedDevices = allDevices.map(device => ({
        ...device,
        online_status: isDeviceOnline(device.updated_at),
        device_name: device.name || `设备-${device.id}`,
        location_info: getDeviceLocationInfo(device),
        last_activity: formatDate(device.updated_at),
        new_messages_count: 0, // 这里可以后续实现新消息计数
        total_recordings: 0, // 这里可以后续实现录音总数
      }));
      
      setRecordingDevices(processedDevices);
      
      // 如果是从设备进入，自动选中该设备
      if (recordingSearchParams.mac_address) {
        setSelectedDeviceMac(recordingSearchParams.mac_address);
        // 立即加载该设备的录音记录
        await loadDeviceRecordings(recordingSearchParams.mac_address);
      } else if (processedDevices.length > 0 && !selectedDeviceMac) {
        // 否则选中第一个设备
        setSelectedDeviceMac(processedDevices[0].mac_address);
        // 加载第一个设备的录音记录
        await loadDeviceRecordings(processedDevices[0].mac_address);
      }
      
    } catch (error) {
      message.error('获取设备列表失败');
    } finally {
      setRecordingLoading(false);
    }
  };

  // 获取设备位置信息
  const getDeviceLocationInfo = (device: any) => {
    if (device.location_id) {
      const location = locations.find(l => l.id === device.location_id);
      if (location) {
        const store = stores.find(s => s.id === location.store_id);
        return `${store?.name || '未知门店'} - ${location.name}`;
      }
    }
    
    if (device.store_id) {
      const store = stores.find(s => s.id === device.store_id);
      return `${store?.name || '未知门店'} - 未分配点位`;
    }
    
    return '未分配';
  };

  // 选择设备
  const handleDeviceSelect = async (macAddress: string) => {
    setSelectedDeviceMac(macAddress);
    // 重置分页并加载录音，应用当前的筛选条件
    setCurrentDeviceRecordingsPage(1);
    await loadDeviceRecordings(macAddress, 1, currentDeviceRecordingsPageSize);
  };

  // 加载指定设备的录音记录
  const loadDeviceRecordings = async (macAddress: string, page: number = 1, pageSize: number = 20) => {
    try {
      setRecordingLoading(true);
      
      const queryParams: RecordingQueryParams = {
        mac_address: macAddress,
        status: recordingStatusFilter,
        offset: (page - 1) * pageSize,
        limit: pageSize,
      };

      // 添加日期范围筛选
      if (dateRange && dateRange[0] && dateRange[1]) {
        queryParams.start_time = dateRange[0].unix(); // dayjs的unix()方法获取时间戳
        queryParams.end_time = dateRange[1].unix();
      }
      
      const response = await recordingService.getRecordings(queryParams);
      
      setCurrentDeviceRecordings(response.items);
      setCurrentDeviceRecordingsTotal(response.total);
      setCurrentDeviceRecordingsPage(page);
      setCurrentDeviceRecordingsPageSize(pageSize);
      
      // 初始化时间轴播放器
      if (viewMode === 'timeline') {
        initializeTimeline(response.items);
      }
      
    } catch (error) {
      message.error('获取录音记录失败');
      setCurrentDeviceRecordings([]);
      setCurrentDeviceRecordingsTotal(0);
    } finally {
      setRecordingLoading(false);
    }
  };

  // 刷新录音数据
  const handleRefreshRecordings = async () => {
    await fetchRecordingDevices();
    if (selectedDeviceMac) {
      await loadDeviceRecordings(selectedDeviceMac);
    }
  };

  // 刷新指定设备的录音
  const handleRefreshDeviceRecordings = async (macAddress: string) => {
    await loadDeviceRecordings(macAddress, currentDeviceRecordingsPage, currentDeviceRecordingsPageSize);
  };

  // 处理录音分页变化
  const handleDeviceRecordingPageChange = (page: number, size: number) => {
    // 限制页面大小最大为100
    const limitedSize = Math.min(size, 100);
    if (selectedDeviceMac) {
      loadDeviceRecordings(selectedDeviceMac, page, limitedSize);
    }
  };

  // 应用筛选条件
  const handleRecordingFilter = async () => {
    // 重新获取设备列表（应用设备状态筛选）
    await fetchRecordingDevices();
    
    // 如果有选中的设备，重新加载其录音（应用录音状态筛选）
    if (selectedDeviceMac) {
      setCurrentDeviceRecordingsPage(1);
      await loadDeviceRecordings(selectedDeviceMac, 1, currentDeviceRecordingsPageSize);
    }
  };

  // 重置所有筛选条件
  const handleResetRecordingFilters = async () => {
    // 重置筛选条件
    setDeviceSearchKeyword('');
    setRecordingStatusFilter(undefined);
    setDeviceStatusFilter(undefined);
    setDateRange(null);
    
    // 重新获取设备列表
    await fetchRecordingDevices();
    
    // 如果有选中的设备，重新加载其录音
    if (selectedDeviceMac) {
      setCurrentDeviceRecordingsPage(1);
      await loadDeviceRecordings(selectedDeviceMac, 1, currentDeviceRecordingsPageSize);
    }
  };

  // 时间轴播放器相关函数
  const initializeTimeline = (recordings: Recording[]) => {
    // 过滤掉 play_url 为空的录音，按时间排序
    const validRecordings = recordings
      .filter(recording => recording.play_url && recording.play_url.trim() !== '')
      .sort((a, b) => a.created_at - b.created_at);
    
    // 计算总时长（假设每段录音5秒，实际应该从音频文件获取）
    const estimatedDuration = validRecordings.length * 5;
    
    setAudioQueue(validRecordings);
    setTotalDuration(estimatedDuration);
    setCurrentTime(0);
    setCurrentAudioIndex(0);
    setCurrentRecording(validRecordings[0] || null);
  };

  // 播放指定索引的音频
  const playAudioAtIndex = (index: number) => {
    if (index >= audioQueue.length || !audioQueue[index].play_url) {
      console.log('播放结束或无效索引');
      setIsPlaying(false);
      return;
    }
    
    // 停止当前播放的音频
    if (currentAudioRef) {
      console.log('停止上一个音频');
      currentAudioRef.pause();
      currentAudioRef.currentTime = 0;
      currentAudioRef.removeEventListener('ended', () => {});
      currentAudioRef.removeEventListener('error', () => {});
    }
    
    const recording = audioQueue[index];
    console.log(`播放第 ${index + 1} 段音频:`, recording.play_url);
    
    // 更新状态
    setCurrentAudioIndex(index);
    setCurrentRecording(recording);
    setCurrentTime(index * 5);
    
    // 创建新的音频元素并播放
    const audio = new Audio();
    audio.src = recording.play_url;
    audio.volume = volume / 100;
    audio.playbackRate = playbackRate;
    
    // 保存当前音频引用
    setCurrentAudioRef(audio);
    
    audio.addEventListener('ended', () => {
      console.log(`第 ${index + 1} 段音频播放结束`);
      // 自动播放下一段
      playAudioAtIndex(index + 1);
    });
    
    audio.addEventListener('error', (e) => {
      console.error('音频播放错误:', e);
      message.error('音频播放失败');
      setIsPlaying(false);
    });
    
    audio.play().catch(error => {
      console.error('音频播放失败:', error);
      message.error('音频播放失败');
      setIsPlaying(false);
    });
  };

  const handlePlay = () => {
    if (audioQueue.length === 0) {
      message.warning('没有可播放的录音');
      return;
    }
    
    if (!currentRecording || !currentRecording.play_url) {
      message.warning('当前录音没有音频链接');
      return;
    }
    
    setIsPlaying(true);
    
    // 从当前索引开始播放
    playAudioAtIndex(currentAudioIndex);
  };

  const handlePause = () => {
    setIsPlaying(false);
    // 暂停当前音频
    if (currentAudioRef) {
      currentAudioRef.pause();
    }
  };

  const handleStop = () => {
    setIsPlaying(false);
    // 停止当前音频
    if (currentAudioRef) {
      currentAudioRef.pause();
      currentAudioRef.currentTime = 0;
      currentAudioRef.removeEventListener('ended', () => {});
      currentAudioRef.removeEventListener('error', () => {});
      setCurrentAudioRef(null);
    }
    setCurrentTime(0);
    setCurrentAudioIndex(0);
    setCurrentRecording(audioQueue[0] || null);
  };

  const handleNext = () => {
    const nextIndex = Math.min(audioQueue.length - 1, currentAudioIndex + 1);
    if (nextIndex !== currentAudioIndex) {
      if (isPlaying) {
        // 如果正在播放，继续播放下一段
        playAudioAtIndex(nextIndex);
      } else {
        // 如果没在播放，只更新状态
        setCurrentAudioIndex(nextIndex);
        setCurrentRecording(audioQueue[nextIndex]);
        setCurrentTime(nextIndex * 5);
      }
    }
  };

  const handlePrevious = () => {
    const prevIndex = Math.max(0, currentAudioIndex - 1);
    if (prevIndex !== currentAudioIndex) {
      if (isPlaying) {
        // 如果正在播放，继续播放上一段
        playAudioAtIndex(prevIndex);
      } else {
        // 如果没在播放，只更新状态
        setCurrentAudioIndex(prevIndex);
        setCurrentRecording(audioQueue[prevIndex]);
        setCurrentTime(prevIndex * 5);
      }
    }
  };

  const handleTimelineSeek = (time: number) => {
    setCurrentTime(time);
    const index = Math.floor(time / 5);
    if (index < audioQueue.length) {
      if (isPlaying) {
        // 如果正在播放，跳转到指定位置并继续播放
        playAudioAtIndex(index);
      } else {
        // 如果没在播放，只更新状态
        setCurrentRecording(audioQueue[index]);
        setCurrentAudioIndex(index);
      }
    }
  };

  const handleVolumeChange = (newVolume: number) => {
    setVolume(newVolume);
  };

  const handlePlaybackRateChange = (rate: number) => {
    setPlaybackRate(rate);
  };

  // 格式化时间显示
  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '门店名称',
      dataIndex: 'name',
      key: 'name',
      width: 150,
      render: (text: string) => (
        <span className="font-medium text-primary-700">{text}</span>
      ),
    },
    {
      title: '门店代码',
      dataIndex: 'code',
      key: 'code',
      width: 120,
      render: (text: string) => (
        <Tag color="blue" className="font-mono">{text}</Tag>
      ),
    },
    {
      title: '地址',
      dataIndex: 'address',
      key: 'address',
      ellipsis: true,
      render: (text: string) => (
        <Tooltip title={text} placement="topLeft">
          <span className="max-w-xs truncate">{text}</span>
        </Tooltip>
      ),
    },
    {
      title: '联系人',
      dataIndex: 'contact',
      key: 'contact',
      width: 100,
    },
    {
      title: '联系电话',
      dataIndex: 'phone',
      key: 'phone',
      width: 120,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: number) => (
        <Tag color={status === 1 ? 'success' : 'error'}>
          {status === 1 ? '正常' : '停用'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_: any, record: Store) => (
        <Space size="small">
          <Tooltip title="查看点位">
            <Button
              type="text"
              icon={<EnvironmentOutlined />}
              size="small"
              className="text-green-600 hover:text-green-700 hover:bg-green-50"
              onClick={() => handleViewLocations(record)}
            />
          </Tooltip>
          <Tooltip title="查看设备">
            <Button
              type="text"
              icon={<MobileOutlined />}
              size="small"
              className="text-blue-600 hover:text-blue-700 hover:bg-blue-50"
              onClick={() => handleViewStoreDevices(record)}
            />
          </Tooltip>
          <Tooltip title="查看录音">
            <Button
              type="text"
              icon={<SoundOutlined />}
              size="small"
              className="text-green-600 hover:text-green-700 hover:bg-green-50"
              onClick={() => handleViewStoreRecordings(record)}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button
              type="text"
              icon={<EditOutlined />}
              size="small"
              className="text-primary-600 hover:text-primary-700 hover:bg-primary-50"
              onClick={() => handleEdit(record)}
            />
          </Tooltip>
          <Popconfirm
            title="确定要删除这个门店吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="删除">
              <Button
                type="text"
                icon={<DeleteOutlined />}
                size="small"
                className="text-red-500 hover:text-red-600 hover:bg-red-50"
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  // 点位管理表格列定义
  const locationColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '点位名称',
      dataIndex: 'name',
      key: 'name',
      width: 150,
      render: (text: string) => (
        <span className="font-medium text-primary-700">{text}</span>
      ),
    },
    {
      title: '点位代码',
      dataIndex: 'code',
      key: 'code',
      width: 120,
      render: (text: string) => (
        <Tag color="green" className="font-mono">{text}</Tag>
      ),
    },
    // 仅在未选择门店时显示所属门店列
    ...(!selectedStore ? [{
      title: '所属门店',
      dataIndex: 'store_id',
      key: 'store_id',
      width: 120,
      render: (storeId: number) => {
        const store = stores.find(s => s.id === storeId);
        return store ? (
          <Tag color="blue">{store.name}</Tag>
        ) : (
          <span className="text-gray-400">未知门店</span>
        );
      },
    }] : []),
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
      render: (text: string) => (
        <Tooltip title={text} placement="topLeft">
          <span className="max-w-xs truncate">{text}</span>
        </Tooltip>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: number) => (
        <Tag color={status === 1 ? 'success' : 'error'}>
          {status === 1 ? '正常' : '停用'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: any, record: Location) => (
        <Space size="small">
          <Tooltip title="查看设备">
            <Button
              type="text"
              icon={<MobileOutlined />}
              size="small"
              className="text-blue-600 hover:text-blue-700 hover:bg-blue-50"
              onClick={() => handleViewLocationDevices(record)}
            />
          </Tooltip>
          <Tooltip title="查看录音">
            <Button
              type="text"
              icon={<SoundOutlined />}
              size="small"
              className="text-green-600 hover:text-green-700 hover:bg-green-50"
              onClick={() => handleViewLocationRecordings(record)}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button
              type="text"
              icon={<EditOutlined />}
              size="small"
              className="text-primary-600 hover:text-primary-700 hover:bg-primary-50"
              onClick={() => handleLocationEdit(record)}
            />
          </Tooltip>
          <Popconfirm
            title="确定要删除这个点位吗？"
            onConfirm={() => handleLocationDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="删除">
              <Button
                type="text"
                icon={<DeleteOutlined />}
                size="small"
                className="text-red-500 hover:text-red-600 hover:bg-red-50"
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  // 设备管理表格列定义
  const deviceColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '设备名称',
      dataIndex: 'name',
      key: 'name',
      width: 200,
      render: (text: string, record: Device) => {
        const isEditing = editingDeviceId === record.id;
        
        if (isEditing) {
          return (
            <div className="flex items-center space-x-2">
              <Input
                value={editingDeviceName}
                onChange={(e) => setEditingDeviceName(e.target.value)}
                onPressEnter={() => handleDeviceNameSave(record.id)}
                size="small"
                className="flex-1"
                autoFocus
              />
              <Tooltip title="保存">
                <Button
                  type="link"
                  size="small"
                  icon={<CheckOutlined />}
                  onClick={() => handleDeviceNameSave(record.id)}
                  className="text-green-600 hover:text-green-700"
                />
              </Tooltip>
              <Tooltip title="取消">
                <Button
                  type="link"
                  size="small"
                  icon={<CloseOutlined />}
                  onClick={handleDeviceNameCancel}
                  className="text-red-600 hover:text-red-700"
                />
              </Tooltip>
            </div>
          );
        }
        
        return (
          <div className="flex items-center space-x-2 group">
            <span className="font-medium text-primary-700">{text}</span>
            <Tooltip title="编辑设备名称">
              <Button
                type="link"
                size="small"
                icon={<EditOutlined />}
                onClick={() => handleDeviceNameEdit(record)}
                className="opacity-0 group-hover:opacity-100 text-gray-500 hover:text-primary-600"
              />
            </Tooltip>
          </div>
        );
      },
    },
    {
      title: 'MAC地址',
      dataIndex: 'mac_address',
      key: 'mac_address',
      width: 150,
      render: (text: string) => (
        <span className="font-mono text-gray-600">{text}</span>
      ),
    },
    {
      title: 'IP地址',
      dataIndex: 'ip_address',
      key: 'ip_address',
      width: 120,
      render: (text: string) => (
        <span className="font-mono text-blue-600">
          {text || <span className="text-gray-400">未分配</span>}
        </span>
      ),
    },
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      width: 100,
      render: (text: string) => (
        <Tag color="cyan">{text}</Tag>
      ),
    },
    // 仅在未选择门店和点位时显示所属门店列
    ...(!selectedStore && !selectedLocation ? [{
      title: '所属门店',
      dataIndex: 'store_id',
      key: 'store_id',
      width: 120,
      render: (storeId: number) => {
        const store = stores.find(s => s.id === storeId);
        return store ? (
          <Tag color="blue">{store.name}</Tag>
        ) : (
          <span className="text-gray-400">未分配</span>
        );
      },
    }] : []),
    // 仅在未选择点位时显示所属点位列
    ...(!selectedLocation ? [{
      title: '所属点位',
      dataIndex: 'location_id',
      key: 'location_id',
      width: 120,
      render: (locationId: number) => {
        const location = locations.find(l => l.id === locationId);
        return location ? (
          <Tag color="green">{location.name}</Tag>
        ) : (
          <span className="text-gray-400">未分配</span>
        );
      },
    }] : []),
    {
      title: 'CPU使用率',
      dataIndex: 'cpu_usage_percent',
      key: 'cpu_usage_percent',
      width: 120,
      render: (percent: number) => (
        <div className="flex items-center space-x-2">
          <span className="text-sm">{percent.toFixed(1)}%</span>
          <div className="w-12 bg-gray-200 rounded-full h-2">
            <div 
              className={`h-2 rounded-full ${
                percent > 80 ? 'bg-red-500' : 
                percent > 60 ? 'bg-yellow-500' : 
                'bg-green-500'
              }`}
              style={{ width: `${Math.min(percent, 100)}%` }}
            />
          </div>
        </div>
      ),
    },
    {
      title: '内存使用',
      dataIndex: 'memory_used_bytes',
      key: 'memory_used_bytes',
      width: 120,
      render: (bytes: number, record: Device) => {
        const totalBytes = record.memory_total_bytes;
        if (!totalBytes || totalBytes === 0) {
          return <span className="text-sm text-gray-400">未知</span>;
        }
        const percent = (bytes / totalBytes) * 100;
        return (
          <div className="flex items-center space-x-2">
            <span className="text-sm">{percent.toFixed(1)}%</span>
            <div className="w-12 bg-gray-200 rounded-full h-2">
              <div 
                className={`h-2 rounded-full ${
                  percent > 80 ? 'bg-red-500' : 
                  percent > 60 ? 'bg-yellow-500' : 'bg-green-500'
                }`}
                style={{ width: `${Math.min(percent, 100)}%` }}
              />
            </div>
          </div>
        );
      },
    },
    {
      title: '磁盘使用',
      dataIndex: 'disk_used_bytes',
      key: 'disk_used_bytes',
      width: 150,
      render: (bytes: number, record: Device) => {
        const totalBytes = record.disk_total_bytes;
        if (!totalBytes || totalBytes === 0) {
          return <span className="text-sm text-gray-400">未知</span>;
        }
        
        const percent = (bytes / totalBytes) * 100;
        const remainingBytes = totalBytes - bytes;
        
        // 计算剩余录音时长
        // 16位单通道WAV音频：采样率通常为16kHz或8kHz，这里假设16kHz
        // 数据率 = 采样率 × 位深度 / 8 = 16000 × 16 / 8 = 32000 bytes/秒
        const audioDataRate = 16000 * 16 / 8; // 32KB/s
        const remainingSeconds = remainingBytes / audioDataRate;
        
        // 格式化时长显示
        const formatDuration = (seconds: number) => {
          if (seconds < 60) {
            return `${Math.floor(seconds)}秒`;
          } else if (seconds < 3600) {
            return `${Math.floor(seconds / 60)}分钟`;
          } else if (seconds < 86400) {
            return `${Math.floor(seconds / 3600)}小时`;
          } else {
            return `${Math.floor(seconds / 86400)}天`;
          }
        };
        
        return (
          <div className="flex flex-col space-y-1">
            <div className="flex items-center space-x-2">
              <span className="text-sm">{percent.toFixed(1)}%</span>
              <div className="w-12 bg-gray-200 rounded-full h-2">
                <div 
                  className={`h-2 rounded-full ${
                    percent > 80 ? 'bg-red-500' : 
                    percent > 60 ? 'bg-yellow-500' : 'bg-green-500'
                  }`}
                  style={{ width: `${Math.min(percent, 100)}%` }}
                />
              </div>
            </div>
            <div className="text-xs text-gray-500">
              <b> 剩余录音时长:</b> {formatDuration(remainingSeconds)}
            </div>
          </div>
        );
      },
    },
    {
      title: '状态',
      dataIndex: 'updated_at',
      key: 'status',
      width: 100,
      render: (updatedAt: number) => {
        const statusText = getDeviceStatusText(updatedAt);
        const statusColor = getDeviceStatusColor(updatedAt);
        
        return (
          <Tag color={statusColor}>
            {statusText}
          </Tag>
        );
      },
    },
    {
      title: '最后更新',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 150,
      render: (updatedAt: number) => {
        const isOnline = isDeviceOnline(updatedAt);
        const timeText = new Date(updatedAt).toLocaleString('zh-CN', {
          month: '2-digit',
          day: '2-digit',
          hour: '2-digit',
          minute: '2-digit'
        });
        
        return (
          <div className="flex flex-col">
            <span className="text-gray-600">{timeText}</span>
            <span className={`text-xs ${isOnline ? 'text-green-600' : 'text-red-600'}`}>
              {isOnline ? '活跃' : '离线'}
            </span>
          </div>
        );
      },
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 150,
      render: (time: number) => (
        <span className="text-gray-600">
          {new Date(time).toLocaleString()}
        </span>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_: any, record: Device) => (
        <Space size="small">
          <Tooltip title="查看录音">
            <Button
              type="text"
              icon={<SoundOutlined />}
              size="small"
              className="text-green-600 hover:text-green-700 hover:bg-green-50"
              onClick={() => handleViewDeviceRecordings(record)}
            />
          </Tooltip>
          <Tooltip title="分配点位">
            <Button
              type="text"
              icon={<EnvironmentOutlined />}
              size="small"
              className="text-primary-600 hover:text-primary-700 hover:bg-primary-50"
              onClick={() => handleAssignDevice(record)}
            />
          </Tooltip>
        </Space>
      ),
    },
  ];



  return (
    <div className="space-y-6">
      {/* 从仪表盘跳转的导航提示 */}
      {fromDashboard && (
        <Card className="bg-blue-50 border-blue-200">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <Badge status="processing" />
              <span className="text-blue-700">
                从仪表盘跳转 - 正在查看设备 <strong>{dashboardMacAddress}</strong> 的录音记录
              </span>
            </div>
            <Button 
              type="link" 
              size="small"
              onClick={() => {
                setFromDashboard(false);
                setRecordingSearchParams(prev => ({ ...prev, mac_address: undefined }));
                setSelectedDeviceMac('');
              }}
            >
              清除筛选
            </Button>
          </div>
        </Card>
      )}
      
      <Tabs 
        activeKey={activeTab} 
        onChange={setActiveTab}
        className="tabs-primary"
      >
        <TabPane 
          tab={
            <span>
              <ShopOutlined />
              门店管理
            </span>
          } 
          key="stores"
        >
          {/* 门店统计卡片 */}
          <Row gutter={16} className="mb-6">
            <Col span={8}>
              <Card className="card-primary">
                <Statistic
                  title="总门店数"
                  value={total}
                  valueStyle={{ color: '#8CC020' }}
                  prefix={<ShopOutlined />}
                />
              </Card>
            </Col>
            <Col span={8}>
              <Card className="card-primary">
                <Statistic
                  title="正常运营"
                  value={stores.filter(s => s.status === 1).length}
                  valueStyle={{ color: '#8CC020' }}
                  prefix={<ShopOutlined />}
                />
              </Card>
            </Col>
            <Col span={8}>
              <Card className="card-primary">
                <Statistic
                  title="暂停营业"
                  value={stores.filter(s => s.status === 0).length}
                  valueStyle={{ color: '#8CC020' }}
                  prefix={<ShopOutlined />}
                />
              </Card>
            </Col>
          </Row>

          {/* 门店搜索和操作区域 */}
          <div className="sticky top-0 z-10 bg-white pb-6">
            <Card className="card-primary shadow-lg border-b">
              <div className="space-y-4">
                <Row gutter={16}>
                  <Col span={8}>
                    <Input
                      placeholder="搜索门店名称"
                      value={searchParams.name || ''}
                      onChange={(e) => setSearchParams(prev => ({ ...prev, name: e.target.value }))}
                      prefix={<SearchOutlined className="text-primary-400" />}
                      className="input-primary"
                      allowClear
                    />
                  </Col>
                  <Col span={8}>
                    <Input
                      placeholder="搜索门店代码"
                      value={searchParams.code || ''}
                      onChange={(e) => setSearchParams(prev => ({ ...prev, code: e.target.value }))}
                      prefix={<SearchOutlined className="text-primary-400" />}
                      className="input-primary"
                      allowClear
                    />
                  </Col>
                  <Col span={8}>
                    <Space>
                      <Button
                        type="primary"
                        icon={<SearchOutlined />}
                        onClick={handleSearch}
                        className="btn-primary"
                      >
                        搜索
                      </Button>
                      <Button
                        icon={<SearchOutlined />}
                        onClick={handleReset}
                        className="btn-secondary"
                      >
                        重置
                      </Button>
                      <Button
                        type="primary"
                        icon={<PlusOutlined />}
                        onClick={handleAdd}
                        className="btn-primary"
                      >
                        新增门店
                      </Button>
                    </Space>
                  </Col>
                </Row>
              </div>
            </Card>
          </div>

          {/* 门店数据表格 */}
          <Card className="card-primary">
            <Table
              columns={columns}
              dataSource={stores}
              loading={loading}
              rowKey="id"
              pagination={{
                current: currentPage,
                pageSize: pageSize,
                total: total,
                showSizeChanger: true,
                showQuickJumper: true,
                pageSizeOptions: ['20', '50', '100'],
                showTotal: (total, range) => 
                  `第 ${range[0]}-${range[1]} 条，共 ${total} 条记录`,
                onChange: handlePageChange,
                onShowSizeChange: handlePageChange,
              }}
              className="table-primary"
              rowClassName="table-row-hover"
            />
          </Card>
        </TabPane>

        <TabPane 
          tab={
            <span>
              <EnvironmentOutlined />
              点位管理
              <Badge count={locationTotal} size="small" />
            </span>
          } 
          key="locations"
        >
          <div className="space-y-6">
            {/* 当前选中的门店信息 - 仅在选择门店时显示 */}
            {selectedStore && (
              <Card className="card-primary">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-4">
                    <ShopOutlined className="text-primary-500 text-xl" />
                    <div>
                      <h3 className="text-lg font-semibold text-primary-700">
                        {selectedStore.name}
                      </h3>
                      <p className="text-gray-500">
                        门店代码: {selectedStore.code} | 地址: {selectedStore.address}
                      </p>
                    </div>
                  </div>
                  <Button
                    type="default"
                    onClick={() => {
                      setSelectedStore(null);
                      setActiveTab('stores');
                    }}
                    className="btn-secondary"
                  >
                    返回门店列表
                  </Button>
                </div>
              </Card>
            )}

            {/* 点位统计卡片 */}
            <Row gutter={16} className="mb-6">
              <Col span={8}>
                <Card className="card-primary">
                  <Statistic
                    title="总点位数"
                    value={locationTotal}
                    valueStyle={{ color: '#8CC020' }}
                    prefix={<EnvironmentOutlined />}
                  />
                </Card>
              </Col>
              <Col span={8}>
                <Card className="card-primary">
                  <Statistic
                    title="正常点位"
                    value={locations.filter(l => l.status === 1).length}
                    valueStyle={{ color: '#8CC020' }}
                    prefix={<EnvironmentOutlined />}
                  />
                </Card>
              </Col>
              <Col span={8}>
                <Card className="card-primary">
                  <Statistic
                    title="停用点位"
                    value={locations.filter(l => l.status === 0).length}
                    valueStyle={{ color: '#8CC020' }}
                    prefix={<EnvironmentOutlined />}
                  />
                </Card>
              </Col>
            </Row>

            {/* 点位搜索和操作区域 */}
            <div className="sticky top-0 z-10 bg-white pb-6">
              <Card className="card-primary shadow-lg border-b">
                <div className="space-y-4">
                  <Row gutter={16}>
                    <Col span={8}>
                      <Input
                        placeholder="搜索点位名称"
                        value={locationSearchParams.name || ''}
                        onChange={(e) => setLocationSearchParams(prev => ({ ...prev, name: e.target.value }))}
                        prefix={<SearchOutlined className="text-primary-400" />}
                        className="input-primary"
                        allowClear
                      />
                    </Col>
                    <Col span={8}>
                      <Input
                        placeholder="搜索点位代码"
                        value={locationSearchParams.code || ''}
                        onChange={(e) => setLocationSearchParams(prev => ({ ...prev, code: e.target.value }))}
                        prefix={<SearchOutlined className="text-primary-400" />}
                        className="input-primary"
                        allowClear
                      />
                    </Col>
                    <Col span={8}>
                      <Space>
                        <Button
                          type="primary"
                          icon={<SearchOutlined />}
                          onClick={handleLocationSearch}
                          className="btn-primary"
                        >
                          搜索
                        </Button>
                        <Button
                          icon={<SearchOutlined />}
                          onClick={handleLocationReset}
                          className="btn-secondary"
                        >
                          重置
                        </Button>
                        <Button
                          type="primary"
                          icon={<PlusOutlined />}
                          onClick={handleLocationAdd}
                          className="btn-primary"
                        >
                          新增点位
                        </Button>
                      </Space>
                    </Col>
                  </Row>
                </div>
              </Card>
            </div>

            {/* 点位数据表格 */}
            <Card className="card-primary">
              <Table
                columns={locationColumns}
                dataSource={locations}
                loading={locationLoading}
                rowKey="id"
                pagination={{
                  current: locationCurrentPage,
                  pageSize: locationPageSize,
                  total: locationTotal,
                  showSizeChanger: true,
                  showQuickJumper: true,
                  pageSizeOptions: ['20', '50', '100'],
                  showTotal: (total, range) => 
                    `第 ${range[0]}-${range[1]} 条，共 ${total} 条记录`,
                  onChange: handleLocationPageChange,
                  onShowSizeChange: handleLocationPageChange,
                }}
                className="table-primary"
                rowClassName="table-row-hover"
              />
            </Card>
          </div>
        </TabPane>

        <TabPane 
          tab={
            <span>
              <MobileOutlined />
              设备管理
              <Badge count={deviceTotal} size="small" />
            </span>
          } 
          key="devices"
        >
          <div className="space-y-6">
            {/* 当前选中的门店/点位信息 */}
            {(selectedStore || selectedLocation) && (
              <Card className="card-primary">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-4">
                    {selectedLocation ? (
                      <>
                        <EnvironmentOutlined className="text-green-500 text-xl" />
                        <div>
                          <h3 className="text-lg font-semibold text-primary-700">
                            点位: {selectedLocation.name}
                          </h3>
                          <p className="text-gray-500">
                            点位代码: {selectedLocation.code}
                          </p>
                        </div>
                      </>
                    ) : (
                      <>
                        <ShopOutlined className="text-primary-500 text-xl" />
                        <div>
                          <h3 className="text-lg font-semibold text-primary-700">
                            门店: {selectedStore!.name}
                          </h3>
                          <p className="text-gray-500">
                            门店代码: {selectedStore!.code}
                          </p>
                        </div>
                      </>
                    )}
                  </div>
                  <Button
                    type="default"
                    onClick={() => {
                      setSelectedStore(null);
                      setSelectedLocation(null);
                    }}
                    className="btn-secondary"
                  >
                    查看全部设备
                  </Button>
                </div>
              </Card>
            )}

            {/* 设备统计卡片 */}
            <Row gutter={16} className="mb-6">
              <Col span={8}>
                <Card className="card-primary">
                  <Statistic
                    title="总设备数"
                    value={deviceTotal}
                    valueStyle={{ color: '#8CC020' }}
                    prefix={<MobileOutlined />}
                  />
                </Card>
              </Col>
              <Col span={8}>
                <Card className="card-primary">
                  <Statistic
                    title="在线设备"
                    value={devices.filter(d => isDeviceOnline(d.updated_at)).length}
                    valueStyle={{ color: '#8CC020' }}
                    prefix={<MobileOutlined />}
                  />
                </Card>
              </Col>
              <Col span={8}>
                <Card className="card-primary">
                  <Statistic
                    title="离线设备"
                    value={devices.filter(d => !isDeviceOnline(d.updated_at)).length}
                    valueStyle={{ color: '#8CC020' }}
                    prefix={<MobileOutlined />}
                  />
                </Card>
              </Col>
            </Row>

            {/* 设备搜索和操作区域 */}
            <div className="sticky top-0 z-10 bg-white pb-6">
              <Card className="card-primary shadow-lg border-b">
                <div className="space-y-4">
                  <Row gutter={16}>
                    <Col span={5}>
                      <Input
                        placeholder="搜索设备名称"
                        value={deviceSearchParams.name || ''}
                        onChange={(e) => setDeviceSearchParams(prev => ({ ...prev, name: e.target.value }))}
                        prefix={<SearchOutlined className="text-primary-400" />}
                        className="input-primary"
                        allowClear
                      />
                    </Col>
                    <Col span={5}>
                      <Input
                        placeholder="搜索MAC地址"
                        value={deviceSearchParams.mac_address || ''}
                        onChange={(e) => setDeviceSearchParams(prev => ({ ...prev, mac_address: e.target.value }))}
                        prefix={<SearchOutlined className="text-primary-400" />}
                        className="input-primary"
                        allowClear
                      />
                    </Col>
                    <Col span={6}>
                      <Select
                        placeholder="在线状态"
                        value={deviceSearchParams.status}
                        onChange={(value) => setDeviceSearchParams(prev => ({ ...prev, status: value }))}
                        className="input-primary"
                        allowClear
                      >
                        <Option value="online">在线</Option>
                        <Option value="offline">离线</Option>
                      </Select>
                    </Col>
                    <Col span={8}>
                      <Space>
                        <Button
                          type="primary"
                          icon={<SearchOutlined />}
                          onClick={handleDeviceSearch}
                          className="btn-primary"
                        >
                          搜索
                        </Button>
                        <Button
                          icon={<SearchOutlined />}
                          onClick={handleDeviceReset}
                          className="btn-secondary"
                        >
                          重置
                        </Button>
                      </Space>
                    </Col>
                  </Row>
                </div>
              </Card>
            </div>

            {/* 设备状态指示器 */}
            <Card className="card-primary mb-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-6">
                  <div className="flex items-center space-x-2">
                    <div className="w-3 h-3 bg-green-500 rounded-full animate-pulse"></div>
                    <span className="text-sm text-gray-600">
                      在线: {devices.filter(d => isDeviceOnline(d.updated_at)).length}
                    </span>
                  </div>
                  <div className="flex items-center space-x-2">
                    <div className="w-3 h-3 bg-red-500 rounded-full"></div>
                    <span className="text-sm text-gray-600">
                      离线: {devices.filter(d => !isDeviceOnline(d.updated_at)).length}
                    </span>
                  </div>
                </div>
                <div className="flex items-center space-x-4">
                  <div className="text-sm text-gray-500">
                    最后更新: {new Date().toLocaleTimeString('zh-CN')}
                  </div>
                  <div className="text-xs text-gray-400">
                    ⏰ 每30秒自动刷新
                  </div>
                </div>
              </div>
            </Card>

            {/* 设备数据表格 */}
            <Card className="card-primary">
              <Table
                columns={deviceColumns}
                dataSource={getFilteredDevices()}
                loading={deviceLoading}
                rowKey="id"
                pagination={{
                  current: deviceCurrentPage,
                  pageSize: devicePageSize,
                  total: getFilteredDevices().length,
                  showSizeChanger: true,
                  showQuickJumper: true,
                  pageSizeOptions: ['20', '50', '100'],
                  showTotal: (total, range) => 
                    `第 ${range[0]}-${range[1]} 条，共 ${total} 条记录`,
                  onChange: handleDevicePageChange,
                  onShowSizeChange: handleDevicePageChange,
                }}
                className="table-primary"
                rowClassName="table-row-hover"
              />
            </Card>
          </div>
        </TabPane>

        <TabPane 
          tab={
            <span>
              <SoundOutlined />
              录音管理
              <Badge count={recordingTotal} size="small" />
            </span>
          } 
          key="recordings"
        >
          <div className="h-full flex flex-col space-y-4">
            {/* 录音管理页面标题栏 */}
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-4">
                <SoundOutlined className="text-primary-500 text-xl" />
                <div>
                  <h2 className="text-xl font-semibold text-primary-700">录音管理</h2>
                  <p className="text-gray-500 text-sm">
                    {(selectedStore || selectedLocation || recordingSearchParams.mac_address) ? (
                      recordingSearchParams.mac_address ? 
                        `设备: ${recordingSearchParams.mac_address}` :
                        selectedLocation ? 
                          `点位: ${selectedLocation.name}` : 
                          `门店: ${selectedStore!.name}`
                    ) : '全部设备录音'}
                  </p>
                </div>
              </div>
              
              {/* 统计信息 */}
              <div className="flex items-center space-x-6">
                <div className="text-center">
                  <div className="text-lg font-semibold text-primary-700">{recordingDevices.length}</div>
                  <div className="text-xs text-gray-500">设备数</div>
                </div>
                <div className="text-center">
                  <div className="text-lg font-semibold text-primary-700">{recordingTotal}</div>
                  <div className="text-xs text-gray-500">录音数</div>
                </div>
                <div className="text-center">
                  <div className="text-lg font-semibold text-green-600">{recordingDevices.filter(d => d.online_status).length}</div>
                  <div className="text-xs text-gray-500">在线</div>
                </div>
              </div>
            </div>

            {/* 搜索过滤区域 */}
            <div className="sticky top-0 z-10 bg-white pb-4">
              <Card className="card-primary shadow-lg border-b">
                <Row gutter={16} className="mb-4">
                  <Col span={8}>
                    <Input
                      placeholder="搜索设备MAC地址"
                      value={deviceSearchKeyword}
                      onChange={(e) => setDeviceSearchKeyword(e.target.value)}
                      prefix={<SearchOutlined className="text-primary-400" />}
                      className="input-primary"
                      allowClear
                    />
                  </Col>
                  <Col span={4}>
                    <Select
                      placeholder="录音状态"
                      className="input-primary"
                      value={recordingStatusFilter}
                      onChange={setRecordingStatusFilter}
                      allowClear
                    >
                      <Option value={0}>实时数据</Option>
                      <Option value={1}>离线数据</Option>
                    </Select>
                  </Col>
                  <Col span={4}>
                    <Select
                      placeholder="设备状态"
                      className="input-primary"
                      value={deviceStatusFilter}
                      onChange={setDeviceStatusFilter}
                      allowClear
                    >
                      <Option value="online">在线设备</Option>
                      <Option value="offline">离线设备</Option>
                    </Select>
                  </Col>
                  <Col span={8}>
                    <Space>
                      <Button
                        type="primary"
                        icon={<SearchOutlined />}
                        onClick={handleRecordingFilter}
                        className="btn-primary"
                      >
                        筛选
                      </Button>
                      <Button
                        icon={<ReloadOutlined />}
                        onClick={handleResetRecordingFilters}
                        className="btn-secondary"
                      >
                        重置
                      </Button>
                      <Button
                        icon={<ReloadOutlined />}
                        onClick={handleRefreshRecordings}
                        className="btn-secondary"
                      >
                        刷新
                      </Button>
                      <Button
                        icon={<DownloadOutlined />}
                        onClick={handleRecordingExport}
                        className="btn-secondary"
                      >
                        导出
                      </Button>
                    </Space>
                  </Col>
                </Row>
                <Row gutter={16}>
                  <Col span={8}>
                    <RangePicker
                      placeholder={['开始日期', '结束日期']}
                      value={dateRange}
                      onChange={setDateRange}
                      className="input-primary w-full"
                      showTime={{
                        format: 'HH:mm'
                      }}
                      format="YYYY-MM-DD HH:mm"
                    />
                  </Col>
                  <Col span={16}>
                    <div className="text-sm text-gray-500 pt-2">
                      {dateRange && dateRange[0] && dateRange[1] ? 
                        `已选择时间范围：${dateRange[0].format('YYYY-MM-DD HH:mm')} 至 ${dateRange[1].format('YYYY-MM-DD HH:mm')}` : 
                        '请选择录音时间范围进行筛选'
                      }
                    </div>
                  </Col>
                </Row>
              </Card>
            </div>

            {/* 主内容区域：左侧设备列表 + 右侧对话框 */}
            <div className="flex-1 flex space-x-4 min-h-0">
              {/* 左侧设备MAC列表 */}
              <Card className="card-primary w-80 flex flex-col">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-lg font-medium text-primary-700">设备列表</h3>
                  <Badge count={filteredDevices.length} showZero color="#8CC020" />
                </div>
                
                <div className="flex-1 overflow-y-auto space-y-2">
                  {filteredDevices.map((device) => (
                    <div
                      key={device.mac_address}
                      className={`p-3 rounded-lg border cursor-pointer transition-all duration-200 ${
                        selectedDeviceMac === device.mac_address
                          ? 'border-primary-500 bg-primary-50 shadow-sm'
                          : 'border-gray-200 hover:border-primary-300 hover:bg-gray-50'
                      }`}
                      onClick={() => handleDeviceSelect(device.mac_address)}
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center space-x-2 mb-1">
                            <div className={`w-2 h-2 rounded-full ${device.online_status ? 'bg-green-500' : 'bg-gray-400'}`} />
                            <span className="font-mono text-sm font-medium text-gray-900 truncate">
                              {device.mac_address}
                            </span>
                          </div>
                          
                          <div className="text-xs text-gray-600 mb-1">
                            {device.device_name || '未命名设备'}
                          </div>
                          
                          <div className="text-xs text-gray-500 flex items-center space-x-1">
                            <EnvironmentOutlined className="w-3 h-3" />
                            <span className="truncate">{device.location_info}</span>
                          </div>
                          
                          <div className="text-xs text-gray-400 mt-1">
                            {device.last_activity}
                          </div>
                        </div>
                        
                        {device.new_messages_count > 0 && (
                          <Badge count={device.new_messages_count} size="small" />
                        )}
                      </div>
                    </div>
                  ))}
                  
                  {filteredDevices.length === 0 && (
                    <div className="text-center py-8 text-gray-500">
                      <MobileOutlined className="text-4xl mb-2" />
                      <p>暂无设备数据</p>
                    </div>
                  )}
                </div>
              </Card>

              {/* 右侧对话框展示区域 */}
              <Card className="card-primary flex-1 flex flex-col">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center space-x-3">
                    <div className="flex items-center space-x-2">
                      <div className={`w-3 h-3 rounded-full ${selectedDevice?.online_status ? 'bg-green-500' : 'bg-gray-400'}`} />
                      <h3 className="text-lg font-medium text-primary-700">
                        {selectedDevice ? selectedDevice.mac_address : '请选择设备'}
                      </h3>
                    </div>
                    {selectedDevice && (
                      <div className="text-sm text-gray-500">
                        {selectedDevice.device_name} • {selectedDevice.location_info}
                      </div>
                    )}
                  </div>
                  
                  {selectedDevice && (
                    <div className="flex items-center space-x-4">
                      <span className="text-sm text-gray-500">
                        共 {currentDeviceRecordingsTotal} 条录音
                      </span>
                      <Button
                        size="small"
                        icon={<ReloadOutlined />}
                        onClick={() => handleRefreshDeviceRecordings(selectedDeviceMac)}
                        loading={recordingLoading}
                      >
                        刷新
                      </Button>
                    </div>
                  )}
                </div>

                {/* 对话框内容区域 */}
                <div className="flex-1 overflow-y-auto">
                  {selectedDevice ? (
                    <div className="space-y-4">
                      {/* 视图切换按钮 */}
                      <div className="flex items-center justify-between mb-4">
                        <div className="flex items-center space-x-2">
                          <span className="text-sm text-gray-500">查看模式：</span>
                          <Button.Group>
                            <Button 
                              type={viewMode === 'chat' ? 'primary' : 'default'}
                              size="small"
                              icon={<MessageOutlined />}
                              onClick={() => setViewMode('chat')}
                            >
                              对话模式
                            </Button>
                            <Button 
                              type={viewMode === 'timeline' ? 'primary' : 'default'}
                              size="small"
                              icon={<ClockCircleOutlined />}
                              onClick={() => {
                                setViewMode('timeline');
                                if (currentDeviceRecordings.length > 0) {
                                  initializeTimeline(currentDeviceRecordings);
                                }
                              }}
                            >
                              时间轴模式
                            </Button>
                          </Button.Group>
                        </div>
                        
                        {viewMode === 'timeline' && currentDeviceRecordings.length > 0 && (
                          <div className="flex items-center space-x-4">
                            <span className="text-sm text-gray-500">
                              {formatTime(currentTime)} / {formatTime(totalDuration)}
                            </span>
                            <span className="text-xs text-gray-400">
                              有效录音: {audioQueue.length} / {currentDeviceRecordings.length}
                            </span>
                          </div>
                        )}
                      </div>

                      {currentDeviceRecordings.length > 0 ? (
                        viewMode === 'timeline' ? (
                          // 时间轴模式
                          audioQueue.length > 0 ? (
                            <div className="space-y-4">
                            {/* 播放控制栏 */}
                            <Card className="bg-gray-50">
                              <div className="flex items-center justify-between">
                                <div className="flex items-center space-x-2">
                                  <Button
                                    type="text"
                                    icon={<StepBackwardOutlined />}
                                    size="small"
                                    onClick={handlePrevious}
                                  />
                                  <Button
                                    type="primary"
                                    icon={isPlaying ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
                                    size="small"
                                    onClick={isPlaying ? handlePause : handlePlay}
                                  />
                                  <Button
                                    type="text"
                                    icon={<StepForwardOutlined />}
                                    size="small"
                                    onClick={handleNext}
                                  />
                                  <Button
                                    type="text"
                                    icon={<StopOutlined />}
                                    size="small"
                                    onClick={handleStop}
                                  />
                                </div>
                                
                                <div className="flex items-center space-x-4">
                                  <div className="flex items-center space-x-2">
                                    <span className="text-sm text-gray-500">音量:</span>
                                    <input
                                      type="range"
                                      min="0"
                                      max="100"
                                      value={volume}
                                      onChange={(e) => handleVolumeChange(parseInt(e.target.value))}
                                      className="w-20"
                                    />
                                    <span className="text-xs text-gray-400">{volume}%</span>
                                  </div>
                                  
                                  <div className="flex items-center space-x-2">
                                    <span className="text-sm text-gray-500">倍速:</span>
                                    <select
                                      value={playbackRate}
                                      onChange={(e) => handlePlaybackRateChange(parseFloat(e.target.value))}
                                      className="text-sm border rounded px-2 py-1"
                                    >
                                      <option value={0.5}>0.5x</option>
                                      <option value={1.0}>1.0x</option>
                                      <option value={1.5}>1.5x</option>
                                      <option value={2.0}>2.0x</option>
                                    </select>
                                  </div>
                                </div>
                              </div>
                              
                              {/* 时间轴进度条 */}
                              <div className="mt-4">
                                <div className="relative">
                                  <div className="w-full bg-gray-200 rounded-full h-2">
                                    <div 
                                      className="bg-primary-500 h-2 rounded-full transition-all duration-100"
                                      style={{ width: `${(currentTime / totalDuration) * 100}%` }}
                                    />
                                  </div>
                                  <input
                                    type="range"
                                    min="0"
                                    max={totalDuration}
                                    value={currentTime}
                                    onChange={(e) => handleTimelineSeek(parseFloat(e.target.value))}
                                    className="absolute top-0 left-0 w-full h-2 opacity-0 cursor-pointer"
                                  />
                                </div>
                                
                                {/* 时间刻度 */}
                                <div className="flex justify-between text-xs text-gray-500 mt-2">
                                  <span>00:00</span>
                                  <span>{formatTime(totalDuration / 2)}</span>
                                  <span>{formatTime(totalDuration)}</span>
                                </div>
                              </div>
                            </Card>

                            {/* 当前播放内容 */}
                            {currentRecording && (
                              <Card className="bg-blue-50 border-blue-200">
                                <div className="flex items-start space-x-3">
                                  <div className="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
                                    <UserOutlined className="text-blue-600 text-sm" />
                                  </div>
                                  <div className="flex-1">
                                    <div className="flex items-center space-x-2 mb-1">
                                      <span className="font-medium text-blue-700">
                                        {currentRecording.speaker_name || '未知用户'}
                                      </span>
                                      <span className="text-xs text-blue-500">
                                        {currentRecording.speaker_id}
                                      </span>
                                      <Tag color="blue">
                                        {currentRecording.status === 1 ? '离线' : '实时'}
                                      </Tag>
                                    </div>
                                    <div className="text-gray-800 text-sm">
                                      {currentRecording.text}
                                    </div>
                                    <div className="text-xs text-gray-500 mt-1">
                                      {formatDate(currentRecording.created_at)}
                                    </div>
                                    {/* 调试信息 */}
                                    <div className="text-xs text-gray-400 mt-2 p-2 bg-gray-100 rounded">
                                      <div>音频URL: {currentRecording.play_url}</div>
                                      <div>播放状态: {isPlaying ? '播放中' : '已暂停'}</div>
                                      <div>音量: {volume}% | 倍速: {playbackRate}x</div>
                                      <div>当前索引: {currentAudioIndex + 1} / {audioQueue.length}</div>
                                    </div>
                                  </div>
                                </div>
                              </Card>
                            )}

                            {/* 录音片段时间轴 */}
                            <Card>
                              <div className="space-y-2">
                                <h4 className="text-sm font-medium text-gray-700 mb-3">录音片段</h4>
                                <div className="space-y-1">
                                  {audioQueue.map((recording, index) => (
                                    <div
                                      key={recording.id}
                                      className={`p-2 rounded border cursor-pointer transition-colors ${
                                        index === currentAudioIndex
                                          ? 'border-primary-500 bg-primary-50'
                                          : 'border-gray-200 hover:border-primary-300 hover:bg-gray-50'
                                      }`}
                                      onClick={() => {
                                        setCurrentAudioIndex(index);
                                        setCurrentRecording(recording);
                                        setCurrentTime(index * 5);
                                      }}
                                    >
                                      <div className="flex items-center justify-between">
                                        <div className="flex items-center space-x-2">
                                          <div className="w-6 h-6 bg-gray-100 rounded-full flex items-center justify-center text-xs">
                                            {index + 1}
                                          </div>
                                          <span className="text-sm font-medium">
                                            {recording.speaker_name || '未知用户'}
                                          </span>
                                          <span className="text-xs text-gray-500">
                                            {formatTime(index * 5)} - {formatTime((index + 1) * 5)}
                                          </span>
                                        </div>
                                        <div className="text-xs text-gray-600 truncate max-w-xs">
                                          {recording.text}
                                        </div>
                                      </div>
                                    </div>
                                  ))}
                                </div>
                              </div>
                            </Card>
                          </div>
                          ) : (
                            // 没有有效录音时的提示
                            <div className="text-center py-12 text-gray-500">
                              <SoundOutlined className="text-4xl mb-4" />
                              <p className="text-lg mb-2">没有可播放的录音</p>
                              <p className="text-sm">
                                当前设备有 {currentDeviceRecordings.length} 条录音记录，
                                但没有包含有效的音频链接
                              </p>
                            </div>
                          )
                        ) : (
                          // 对话模式
                          <div className="space-y-4">
                            {currentDeviceRecordings.map((recording) => (
                          <div key={recording.id} className="flex space-x-3">
                            <div className="flex-shrink-0">
                              <div className="w-8 h-8 bg-primary-100 rounded-full flex items-center justify-center">
                                <UserOutlined className="text-primary-600 text-sm" />
                              </div>
                            </div>
                            
                            <div className="flex-1 min-w-0">
                              <div className="bg-white rounded-lg border border-gray-200 p-3 shadow-sm">
                                <div className="flex items-center justify-between mb-2">
                                  <div className="flex items-center space-x-2">
                                    <span className="font-medium text-gray-900">
                                      {recording.speaker_name || '未知用户'}
                                    </span>
                                    <span className="text-xs text-gray-500">
                                      {recording.speaker_id}
                                    </span>
                                  </div>
                                  
                                  <div className="flex items-center space-x-2">
                                    {recording.play_url && recording.play_url.trim() !== '' && (
                                      <Button
                                        type="text"
                                        size="small"
                                        icon={<PlayCircleOutlined />}
                                        onClick={() => {
                                          setViewMode('timeline');
                                          initializeTimeline(currentDeviceRecordings);
                                          // 找到当前录音在时间轴中的位置
                                          const validRecordings = currentDeviceRecordings
                                            .filter(r => r.play_url && r.play_url.trim() !== '')
                                            .sort((a, b) => a.created_at - b.created_at);
                                          const index = validRecordings.findIndex(r => r.id === recording.id);
                                          if (index !== -1) {
                                            setCurrentAudioIndex(index);
                                            setCurrentRecording(validRecordings[index]);
                                            setCurrentTime(index * 5);
                                          }
                                        }}
                                        className="text-primary-600 hover:text-primary-700"
                                        title="在时间轴中播放"
                                      />
                                    )}
                                    <Tag 
                                      color={recording.status === 1 ? 'orange' : 'green'}
                                    >
                                      {recording.status === 1 ? '离线' : '实时'}
                                    </Tag>
                                    <span className="text-xs text-gray-400">
                                      {formatDate(recording.created_at)}
                                    </span>
                                  </div>
                                </div>
                                
                                <div className="text-gray-800">
                                  {recording.text}
                                </div>
                                
                                {recording.device_time !== recording.created_at && (
                                  <div className="text-xs text-gray-400 mt-2">
                                    设备时间: {formatDate(recording.device_time)}
                                  </div>
                                )}
                              </div>
                            </div>
                          </div>
                            ))}
                          </div>
                        )
                      ) : (
                        <div className="text-center py-12 text-gray-500">
                          <SoundOutlined className="text-4xl mb-4" />
                          <p className="text-lg mb-2">暂无录音记录</p>
                          <p className="text-sm">该设备还没有产生录音数据</p>
                        </div>
                      )}
                      
                      {recordingLoading && (
                        <div className="text-center py-4">
                          <div className="inline-flex items-center space-x-2 text-gray-500">
                            <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-primary-500"></div>
                            <span>加载中...</span>
                          </div>
                        </div>
                      )}
                      
                      {/* 分页组件 */}
                      {currentDeviceRecordingsTotal > 0 && (
                        <div className="flex justify-center pt-4 border-t border-gray-100">
                          <Pagination
                            current={currentDeviceRecordingsPage}
                            pageSize={currentDeviceRecordingsPageSize}
                            total={currentDeviceRecordingsTotal}
                            showSizeChanger={true}
                            showQuickJumper={true}
                            pageSizeOptions={['10', '20', '50', '100']}
                            showTotal={(total, range) => 
                              `第 ${range[0]}-${range[1]} 条，共 ${total} 条录音`
                            }
                            onChange={handleDeviceRecordingPageChange}
                            onShowSizeChange={handleDeviceRecordingPageChange}
                            size="small"
                          />
                        </div>
                      )}
                    </div>
                  ) : (
                    <div className="flex items-center justify-center h-full text-gray-500">
                      <div className="text-center">
                        <SoundOutlined className="text-6xl mb-4" />
                        <p className="text-xl mb-2">选择设备查看录音</p>
                        <p className="text-sm">从左侧列表中选择一个设备来查看其录音记录</p>
                      </div>
                    </div>
                  )}
                </div>
              </Card>
            </div>
          </div>
        </TabPane>
      </Tabs>

      {/* 新增/编辑门店模态框 */}
      <Modal
        title={editingStore ? '编辑门店' : '新增门店'}
        open={modalVisible}
        onOk={handleModalOk}
        onCancel={() => setModalVisible(false)}
        width={600}
        okText="确定"
        cancelText="取消"
      >
        <Form
          form={form}
          layout="vertical"
          className="mt-4"
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="name"
                label="门店名称"
                rules={[
                  { required: true, message: '请输入门店名称' },
                  { max: 50, message: '门店名称最多50个字符' },
                ]}
              >
                <Input placeholder="请输入门店名称" className="input-primary" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="code"
                label="门店代码"
                rules={[
                  { required: true, message: '请输入门店代码' },
                  { max: 20, message: '门店代码最多20个字符' },
                ]}
              >
                <Input placeholder="请输入门店代码" className="input-primary" />
              </Form.Item>
            </Col>
          </Row>
          
          <Form.Item
            name="address"
            label="门店地址"
            rules={[
              { required: true, message: '请输入门店地址' },
              { max: 200, message: '门店地址最多200个字符' },
            ]}
          >
            <Input.TextArea 
              placeholder="请输入门店地址" 
              rows={3}
              className="input-primary"
            />
          </Form.Item>
          
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="contact"
                label="联系人"
                rules={[
                  { required: true, message: '请输入联系人' },
                  { max: 20, message: '联系人最多20个字符' },
                ]}
              >
                <Input placeholder="请输入联系人" className="input-primary" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="phone"
                label="联系电话"
                rules={[
                  { required: true, message: '请输入联系电话' },
                  { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号码' },
                ]}
              >
                <Input placeholder="请输入联系电话" className="input-primary" />
              </Form.Item>
            </Col>
          </Row>
          
          <Form.Item
            name="status"
            label="门店状态"
            initialValue={1}
          >
            <Select className="input-primary">
              <Option value={1}>正常运营</Option>
              <Option value={0}>暂停营业</Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>

      {/* 新增/编辑点位模态框 */}
      <Modal
        title={editingLocation ? '编辑点位' : '新增点位'}
        open={locationModalVisible}
        onOk={handleLocationModalOk}
        onCancel={() => setLocationModalVisible(false)}
        width={600}
        okText="确定"
        cancelText="取消"
      >
        <Form
          form={locationForm}
          layout="vertical"
          className="mt-4"
        >
          {!selectedStore && (
            <Form.Item
              name="store_id"
              label="所属门店"
              rules={[
                { required: true, message: '请选择所属门店' },
              ]}
            >
              <Select placeholder="请选择门店" className="input-primary">
                {stores.map(store => (
                  <Option key={store.id} value={store.id}>
                    {store.name} ({store.code})
                  </Option>
                ))}
              </Select>
            </Form.Item>
          )}
          
          {selectedStore && (
            <Form.Item
              name="store_id"
              label="所属门店"
              hidden
            >
              <Input />
            </Form.Item>
          )}
          
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="name"
                label="点位名称"
                rules={[
                  { required: true, message: '请输入点位名称' },
                  { max: 50, message: '点位名称最多50个字符' },
                ]}
              >
                <Input placeholder="请输入点位名称" className="input-primary" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="code"
                label="点位代码"
                rules={[
                  { required: true, message: '请输入点位代码' },
                  { max: 20, message: '点位代码最多20个字符' },
                ]}
              >
                <Input placeholder="请输入点位代码" className="input-primary" />
              </Form.Item>
            </Col>
          </Row>
          
          <Form.Item
            name="description"
            label="点位描述"
            rules={[
              { max: 200, message: '点位描述最多200个字符' },
            ]}
          >
            <Input.TextArea 
              placeholder="请输入点位描述" 
              rows={3}
              className="input-primary"
            />
          </Form.Item>
          
          <Form.Item
            name="status"
            label="点位状态"
            initialValue={1}
          >
            <Select className="input-primary">
              <Option value={1}>正常</Option>
              <Option value={0}>停用</Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>

      {/* 设备分配模态框 */}
      <Modal
        title="设备分配"
        open={assignModalVisible}
        onOk={handleAssignModalOk}
        onCancel={() => setAssignModalVisible(false)}
        width={500}
        okText="确定"
        cancelText="取消"
      >
        <Form
          form={assignForm}
          layout="vertical"
          className="mt-4"
        >
          <Form.Item
            name="device_id"
            label="设备ID"
            hidden
          >
            <Input />
          </Form.Item>
          
          <Form.Item
            name="location_id"
            label="分配到点位"
          >
            <Select 
              placeholder="选择点位（留空则取消分配）" 
              className="input-primary"
              allowClear
            >
              {locations.map(location => (
                <Option key={location.id} value={location.id}>
                  {location.name} ({location.code})
                </Option>
              ))}
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default StoresPage;
