import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Table, 
  Input, 
  Button, 
  Space, 
  Tag, 
  Tooltip, 
  message,
  Popconfirm,
  Select,
  DatePicker,
  Row,
  Col,
  Statistic
} from 'antd';
import {
  SearchOutlined,
  ReloadOutlined,
  DownloadOutlined,
  DeleteOutlined,
  EyeOutlined,
  SettingOutlined,
} from '@ant-design/icons';
import { recordingService } from '@/services/recordings';
import { Recording, RecordingQueryParams } from '@/types';
import { formatDate } from '@/utils/dateUtils';

const { RangePicker } = DatePicker;
const { Option } = Select;

const RecordingsPage: React.FC = () => {
  const [recordings, setRecordings] = useState<Recording[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [searchParams, setSearchParams] = useState<RecordingQueryParams>({
    offset: 0,
    limit: 20,
  });

  // 统计数据
  const [stats, setStats] = useState({
    totalRecordings: 0,
    todayRecordings: 0,
    activeDevices: 0,
  });

  useEffect(() => {
    fetchRecordings();
    fetchStats();
  }, [searchParams]);

  const fetchRecordings = async () => {
    try {
      setLoading(true);
      const response = await recordingService.getRecordings(searchParams);
      setRecordings(response.items);
      setTotal(response.total);
    } catch (error) {
      message.error('获取录音记录失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    // 这里可以调用统计API，暂时使用模拟数据
    setStats({
      totalRecordings: total,
      todayRecordings: Math.floor(total * 0.1),
      activeDevices: Math.floor(total * 0.05),
    });
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
    const offset = (page - 1) * size;
    setSearchParams(prev => ({
      ...prev,
      offset,
      limit: size,
    }));
    setCurrentPage(page);
    setPageSize(size);
  };

  const handleDelete = async (id: number) => {
    try {
      await recordingService.deleteRecording(id);
      message.success('删除成功');
      fetchRecordings();
    } catch (error) {
      message.error('删除失败');
    }
  };

  const handleExport = async () => {
    try {
      const blob = await recordingService.exportRecordings(searchParams);
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `录音记录_${new Date().toISOString().split('T')[0]}.xlsx`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      message.success('导出成功');
    } catch (error) {
      message.error('导出失败');
    }
  };

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
      fixed: 'left' as const,
    },
    {
      title: 'MAC地址',
      dataIndex: 'mac_address',
      key: 'mac_address',
      width: 150,
      render: (text: string) => (
        <Tag color="blue" className="font-mono text-xs">
          {text}
        </Tag>
      ),
    },
    {
      title: '说话人',
      dataIndex: 'speaker_name',
      key: 'speaker_name',
      width: 120,
      render: (text: string, record: Recording) => (
        <div>
          <div className="font-medium text-primary-700">{text}</div>
          <div className="text-xs text-gray-500">{record.speaker_id}</div>
        </div>
      ),
    },
    {
      title: '录音内容',
      dataIndex: 'text',
      key: 'text',
      ellipsis: true,
      render: (text: string) => (
        <Tooltip title={text} placement="topLeft">
          <div className="max-w-xs truncate">{text}</div>
        </Tooltip>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: number) => (
        <Tag color={status === 1 ? 'success' : 'processing'}>
          {status === 1 ? '最终结果' : '中间结果'}
        </Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (timestamp: number) => formatDate(timestamp),
      sorter: true,
    },
    {
      title: '设备时间',
      dataIndex: 'device_time',
      key: 'device_time',
      width: 180,
      render: (timestamp: number) => formatDate(timestamp),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      fixed: 'right' as const,
      render: (_: any, record: Recording) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button
              type="text"
              icon={<EyeOutlined />}
              size="small"
              className="text-primary-600 hover:text-primary-700 hover:bg-primary-50"
            />
          </Tooltip>
          <Popconfirm
            title="确定要删除这条录音记录吗？"
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

  return (
    <div className="space-y-6">
      {/* 统计卡片 */}
      <Row gutter={16}>
        <Col span={8}>
          <Card className="bg-white rounded-xl shadow-lg border border-primary-100">
            <Statistic
              title="总录音数"
              value={stats.totalRecordings}
              valueStyle={{ color: '#8CC020' }}
              prefix={<SearchOutlined />}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card className="bg-white rounded-xl shadow-lg border border-primary-100">
            <Statistic
              title="今日录音数"
              value={stats.todayRecordings}
              valueStyle={{ color: '#8CC020' }}
              prefix={<SearchOutlined />}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card className="bg-white rounded-xl shadow-lg border border-primary-100">
            <Statistic
              title="活跃设备数"
              value={stats.activeDevices}
              valueStyle={{ color: '#8CC020' }}
              prefix={<SearchOutlined />}
            />
          </Card>
        </Col>
      </Row>

      {/* 搜索过滤区域 */}
      <Card className="bg-white rounded-xl shadow-lg border border-primary-100">
        <div className="space-y-4">
          <div className="flex items-center space-x-4">
            <SettingOutlined className="text-primary-500 text-lg" />
            <span className="text-lg font-semibold text-primary-700">搜索过滤</span>
          </div>
          
          <Row gutter={16}>
            <Col span={8}>
              <Input
                placeholder="输入MAC地址进行搜索"
                value={searchParams.mac_address || ''}
                onChange={(e) => setSearchParams(prev => ({ ...prev, mac_address: e.target.value }))}
                prefix={<SearchOutlined className="text-primary-400" />}
                className="w-full border-primary-200 focus:border-primary-500 focus:ring-primary-200"
                allowClear
              />
            </Col>
            <Col span={8}>
              <Select
                placeholder="选择状态"
                className="w-full border-primary-200 focus:border-primary-500"
                allowClear
                onChange={(value) => setSearchParams(prev => ({ ...prev, status: value }))}
              >
                <Option value={1}>最终结果</Option>
                <Option value={0}>中间结果</Option>
              </Select>
            </Col>
            <Col span={8}>
              <RangePicker
                className="w-full border-primary-200 focus:border-primary-500"
                placeholder={['开始时间', '结束时间']}
                onChange={(dates) => {
                  if (dates) {
                    setSearchParams(prev => ({
                      ...prev,
                      start_time: dates[0]?.valueOf(),
                      end_time: dates[1]?.valueOf(),
                    }));
                  }
                }}
              />
            </Col>
          </Row>

          <div className="flex justify-between items-center">
            <Space>
              <Button
                type="primary"
                icon={<SearchOutlined />}
                onClick={handleSearch}
                className="bg-primary-500 hover:bg-primary-600 border-primary-500 hover:border-primary-600"
              >
                搜索
              </Button>
              <Button
                icon={<ReloadOutlined />}
                onClick={handleReset}
                className="border-primary-300 text-primary-600 hover:bg-primary-50"
              >
                重置
              </Button>
            </Space>

            <Button
              icon={<DownloadOutlined />}
              onClick={handleExport}
              className="border-primary-300 text-primary-600 hover:bg-primary-50"
            >
              导出数据
            </Button>
          </div>
        </div>
      </Card>

      {/* 数据表格 */}
      <Card className="bg-white rounded-xl shadow-lg border border-primary-100">
        <Table
          columns={columns}
          dataSource={recordings}
          loading={loading}
          rowKey="id"
          pagination={{
            current: currentPage,
            pageSize: pageSize,
            total: total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => 
              `第 ${range[0]}-${range[1]} 条，共 ${total} 条记录`,
            onChange: handlePageChange,
            onShowSizeChange: handlePageChange,
          }}
          scroll={{ x: 1200 }}
          className="w-full"
          rowClassName="hover:bg-primary-50 transition-colors duration-200"
        />
      </Card>
    </div>
  );
};

export default RecordingsPage;
