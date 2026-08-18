import React, { useState, useEffect } from 'react';
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
  Tabs,
  ColorPicker,
  Switch
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  SearchOutlined,
  UserOutlined,
  SettingOutlined,
  TagsOutlined,
  RobotOutlined,
} from '@ant-design/icons';
import { keywordService } from '@/services/keywords';
import { userService } from '@/services/users';
import { systemPromptService } from '@/services/systemPrompts';
import { Keyword, KeywordQueryParams, User, UserQueryParams, SystemPrompt, SystemPromptQueryParams } from '@/types';
import { formatDate } from '@/utils/dateUtils';
import { normalizeColorValue } from '@/utils/colorUtils';
import { useAuthStore } from '@/stores/authStore';

const SettingsPage: React.FC = () => {
  // 关键词列表状态
  const [keywords, setKeywords] = useState<Keyword[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [searchParams, setSearchParams] = useState<KeywordQueryParams>({
    offset: 0,
    limit: 10,
  });
  // 搜索输入状态（用于输入框显示）
  const [searchInput, setSearchInput] = useState({
    keyword: '',
    mark_color: '',
  });

  // 用户列表状态
  const [users, setUsers] = useState<User[]>([]);
  const [userLoading, setUserLoading] = useState(false);
  const [userTotal, setUserTotal] = useState(0);
  const [userCurrentPage, setUserCurrentPage] = useState(1);
  const [userPageSize, setUserPageSize] = useState(20);
  const [userSearchParams, setUserSearchParams] = useState<UserQueryParams>({
    offset: 0,
    limit: 20,
  });
  // 用户搜索输入状态（用于输入框显示）
  const [userSearchInput, setUserSearchInput] = useState({
    username: '',
  });

  // 系统提示词列表状态
  const [systemPrompts, setSystemPrompts] = useState<SystemPrompt[]>([]);
  const [systemPromptLoading, setSystemPromptLoading] = useState(false);
  const [systemPromptTotal, setSystemPromptTotal] = useState(0);
  const [systemPromptCurrentPage, setSystemPromptCurrentPage] = useState(1);
  const [systemPromptPageSize, setSystemPromptPageSize] = useState(10);
  const [systemPromptSearchParams, setSystemPromptSearchParams] = useState<SystemPromptQueryParams>({
    offset: 0,
    limit: 10,
  });
  // 系统提示词搜索输入状态（用于输入框显示）
  const [systemPromptSearchInput, setSystemPromptSearchInput] = useState({
    name: '',
  });

  // 模态框状态
  const [modalVisible, setModalVisible] = useState(false);
  const [editingKeyword, setEditingKeyword] = useState<Keyword | null>(null);
  const [form] = Form.useForm();

  // 用户模态框状态
  const [userModalVisible, setUserModalVisible] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [userForm] = Form.useForm();

  // 密码修改模态框状态
  const [passwordModalVisible, setPasswordModalVisible] = useState(false);
  const [editingPasswordUser, setEditingPasswordUser] = useState<User | null>(null);
  const [passwordForm] = Form.useForm();

  // 系统提示词模态框状态
  const [systemPromptModalVisible, setSystemPromptModalVisible] = useState(false);
  const [editingSystemPrompt, setEditingSystemPrompt] = useState<SystemPrompt | null>(null);
  const [systemPromptForm] = Form.useForm();

  // 系统提示词批量选择状态
  const [selectedSystemPromptRowKeys, setSelectedSystemPromptRowKeys] = useState<React.Key[]>([]);

  // 当前用户信息
  const { user: currentUser } = useAuthStore();

  useEffect(() => {
    fetchKeywords();
  }, [searchParams]);

  useEffect(() => {
    fetchUsers();
  }, [userSearchParams]);

  useEffect(() => {
    fetchSystemPrompts();
  }, [systemPromptSearchParams]);

  // 获取关键词列表
  const fetchKeywords = async () => {
    try {
      setLoading(true);
      const response = await keywordService.getKeywords(searchParams);
      setKeywords(response.items);
      setTotal(response.total);
    } catch (error) {
      message.error('获取关键词列表失败');
    } finally {
      setLoading(false);
    }
  };

  // 获取用户列表
  const fetchUsers = async () => {
    try {
      setUserLoading(true);
      const response = await userService.getUsers(userSearchParams);
      setUsers(response.items);
      setUserTotal(response.total);
    } catch (error) {
      message.error('获取用户列表失败');
    } finally {
      setUserLoading(false);
    }
  };

  // 关键词搜索处理
  const handleKeywordSearch = () => {
    setSearchParams({
      offset: 0,
      limit: 10,
      keyword: searchInput.keyword || undefined,
      mark_color: searchInput.mark_color || undefined,
    });
    setCurrentPage(1);
  };

  // 关键词重置搜索
  const handleKeywordReset = () => {
    setSearchInput({
      keyword: '',
      mark_color: '',
    });
    setSearchParams({
      offset: 0,
      limit: 10,
    });
    setCurrentPage(1);
  };

  // 用户搜索处理
  const handleUserSearch = () => {
    setUserSearchParams({
      offset: 0,
      limit: 20,
      username: userSearchInput.username || undefined,
    });
    setUserCurrentPage(1);
  };

  // 用户重置搜索
  const handleUserReset = () => {
    setUserSearchInput({
      username: '',
    });
    setUserSearchParams({
      offset: 0,
      limit: 20,
    });
    setUserCurrentPage(1);
  };

  // 关键词分页处理
  const handleKeywordPageChange = (page: number, size: number) => {
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

  // 用户分页处理
  const handleUserPageChange = (page: number, size: number) => {
    const limitedSize = Math.min(size, 100);
    const offset = (page - 1) * limitedSize;
    setUserSearchParams(prev => ({
      ...prev,
      offset,
      limit: limitedSize,
    }));
    setUserCurrentPage(page);
    setUserPageSize(limitedSize);
  };

  // 获取系统提示词列表
  const fetchSystemPrompts = async () => {
    try {
      setSystemPromptLoading(true);
      const response = await systemPromptService.getSystemPrompts(systemPromptSearchParams);
      setSystemPrompts(response.items);
      setSystemPromptTotal(response.total);
    } catch (error) {
      message.error('获取系统提示词列表失败');
    } finally {
      setSystemPromptLoading(false);
    }
  };

  // 系统提示词分页处理
  const handleSystemPromptPageChange = (page: number, size: number) => {
    const limitedSize = Math.min(size, 100);
    const offset = (page - 1) * limitedSize;
    setSystemPromptSearchParams(prev => ({
      ...prev,
      offset,
      limit: limitedSize,
    }));
    setSystemPromptCurrentPage(page);
    setSystemPromptPageSize(limitedSize);
  };

  // 添加关键词
  const handleAddKeyword = () => {
    setEditingKeyword(null);
    form.resetFields();
    // 设置默认颜色为红色
    form.setFieldsValue({
      mark_color: '#ff4d4f'
    });
    setModalVisible(true);
  };

  // 编辑关键词
  const handleEditKeyword = (keyword: Keyword) => {
    setEditingKeyword(keyword);
    form.setFieldsValue({
      keyword: keyword.keyword,
      synonyms: keyword.synonyms,
      mark_color: keyword.mark_color, // 确保是字符串格式
    });
    setModalVisible(true);
  };

  // 删除关键词
  const handleDeleteKeyword = async (id: number) => {
    try {
      await keywordService.deleteKeyword(id);
      message.success('删除成功');
      fetchKeywords();
    } catch (error) {
      message.error('删除失败');
    }
  };

  // 提交关键词信息
  const handleKeywordModalOk = async () => {
    try {
      const values = await form.validateFields();
      
      // 确保mark_color是字符串格式
      const markColor = normalizeColorValue(values.mark_color);
      
      if (editingKeyword) {
        // 编辑关键词
        await keywordService.updateKeyword(editingKeyword.id, {
          keyword: values.keyword,
          synonyms: values.synonyms,
          mark_color: markColor,
        });
        message.success('更新成功');
      } else {
        // 添加关键词
        await keywordService.createKeyword({
          keyword: values.keyword,
          synonyms: values.synonyms,
          mark_color: markColor,
        });
        message.success('创建成功');
      }
      
      setModalVisible(false);
      fetchKeywords();
    } catch (error) {
      message.error('操作失败');
    }
  };

  // 添加用户
  const handleAddUser = () => {
    setEditingUser(null);
    userForm.resetFields();
    setUserModalVisible(true);
  };

  // 编辑用户
  const handleEditUser = (user: User) => {
    setEditingUser(user);
    userForm.setFieldsValue({
      username: user.username,
    });
    setUserModalVisible(true);
  };

  // 删除用户
  const handleDeleteUser = async (id: number) => {
    if (currentUser?.id === id) {
      message.error('不能删除当前登录的用户');
      return;
    }

    try {
      await userService.deleteUser(id);
      message.success('删除成功');
      fetchUsers();
    } catch (error) {
      message.error('删除失败');
    }
  };

  // 修改密码
  const handleChangePassword = (user: User) => {
    setEditingPasswordUser(user);
    passwordForm.resetFields();
    setPasswordModalVisible(true);
  };

  // 提交用户信息
  const handleUserModalOk = async () => {
    try {
      const values = await userForm.validateFields();
      
      if (editingUser) {
        // 编辑用户
        await userService.updateUser(editingUser.id, { username: values.username });
        message.success('更新成功');
      } else {
        // 添加用户
        await userService.register({
          username: values.username,
          password: values.password,
        });
        message.success('创建成功');
      }
      
      setUserModalVisible(false);
      fetchUsers();
    } catch (error: any) {
      if (error?.message?.includes('username already exists')) {
        message.error('用户名已存在');
      } else {
        message.error('操作失败');
      }
    }
  };

  // 提交密码修改
  const handlePasswordModalOk = async () => {
    try {
      const values = await passwordForm.validateFields();
      
      if (!editingPasswordUser) return;

      await userService.changePassword(editingPasswordUser.id, {
        old_password: values.old_password,
        new_password: values.new_password,
      });
      
      message.success('密码修改成功');
      setPasswordModalVisible(false);
    } catch (error) {
      message.error('密码修改失败，请检查原密码是否正确');
    }
  };

  // 批量操作状态
  const [selectedKeywordKeys, setSelectedKeywordKeys] = useState<React.Key[]>([]);
  const [selectedUserKeys, setSelectedUserKeys] = useState<React.Key[]>([]);
  const [batchLoading, setBatchLoading] = useState(false);

  // 批量删除关键词
  const handleBatchDeleteKeywords = async () => {
    if (selectedKeywordKeys.length === 0) {
      message.warning('请选择要删除的关键词');
      return;
    }

    try {
      setBatchLoading(true);
      const result = await keywordService.batchDeleteKeywords(selectedKeywordKeys as number[]);
      message.success(`成功删除 ${result.deleted_count} 个关键词`);
      setSelectedKeywordKeys([]);
      fetchKeywords();
    } catch (error) {
      message.error('批量删除失败');
    } finally {
      setBatchLoading(false);
    }
  };

  // 批量删除用户
  const handleBatchDeleteUsers = async () => {
    if (selectedUserKeys.length === 0) {
      message.warning('请选择要删除的用户');
      return;
    }

    // 检查是否包含当前用户
    if (currentUser && selectedUserKeys.includes(currentUser.id)) {
      message.error('不能删除当前登录的用户');
      return;
    }

    try {
      setBatchLoading(true);
      await userService.batchDeleteUsers(selectedUserKeys as number[]);
      message.success(`成功删除 ${selectedUserKeys.length} 个用户`);
      setSelectedUserKeys([]);
      fetchUsers();
    } catch (error) {
      message.error('批量删除失败');
    } finally {
      setBatchLoading(false);
    }
  };

  // 关键词行选择配置
  const keywordRowSelection = {
    selectedRowKeys: selectedKeywordKeys,
    onChange: (newSelectedRowKeys: React.Key[]) => {
      setSelectedKeywordKeys(newSelectedRowKeys);
    },
  };

  // 用户行选择配置
  const userRowSelection = {
    selectedRowKeys: selectedUserKeys,
    onChange: (newSelectedRowKeys: React.Key[]) => {
      setSelectedUserKeys(newSelectedRowKeys);
    },
    getCheckboxProps: (record: User) => ({
      disabled: currentUser?.id === record.id, // 禁止选择当前用户
      name: record.username,
    }),
  };

  // 关键词表格列定义
  const keywordColumns = [
    {
      title: '序号',
      dataIndex: 'id',
      key: 'id',
      width: 80,
      render: (_: any, __: any, index: number) => index + 1 + (currentPage - 1) * pageSize,
    },
    {
      title: '关键词',
      dataIndex: 'keyword',
      key: 'keyword',
      width: 200,
      render: (text: string) => (
        <span className="font-medium text-primary-700">{text}</span>
      ),
    },
    {
      title: '近义词',
      dataIndex: 'synonyms',
      key: 'synonyms',
      width: 300,
      render: (text: string) => (
        <span className="text-gray-600">{text}</span>
      ),
    },
    {
      title: '标记颜色',
      dataIndex: 'mark_color',
      key: 'mark_color',
      width: 120,
      render: (color: string) => (
        <div className="flex items-center space-x-2">
          <div 
            className="w-6 h-3 rounded"
            style={{ backgroundColor: color }}
          />
          <span className="text-sm text-gray-600">{color}</span>
        </div>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_: any, record: Keyword) => (
        <Space size="small">
          <Tooltip title="编辑关键词">
            <Button
              type="text"
              icon={<EditOutlined />}
              size="small"
              className="text-primary-600 hover:text-primary-700 hover:bg-primary-50"
              onClick={() => handleEditKeyword(record)}
            />
          </Tooltip>
          <Popconfirm
            title="确定要删除这个关键词吗？"
            onConfirm={() => handleDeleteKeyword(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="删除关键词">
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

  // 添加系统提示词
  const handleAddSystemPrompt = () => {
    setEditingSystemPrompt(null);
    systemPromptForm.resetFields();
    systemPromptForm.setFieldsValue({
      role: 'system',
      is_active: false
    });
    setSystemPromptModalVisible(true);
  };

  // 编辑系统提示词
  const handleEditSystemPrompt = (systemPrompt: SystemPrompt) => {
    setEditingSystemPrompt(systemPrompt);
    systemPromptForm.setFieldsValue({
      name: systemPrompt.name,
      content: systemPrompt.content,
      tags: systemPrompt.tags ? (() => {
        try {
          const parsed = JSON.parse(systemPrompt.tags);
          return Array.isArray(parsed) ? parsed.join(',') : '';
        } catch {
          return '';
        }
      })() : '',
      is_active: systemPrompt.is_active
    });
    setSystemPromptModalVisible(true);
  };

  // 删除系统提示词
  const handleDeleteSystemPrompt = async (id: number) => {
    try {
      await systemPromptService.deleteSystemPrompt(id);
      message.success('删除成功');
      fetchSystemPrompts();
    } catch (error) {
      message.error('删除失败');
    }
  };

  // 系统提示词模态框确定
  const handleSystemPromptModalOk = async () => {
    try {
      const values = await systemPromptForm.validateFields();
      console.log('表单值:', values); // 调试信息
      const tagsArray = values.tags ? values.tags.split(',').map((tag: string) => tag.trim()).filter(Boolean) : [];
      
      if (editingSystemPrompt) {
        // 更新
        console.log('更新数据:', { ...values, is_active: values.is_active }); // 调试信息
        await systemPromptService.updateSystemPrompt(editingSystemPrompt.id, {
          name: values.name,
          role: 'system',
          content: values.content,
          tags: JSON.stringify(tagsArray),
          is_active: values.is_active
        });
        message.success('系统提示词更新成功');
      } else {
        // 创建
        console.log('创建数据:', { ...values, is_active: values.is_active }); // 调试信息
        await systemPromptService.createSystemPrompt({
          name: values.name,
          role: 'system',
          content: values.content,
          tags: JSON.stringify(tagsArray),
          is_active: values.is_active
        });
        message.success('系统提示词创建成功');
      }
      
      setSystemPromptModalVisible(false);
      setEditingSystemPrompt(null);
      systemPromptForm.resetFields();
      fetchSystemPrompts();
    } catch (error: any) {
      // 尝试从错误响应中获取具体错误信息
      let errorMessage = '保存失败';
      if (error?.message) {
        errorMessage = error.message;
      } else if (error?.response?.data?.message) {
        errorMessage = error.response.data.message;
      }
      message.error(errorMessage);
    }
  };

  // 系统提示词搜索
  const handleSystemPromptSearch = async () => {
    if (!systemPromptSearchInput.name.trim()) {
      // 如果搜索框为空，使用普通列表接口
      setSystemPromptSearchParams(prev => ({
        ...prev,
        name: undefined,
        offset: 0,
      }));
      setSystemPromptCurrentPage(1);
      return;
    }

    try {
      setSystemPromptLoading(true);
      const searchResults = await systemPromptService.searchSystemPrompts(
        systemPromptSearchInput.name.trim(),
        100 // 搜索时获取更多结果
      );
      setSystemPrompts(searchResults);
      setSystemPromptTotal(searchResults.length);
      setSystemPromptCurrentPage(1);
    } catch (error) {
      message.error('搜索失败');
    } finally {
      setSystemPromptLoading(false);
    }
  };

  // 批量删除系统提示词
  const handleBatchDeleteSystemPrompts = async () => {
    if (selectedSystemPromptRowKeys.length === 0) {
      message.warning('请选择要删除的系统提示词');
      return;
    }

    try {
      // 使用批量删除接口
      await systemPromptService.batchDeleteSystemPrompts(selectedSystemPromptRowKeys.map(Number));
      message.success(`成功删除 ${selectedSystemPromptRowKeys.length} 个系统提示词`);
      setSelectedSystemPromptRowKeys([]);
      fetchSystemPrompts();
    } catch (error) {
      message.error('批量删除失败');
    }
  };

  // 更新系统提示词状态
  const handleUpdateSystemPromptStatus = async (id: number, isActive: boolean) => {
    try {
      await systemPromptService.updateSystemPromptStatus(id, isActive);
      message.success(`系统提示词已${isActive ? '启用' : '停用'}`);
      fetchSystemPrompts();
    } catch (error) {
      message.error('更新状态失败');
    }
  };

  // 系统提示词行选择配置
  const systemPromptRowSelection = {
    selectedRowKeys: selectedSystemPromptRowKeys,
    onChange: (selectedRowKeys: React.Key[]) => {
      setSelectedSystemPromptRowKeys(selectedRowKeys);
    },
    getCheckboxProps: (record: SystemPrompt) => ({
      name: record.name,
    }),
  };

  // 用户表格列定义
  const userColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
      sorter: (a: User, b: User) => a.id - b.id,
    },
    {
      title: '头像',
      dataIndex: 'username',
      key: 'avatar',
      width: 80,
      render: () => (
        <div className="w-8 h-8 bg-primary-500 rounded-full flex items-center justify-center">
          <UserOutlined className="text-white text-sm" />
        </div>
      ),
    },
    {
      title: '用户名',
      dataIndex: 'username',
      key: 'username',
      width: 200,
      render: (text: string, record: User) => (
        <div className="flex items-center space-x-2">
          <span className="font-medium text-primary-700">{text}</span>
          {currentUser?.id === record.id && (
            <Tag color="green">当前用户</Tag>
          )}
        </div>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 200,
      render: (time: number) => (
        <span className="text-gray-600">
          {formatDate(time)}
        </span>
      ),
      sorter: (a: User, b: User) => a.created_at - b.created_at,
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_: any, record: User) => (
        <Space size="small">
          <Tooltip title="编辑用户">
            <Button
              type="text"
              icon={<EditOutlined />}
              size="small"
              className="text-primary-600 hover:text-primary-700 hover:bg-primary-50"
              onClick={() => handleEditUser(record)}
            />
          </Tooltip>
          <Tooltip title="修改密码">
            <Button
              type="text"
              icon={<SettingOutlined />}
              size="small"
              className="text-orange-600 hover:text-orange-700 hover:bg-orange-50"
              onClick={() => handleChangePassword(record)}
            />
          </Tooltip>
          <Popconfirm
            title="确定要删除这个用户吗？"
            description={currentUser?.id === record.id ? "不能删除当前登录的用户" : undefined}
            onConfirm={() => handleDeleteUser(record.id)}
            okText="确定"
            cancelText="取消"
            disabled={currentUser?.id === record.id}
          >
            <Tooltip title={currentUser?.id === record.id ? "不能删除当前用户" : "删除用户"}>
              <Button
                type="text"
                icon={<DeleteOutlined />}
                size="small"
                className="text-red-500 hover:text-red-600 hover:bg-red-50"
                disabled={currentUser?.id === record.id}
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  // 系统提示词表格列定义
  const systemPromptColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 120,
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 150,
      render: (text: string) => (
        <Tooltip title={text}>
          <span className="font-medium text-primary-700">
            {text.length > 7 ? text.substring(0, 7) + '...' : text}
          </span>
        </Tooltip>
      ),
    },
    {
      title: '内容',
      dataIndex: 'content',
      key: 'content',
      ellipsis: true,
      render: (text: string) => (
        <Tooltip title={text}>
          <span className="text-gray-600">{text}</span>
        </Tooltip>
      ),
    },
    {
      title: '标签',
      dataIndex: 'tags',
      key: 'tags',
      width: 150,
      render: (tags: string) => {
        try {
          const parsed = JSON.parse(tags);
          return Array.isArray(parsed) ? (
            <div className="flex flex-wrap gap-1">
              {parsed.map((tag: string, index: number) => (
                <Tag key={index} bordered={false}>{tag}</Tag>
              ))}
            </div>
          ) : null;
        } catch {
          return null;
        }
      },
    },
    {
      title: '状态',
      dataIndex: 'is_active',
      key: 'is_active',
      width: 100,
      render: (isActive: boolean, record: SystemPrompt) => (
        <Switch 
          checked={isActive} 
          checkedChildren="启用" 
          unCheckedChildren="停用"
          onChange={(checked) => handleUpdateSystemPromptStatus(record.id, checked)}
        />
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 150,
      render: (timestamp: number) => formatDate(timestamp),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_: any, record: SystemPrompt) => (
        <Space size="small">
          <Tooltip title="编辑系统提示词">
            <Button
              type="text"
              icon={<EditOutlined />}
              size="small"
              className="text-primary-600 hover:text-primary-700 hover:bg-primary-50"
              onClick={() => handleEditSystemPrompt(record)}
            />
          </Tooltip>
          <Popconfirm
            title="确定要删除这个系统提示词吗？"
            onConfirm={() => handleDeleteSystemPrompt(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="删除系统提示词">
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

  // 关键词管理标签页内容
  const keywordTabContent = (
    <div className="space-y-6">
      {/* 搜索和操作区域 */}
      <Card className="card-primary">
        <div className="space-y-4">
          <Row gutter={16}>
            <Col span={8}>
              <Input
                placeholder="搜索关键词"
                value={searchInput.keyword}
                onChange={(e) => setSearchInput(prev => ({ ...prev, keyword: e.target.value }))}
                prefix={<SearchOutlined className="text-primary-400" />}
                className="input-primary"
                allowClear
              />
            </Col>
            <Col span={16}>
              <Space>
                <Button
                  type="primary"
                  icon={<SearchOutlined />}
                  onClick={handleKeywordSearch}
                  className="btn-primary"
                >
                  搜索
                </Button>
                <Button
                  icon={<SearchOutlined />}
                  onClick={handleKeywordReset}
                  className="btn-secondary"
                >
                  重置
                </Button>
                <Button
                  type="primary"
                  icon={<PlusOutlined />}
                  onClick={handleAddKeyword}
                  className="btn-primary"
                >
                  新增关键词
                </Button>
                {selectedKeywordKeys.length > 0 && (
                  <Popconfirm
                    title={`确定要删除选中的 ${selectedKeywordKeys.length} 个关键词吗？`}
                    onConfirm={handleBatchDeleteKeywords}
                    okText="确定"
                    cancelText="取消"
                  >
                    <Button
                      danger
                      icon={<DeleteOutlined />}
                      loading={batchLoading}
                    >
                      批量删除 ({selectedKeywordKeys.length})
                    </Button>
                  </Popconfirm>
                )}
              </Space>
            </Col>
          </Row>
        </div>
      </Card>

      {/* 关键词列表表格 */}
      <Card className="card-primary">
        <Table
          columns={keywordColumns}
          dataSource={keywords}
          loading={loading}
          rowKey="id"
          rowSelection={keywordRowSelection}
          pagination={{
            current: currentPage,
            pageSize: pageSize,
            total: total,
            showSizeChanger: true,
            showQuickJumper: true,
            pageSizeOptions: ['10', '20', '50'],
            showTotal: (total) => 
              `共${total}条`,
            onChange: handleKeywordPageChange,
            onShowSizeChange: handleKeywordPageChange,
          }}
          className="table-primary"
          rowClassName="table-row-hover"
        />
      </Card>
    </div>
  );

  // 用户管理标签页内容
  const userTabContent = (
    <div className="space-y-6">
      {/* 用户统计卡片 */}
      <Row gutter={16}>
        <Col span={12}>
          <Card className="card-primary">
            <Statistic
              title="总用户数"
              value={userTotal}
              valueStyle={{ color: '#8CC020' }}
              prefix={<UserOutlined />}
            />
          </Card>
        </Col>
        <Col span={12}>
          <Card className="card-primary">
            <Statistic
              title="管理员用户"
              value={userTotal}
              valueStyle={{ color: '#8CC020' }}
              prefix={<UserOutlined />}
            />
          </Card>
        </Col>
      </Row>

      {/* 搜索和操作区域 */}
      <Card className="card-primary">
        <div className="space-y-4">
          <Row gutter={16}>
            <Col span={8}>
              <Input
                placeholder="搜索用户名"
                value={userSearchInput.username}
                onChange={(e) => setUserSearchInput(prev => ({ ...prev, username: e.target.value }))}
                prefix={<SearchOutlined className="text-primary-400" />}
                className="input-primary"
                allowClear
              />
            </Col>
            <Col span={16}>
              <Space>
                <Button
                  type="primary"
                  icon={<SearchOutlined />}
                  onClick={handleUserSearch}
                  className="btn-primary"
                >
                  搜索
                </Button>
                <Button
                  icon={<SearchOutlined />}
                  onClick={handleUserReset}
                  className="btn-secondary"
                >
                  重置
                </Button>
                <Button
                  type="primary"
                  icon={<PlusOutlined />}
                  onClick={handleAddUser}
                  className="btn-primary"
                >
                  新增用户
                </Button>
                {selectedUserKeys.length > 0 && (
                  <Popconfirm
                    title={`确定要删除选中的 ${selectedUserKeys.length} 个用户吗？`}
                    onConfirm={handleBatchDeleteUsers}
                    okText="确定"
                    cancelText="取消"
                  >
                    <Button
                      danger
                      icon={<DeleteOutlined />}
                      loading={batchLoading}
                    >
                      批量删除 ({selectedUserKeys.length})
                    </Button>
                  </Popconfirm>
                )}
              </Space>
            </Col>
          </Row>
        </div>
      </Card>

      {/* 用户列表表格 */}
      <Card className="card-primary">
        <Table
          columns={userColumns}
          dataSource={users}
          loading={userLoading}
          rowKey="id"
          rowSelection={userRowSelection}
          pagination={{
            current: userCurrentPage,
            pageSize: userPageSize,
            total: userTotal,
            showSizeChanger: true,
            showQuickJumper: true,
            pageSizeOptions: ['20', '50', '100'],
            showTotal: (total, range) => 
              `第 ${range[0]}-${range[1]} 条，共 ${total} 条记录`,
            onChange: handleUserPageChange,
            onShowSizeChange: handleUserPageChange,
          }}
          className="table-primary"
          rowClassName="table-row-hover"
        />
      </Card>
    </div>
  );

  // 系统提示词管理标签页内容
  const systemPromptTabContent = (
    <div className="space-y-6">
      {/* 统计信息 */}
      <Row gutter={16}>
        <Col span={12}>
          <Card className="card-primary">
            <Statistic
              title="系统提示词总数"
              value={systemPromptTotal}
              valueStyle={{ color: '#8CC020' }}
              prefix={<RobotOutlined />}
            />
          </Card>
        </Col>
        <Col span={12}>
          <Card className="card-primary">
            <Statistic
              title="激活提示词"
              value={systemPrompts.filter(p => p.is_active).length}
              valueStyle={{ color: '#8CC020' }}
              prefix={<RobotOutlined />}
            />
          </Card>
        </Col>
      </Row>

      {/* 搜索和操作区域 */}
      <Card className="card-primary">
        <div className="space-y-4">
          <Row gutter={16}>
            <Col span={8}>
              <Input
                placeholder="搜索提示词名称"
                value={systemPromptSearchInput.name}
                onChange={(e) => setSystemPromptSearchInput(prev => ({ ...prev, name: e.target.value }))}
                prefix={<SearchOutlined className="text-primary-400" />}
                className="input-primary"
                allowClear
              />
            </Col>
            <Col span={16}>
              <Space>
                <Button
                  type="primary"
                  icon={<SearchOutlined />}
                  onClick={handleSystemPromptSearch}
                  className="btn-primary"
                >
                  搜索
                </Button>
                <Button
                  icon={<SearchOutlined />}
                  onClick={() => {
                    setSystemPromptSearchInput({ name: '' });
                    setSystemPromptSearchParams(prev => ({
                      ...prev,
                      name: undefined,
                      offset: 0,
                    }));
                    setSystemPromptCurrentPage(1);
                    fetchSystemPrompts(); // 重置后重新获取列表
                  }}
                  className="btn-secondary"
                >
                  重置
                </Button>
                <Button
                  type="primary"
                  icon={<PlusOutlined />}
                  onClick={handleAddSystemPrompt}
                  className="btn-primary"
                >
                  新增提示词
                </Button>
                <Popconfirm
                  title="确定要删除选中的系统提示词吗？"
                  description={`将删除 ${selectedSystemPromptRowKeys.length} 个系统提示词，此操作不可撤销`}
                  onConfirm={handleBatchDeleteSystemPrompts}
                  okText="确定删除"
                  cancelText="取消"
                  disabled={selectedSystemPromptRowKeys.length === 0}
                >
                  <Button
                    danger
                    icon={<DeleteOutlined />}
                    disabled={selectedSystemPromptRowKeys.length === 0}
                    className="btn-danger"
                  >
                    批量删除 ({selectedSystemPromptRowKeys.length})
                  </Button>
                </Popconfirm>
              </Space>
            </Col>
          </Row>
        </div>
      </Card>

      {/* 系统提示词列表表格 */}
      <Card className="card-primary">
        <Table
          columns={systemPromptColumns}
          dataSource={systemPrompts}
          loading={systemPromptLoading}
          rowKey="id"
          rowSelection={systemPromptRowSelection}
          pagination={{
            current: systemPromptCurrentPage,
            pageSize: systemPromptPageSize,
            total: systemPromptTotal,
            showSizeChanger: true,
            showQuickJumper: true,
            pageSizeOptions: ['10', '20', '50'],
            showTotal: (total, range) => 
              `第 ${range[0]}-${range[1]} 条，共 ${total} 条记录`,
            onChange: handleSystemPromptPageChange,
            onShowSizeChange: handleSystemPromptPageChange,
          }}
          className="table-primary"
          rowClassName="table-row-hover"
        />
      </Card>
    </div>
  );

  return (
    <div className="space-y-6">
      {/* 标签页 */}
      <Card className="card-primary">
        <Tabs
          defaultActiveKey="keywords"
          items={[
            {
              key: 'keywords',
              label: (
                <span>
                  <TagsOutlined />
                  关键词
                </span>
              ),
              children: keywordTabContent,
            },
            {
              key: 'users',
              label: (
                <span>
                  <UserOutlined />
                  用户管理
                </span>
              ),
              children: userTabContent,
            },
            {
              key: 'system-prompts',
              label: (
                <span>
                  <RobotOutlined />
                  系统提示词
                </span>
              ),
              children: systemPromptTabContent,
            },
          ]}
        />
      </Card>

      {/* 新增/编辑关键词模态框 */}
      <Modal
        title={
          <div className="flex items-center space-x-2">
            <TagsOutlined className="text-primary-500" />
            <span>{editingKeyword ? '编辑关键词' : '新增关键词'}</span>
          </div>
        }
        open={modalVisible}
        onOk={handleKeywordModalOk}
        onCancel={() => setModalVisible(false)}
        width={500}
        okText="确定"
        cancelText="取消"
      >
        <Form
          form={form}
          layout="vertical"
          className="mt-4"
        >
          <Form.Item
            name="keyword"
            label="关键词"
            rules={[
              { required: true, message: '请输入关键词' },
              { min: 1, message: '关键词不能为空' },
              { max: 50, message: '关键词最多50个字符' },
            ]}
          >
            <Input 
              placeholder="请输入关键词" 
              className="input-primary"
            />
          </Form.Item>
          
          <Form.Item
            name="synonyms"
            label="近义词"
            rules={[
              { required: true, message: '请输入近义词' },
            ]}
            extra="多个近义词用逗号分隔"
          >
            <Input.TextArea 
              placeholder="请输入近义词，多个用逗号分隔" 
              className="input-primary"
              rows={3}
            />
          </Form.Item>

          <Form.Item
            name="mark_color"
            label="标记颜色"
            rules={[
              { required: true, message: '请选择标记颜色' },
            ]}
            extra="点击色块选择颜色，或输入十六进制颜色值"
          >
            <ColorPicker
              showText
              format="hex"
              className="w-full"
              onChange={(color) => {
                // 确保颜色值始终是字符串格式
                const colorValue = normalizeColorValue(color);
                form.setFieldValue('mark_color', colorValue);
              }}
              presets={[
                {
                  label: '推荐颜色',
                  colors: [
                    '#ff4d4f', '#52c41a', '#1890ff', '#fa8c16', 
                    '#722ed1', '#13c2c2', '#eb2f96', '#faad14',
                    '#f5222d', '#389e0d', '#0958d9', '#d46b08',
                    '#531dab', '#08979c', '#c41d7f', '#d48806'
                  ],
                },
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>

      {/* 新增/编辑用户模态框 */}
      <Modal
        title={
          <div className="flex items-center space-x-2">
            <UserOutlined className="text-primary-500" />
            <span>{editingUser ? '编辑用户' : '新增用户'}</span>
          </div>
        }
        open={userModalVisible}
        onOk={handleUserModalOk}
        onCancel={() => setUserModalVisible(false)}
        width={500}
        okText="确定"
        cancelText="取消"
      >
        <Form
          form={userForm}
          layout="vertical"
          className="mt-4"
        >
          <Form.Item
            name="username"
            label="用户名"
            rules={[
              { required: true, message: '请输入用户名' },
              { min: 3, message: '用户名至少3个字符' },
              { max: 20, message: '用户名最多20个字符' },
              { pattern: /^[a-zA-Z0-9_]+$/, message: '用户名只能包含字母、数字和下划线' },
            ]}
          >
            <Input 
              placeholder="请输入用户名" 
              className="input-primary"
              prefix={<UserOutlined className="text-primary-400" />}
            />
          </Form.Item>
          
          {!editingUser && (
            <Form.Item
              name="password"
              label="密码"
              rules={[
                { required: true, message: '请输入密码' },
                { min: 6, message: '密码至少6个字符' },
                { max: 20, message: '密码最多20个字符' },
              ]}
            >
              <Input.Password 
                placeholder="请输入密码" 
                className="input-primary"
              />
            </Form.Item>
          )}

          {!editingUser && (
            <Form.Item
              name="confirmPassword"
              label="确认密码"
              dependencies={['password']}
              rules={[
                { required: true, message: '请确认密码' },
                ({ getFieldValue }) => ({
                  validator(_, value) {
                    if (!value || getFieldValue('password') === value) {
                      return Promise.resolve();
                    }
                    return Promise.reject(new Error('两次输入的密码不一致'));
                  },
                }),
              ]}
            >
              <Input.Password 
                placeholder="请再次输入密码" 
                className="input-primary"
              />
            </Form.Item>
          )}
        </Form>
      </Modal>

      {/* 修改密码模态框 */}
      <Modal
        title={
          <div className="flex items-center space-x-2">
            <SettingOutlined className="text-orange-500" />
            <span>修改密码</span>
          </div>
        }
        open={passwordModalVisible}
        onOk={handlePasswordModalOk}
        onCancel={() => setPasswordModalVisible(false)}
        width={500}
        okText="确定"
        cancelText="取消"
      >
        <Form
          form={passwordForm}
          layout="vertical"
          className="mt-4"
        >
          <div className="mb-4 p-3 bg-blue-50 rounded-lg">
            <div className="flex items-center space-x-2">
              <UserOutlined className="text-blue-500" />
              <span className="font-medium text-blue-700">
                用户: {editingPasswordUser?.username}
              </span>
            </div>
          </div>

          <Form.Item
            name="old_password"
            label="原密码"
            rules={[
              { required: true, message: '请输入原密码' },
            ]}
          >
            <Input.Password 
              placeholder="请输入原密码" 
              className="input-primary"
            />
          </Form.Item>
          
          <Form.Item
            name="new_password"
            label="新密码"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 6, message: '密码至少6个字符' },
              { max: 20, message: '密码最多20个字符' },
            ]}
          >
            <Input.Password 
              placeholder="请输入新密码" 
              className="input-primary"
            />
          </Form.Item>

          <Form.Item
            name="confirm_password"
            label="确认新密码"
            dependencies={['new_password']}
            rules={[
              { required: true, message: '请确认新密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('new_password') === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error('两次输入的密码不一致'));
                },
              }),
            ]}
          >
            <Input.Password 
              placeholder="请再次输入新密码" 
              className="input-primary"
            />
          </Form.Item>
        </Form>
      </Modal>

      {/* 新增/编辑系统提示词模态框 */}
      <Modal
        title={
          <div className="flex items-center space-x-2">
            <RobotOutlined className="text-primary-500" />
            <span>{editingSystemPrompt ? '编辑系统提示词' : '新增系统提示词'}</span>
          </div>
        }
        open={systemPromptModalVisible}
        onOk={handleSystemPromptModalOk}
        onCancel={() => setSystemPromptModalVisible(false)}
        width={600}
        okText="确定"
        cancelText="取消"
      >
        <Form
          form={systemPromptForm}
          layout="vertical"
          className="mt-4"
        >
          <Form.Item
            name="name"
            label="提示词名称"
            rules={[
              { required: true, message: '请输入提示词名称' },
              { min: 2, message: '名称至少2个字符' },
              { max: 50, message: '名称最多50个字符' },
            ]}
          >
            <Input 
              placeholder="请输入提示词名称" 
              className="input-primary"
              prefix={<RobotOutlined className="text-primary-400" />}
            />
          </Form.Item>


          <Form.Item
            name="is_active"
            label="状态"
            initialValue={false}
            valuePropName="checked"
          >
            <Switch 
              checkedChildren="启用" 
              unCheckedChildren="停用"
            />
          </Form.Item>

          <Form.Item
            name="content"
            label="提示词内容"
            rules={[
              { required: true, message: '请输入提示词内容' },
              { max: 2000, message: '内容最多2000个字符' },
            ]}
          >
            <Input.TextArea 
              placeholder="请输入提示词内容，描述AI的角色和行为指导..."
              className="input-primary"
              rows={6}
              showCount
              maxLength={2000}
            />
          </Form.Item>

          <Form.Item
            name="tags"
            label="标签"
            help="用逗号分隔多个标签，便于分类管理"
          >
            <Input 
              placeholder="例如：数据分析,业务顾问,客服助手" 
              className="input-primary"
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default SettingsPage;
