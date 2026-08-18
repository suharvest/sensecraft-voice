import React, { useState, useEffect } from 'react';
import {
  Modal,
  Table,
  Button,
  Input,
  Space,
  Tag,
  Popconfirm,
  message,
  Form,
  Select,
  Switch,
  Card,
  Row,
  Col,
  Typography,
  Tooltip,
  Empty,
  Pagination
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  SearchOutlined,
  SettingOutlined,
  CopyOutlined,
  CheckOutlined
} from '@ant-design/icons';
import { systemPromptService } from '@/services/systemPrompts';
import { SystemPrompt, CreateSystemPromptRequest, UpdateSystemPromptRequest } from '@/types';
import './SystemPromptManager.css';

const { TextArea } = Input;
const { Option } = Select;
const { Text, Paragraph } = Typography;

interface SystemPromptManagerModalProps {
  visible: boolean;
  onClose: () => void;
  onSelect: (prompt: SystemPrompt) => void;
  selectedPromptId?: number;
}

const SystemPromptManagerModal: React.FC<SystemPromptManagerModalProps> = ({
  visible,
  onClose,
  onSelect,
  selectedPromptId
}) => {
  const [prompts, setPrompts] = useState<SystemPrompt[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [roleFilter, setRoleFilter] = useState<string>('');
  const [activeFilter, setActiveFilter] = useState<boolean | undefined>(undefined);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  
  // 编辑相关状态
  const [editingPrompt, setEditingPrompt] = useState<SystemPrompt | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [form] = Form.useForm();

  // 获取系统提示词列表
  const fetchPrompts = async () => {
    try {
      setLoading(true);
      const response = await systemPromptService.getSystemPrompts({
        name: searchText || undefined,
        role: roleFilter || undefined,
        active: activeFilter,
        offset: (currentPage - 1) * pageSize,
        limit: pageSize
      });
      setPrompts(response.items);
      setTotal(response.total);
    } catch (error) {
      message.error('获取系统提示词列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (visible) {
      fetchPrompts();
    }
  }, [visible, currentPage, pageSize, searchText, roleFilter, activeFilter]);

  // 处理搜索
  const handleSearch = () => {
    setCurrentPage(1);
    fetchPrompts();
  };

  // 处理编辑
  const handleEdit = (prompt: SystemPrompt) => {
    setEditingPrompt(prompt);
    setIsEditing(true);
    form.setFieldsValue({
      name: prompt.name,
      role: prompt.role,
      content: prompt.content,
      tags: prompt.tags ? (() => {
        try {
          const parsed = JSON.parse(prompt.tags);
          return Array.isArray(parsed) ? parsed.join(',') : '';
        } catch {
          return '';
        }
      })() : '',
      is_active: prompt.is_active
    });
  };

  // 处理创建
  const handleCreate = () => {
    setEditingPrompt(null);
    setIsEditing(true);
    form.resetFields();
    form.setFieldsValue({
      role: 'system',
      is_active: true
    });
  };

  // 处理保存
  const handleSave = async () => {
    try {
      const values = await form.validateFields();
      const tagsArray = values.tags ? values.tags.split(',').map((tag: string) => tag.trim()).filter(Boolean) : [];
      
      if (editingPrompt) {
        // 更新
        const updateData: UpdateSystemPromptRequest = {
          name: values.name,
          role: values.role,
          content: values.content,
          tags: JSON.stringify(tagsArray),
          is_active: values.is_active
        };
        await systemPromptService.updateSystemPrompt(editingPrompt.id, updateData);
        message.success('系统提示词更新成功');
      } else {
        // 创建
        const createData: CreateSystemPromptRequest = {
          name: values.name,
          role: values.role,
          content: values.content,
          tags: JSON.stringify(tagsArray),
          is_active: values.is_active
        };
        await systemPromptService.createSystemPrompt(createData);
        message.success('系统提示词创建成功');
      }
      
      setIsEditing(false);
      setEditingPrompt(null);
      form.resetFields();
      fetchPrompts();
    } catch (error) {
      message.error('保存失败');
    }
  };

  // 处理删除
  const handleDelete = async (id: number) => {
    try {
      await systemPromptService.deleteSystemPrompt(id);
      message.success('删除成功');
      fetchPrompts();
    } catch (error) {
      message.error('删除失败');
    }
  };

  // 处理选择
  const handleSelect = (prompt: SystemPrompt) => {
    onSelect(prompt);
    onClose();
  };

  // 处理复制
  const handleCopy = (content: string) => {
    navigator.clipboard.writeText(content);
    message.success('已复制到剪贴板');
  };

  // 表格列定义
  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 120,
      render: (text: string, record: SystemPrompt) => (
        <div>
          <div className="font-medium">{text}</div>
          {record.id === selectedPromptId && (
            <Tag color="blue">已选择</Tag>
          )}
        </div>
      )
    },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      width: 80,
      render: (role: string) => <Tag color="green">{role}</Tag>
    },
    {
      title: '内容预览',
      dataIndex: 'content',
      key: 'content',
      ellipsis: true,
      render: (content: string) => (
        <Tooltip title={content}>
          <Paragraph 
            ellipsis={{ rows: 2 }} 
            style={{ margin: 0, maxWidth: 300 }}
          >
            {content}
          </Paragraph>
        </Tooltip>
      )
    },
    {
      title: '标签',
      dataIndex: 'tags',
      key: 'tags',
      width: 120,
      render: (tags: string) => {
        if (!tags) return '-';
        try {
          const tagArray = JSON.parse(tags);
          if (!Array.isArray(tagArray)) return '-';
          return (
            <div>
              {tagArray.map((tag: string, index: number) => (
                <Tag key={`${tag}-${index}`}>{tag}</Tag>
              ))}
            </div>
          );
        } catch {
          return '-';
        }
      }
    },
    {
      title: '状态',
      dataIndex: 'is_active',
      key: 'is_active',
      width: 80,
      render: (isActive: boolean) => (
        <Tag color={isActive ? 'green' : 'red'}>
          {isActive ? '激活' : '未激活'}
        </Tag>
      )
    },
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      width: 60,
      render: (version: number) => <Text code>v{version}</Text>
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      render: (_: any, record: SystemPrompt) => (
        <Space size="small">
          <Tooltip title="选择此提示词">
            <Button
              type="primary"
              size="small"
              icon={<CheckOutlined />}
              onClick={() => handleSelect(record)}
            >
              选择
            </Button>
          </Tooltip>
          <Tooltip title="编辑">
            <Button
              size="small"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
            />
          </Tooltip>
          <Tooltip title="复制内容">
            <Button
              size="small"
              icon={<CopyOutlined />}
              onClick={() => handleCopy(record.content)}
            />
          </Tooltip>
          <Popconfirm
            title="确定要删除这个系统提示词吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="删除">
              <Button
                size="small"
                danger
                icon={<DeleteOutlined />}
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      )
    }
  ];

  return (
    <Modal
      title={
        <div className="flex items-center">
          <SettingOutlined className="mr-2" />
          系统提示词管理
        </div>
      }
      open={visible}
      onCancel={onClose}
      width="90%"
      style={{ maxWidth: 1200 }}
      footer={null}
      className="system-prompt-manager-modal"
    >
      <div className="flex flex-col" style={{ maxHeight: '80vh', minHeight: '600px' }}>
        {/* 搜索和筛选区域 */}
        <Card size="small" className="mb-4">
          <Row gutter={[16, 8]} align="middle">
            <Col xs={24} sm={12} md={6}>
              <Input
                placeholder="搜索提示词名称"
                value={searchText}
                onChange={(e) => setSearchText(e.target.value)}
                onPressEnter={handleSearch}
                prefix={<SearchOutlined />}
                allowClear
              />
            </Col>
            <Col xs={12} sm={6} md={4}>
              <Select
                placeholder="角色筛选"
                value={roleFilter}
                onChange={setRoleFilter}
                allowClear
                style={{ width: '100%' }}
              >
                <Option value="system">system</Option>
                <Option value="assistant">assistant</Option>
                <Option value="user">user</Option>
              </Select>
            </Col>
            <Col xs={12} sm={6} md={4}>
              <Select
                placeholder="状态筛选"
                value={activeFilter}
                onChange={setActiveFilter}
                allowClear
                style={{ width: '100%' }}
              >
                <Option value={true}>激活</Option>
                <Option value={false}>未激活</Option>
              </Select>
            </Col>
            <Col xs={12} sm={6} md={4}>
              <Button type="primary" onClick={handleSearch} icon={<SearchOutlined />} block>
                搜索
              </Button>
            </Col>
            <Col xs={12} sm={6} md={6} className="text-right">
              <Button type="primary" onClick={handleCreate} icon={<PlusOutlined />} block>
                新建提示词
              </Button>
            </Col>
          </Row>
        </Card>

        {/* 编辑表单 */}
        {isEditing && (
          <Card size="small" className="mb-4" title={editingPrompt ? '编辑系统提示词' : '新建系统提示词'}>
            <Form form={form} layout="vertical">
              <Row gutter={[16, 16]}>
                <Col xs={24} sm={12} md={8}>
                  <Form.Item
                    name="name"
                    label="名称"
                    rules={[{ required: true, message: '请输入提示词名称' }]}
                  >
                    <Input placeholder="提示词名称" />
                  </Form.Item>
                </Col>
                <Col xs={12} sm={6} md={4}>
                  <Form.Item
                    name="role"
                    label="角色"
                    rules={[{ required: true, message: '请选择角色' }]}
                  >
                    <Select>
                      <Option value="system">system</Option>
                      <Option value="assistant">assistant</Option>
                      <Option value="user">user</Option>
                    </Select>
                  </Form.Item>
                </Col>
                <Col xs={12} sm={6} md={4}>
                  <Form.Item
                    name="is_active"
                    label="激活状态"
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>
                </Col>
                <Col xs={24} sm={12} md={8}>
                  <Form.Item
                    name="tags"
                    label="标签"
                  >
                    <Input placeholder="用逗号分隔多个标签" />
                  </Form.Item>
                </Col>
              </Row>
              <Form.Item
                name="content"
                label="内容"
                rules={[{ required: true, message: '请输入提示词内容' }]}
              >
                <TextArea
                  rows={4}
                  placeholder="请输入系统提示词内容..."
                  maxLength={2000}
                  showCount
                />
              </Form.Item>
              <div className="text-right">
                <Space>
                  <Button onClick={() => setIsEditing(false)}>
                    取消
                  </Button>
                  <Button type="primary" onClick={handleSave}>
                    保存
                  </Button>
                </Space>
              </div>
            </Form>
          </Card>
        )}

        {/* 表格区域 */}
        <div className="flex-1" style={{ minHeight: '300px', overflow: 'hidden' }}>
          <Table
            columns={columns}
            dataSource={prompts}
            loading={loading}
            rowKey="id"
            pagination={false}
            scroll={{ y: 250 }}
            locale={{
              emptyText: (
                <Empty
                  image={<SettingOutlined className="text-4xl text-gray-300" />}
                  description="暂无系统提示词"
                />
              )
            }}
          />
        </div>

        {/* 分页 */}
        {total > 0 && (
          <div className="mt-4 text-center">
            <Pagination
              current={currentPage}
              pageSize={pageSize}
              total={total}
              showSizeChanger
              showQuickJumper
              showTotal={(total, range) => `${range[0]}-${range[1]} / ${total}`}
              onChange={(page, size) => {
                setCurrentPage(page);
                setPageSize(size || 10);
              }}
            />
          </div>
        )}
      </div>
    </Modal>
  );
};

export default SystemPromptManagerModal;
