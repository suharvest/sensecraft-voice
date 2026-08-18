import React, { useState, useEffect } from 'react';
import {
  Button,
  Card,
  Space,
  Tag,
  Tooltip,
  Dropdown,
  Menu,
  Typography,
  message
} from 'antd';
import {
  SettingOutlined,
  RobotOutlined,
  DownOutlined,
  CheckOutlined,
  ReloadOutlined,
  CloseOutlined
} from '@ant-design/icons';
import { systemPromptService } from '@/services/systemPrompts';
import { SystemPrompt } from '@/types';
import './SystemPromptManager.css';

const { Text, Paragraph } = Typography;

interface SystemPromptSelectorProps {
  selectedPrompt: SystemPrompt | null;
  onPromptChange: (prompt: SystemPrompt | null) => void;
  onManageClick: () => void;
  className?: string;
}

const SystemPromptSelector: React.FC<SystemPromptSelectorProps> = ({
  selectedPrompt,
  onPromptChange,
  onManageClick,
  className = ''
}) => {
  const [prompts, setPrompts] = useState<SystemPrompt[]>([]);
  const [loading, setLoading] = useState(false);

  // 获取激活的系统提示词列表
  const fetchActivePrompts = async () => {
    try {
      setLoading(true);
      const activePrompts = await systemPromptService.getActiveSystemPrompts();
      setPrompts(activePrompts);
      
      // 如果没有激活的提示词，显示提示信息
      if (activePrompts.length === 0) {
        console.log('当前没有激活的系统提示词');
      }
    } catch (error) {
      console.error('获取系统提示词失败:', error);
      message.error('获取系统提示词失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchActivePrompts();
  }, []);

  // 处理选择提示词
  const handleSelectPrompt = (prompt: SystemPrompt) => {
    onPromptChange(prompt);
  };

  // 处理清除选择
  const handleClearSelection = () => {
    onPromptChange(null);
  };

  // 处理刷新
  const handleRefresh = () => {
    fetchActivePrompts();
  };

  // 构建下拉菜单
  const menuItems = prompts.map(prompt => ({
    key: prompt.id?.toString() || `prompt-${Math.random()}`,
    label: (
      <div className="p-2 min-w-64">
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center space-x-2">
            <RobotOutlined className="text-blue-500" />
            <Text strong>{prompt.name}</Text>
            {prompt.id === selectedPrompt?.id && (
              <CheckOutlined className="text-green-500" />
            )}
          </div>
          <Tag color="green">v{prompt.version}</Tag>
        </div>
        <Paragraph 
          ellipsis={{ rows: 2 }} 
          style={{ margin: 0, fontSize: '12px', color: '#666' }}
        >
          {prompt.content}
        </Paragraph>
        {prompt.tags && (
          <div className="mt-2">
            {(() => {
              try {
                const tags = JSON.parse(prompt.tags);
                return Array.isArray(tags) ? tags.map((tag: string, index: number) => (
                  <Tag key={`${tag}-${index}`} color="blue">
                    {tag}
                  </Tag>
                )) : null;
              } catch {
                return null;
              }
            })()}
          </div>
        )}
      </div>
    ),
    onClick: () => handleSelectPrompt(prompt)
  }));

  const menu = (
    <Menu items={menuItems} className="max-h-96 overflow-y-auto" />
  );

  return (
    <Card 
      size="small" 
      className={`system-prompt-selector ${className}`}
      styles={{ body: { padding: '8px 12px' } }}
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-3">
          <div className="flex items-center space-x-2">
            <RobotOutlined className="text-primary-500" />
            <Text strong>AI角色:</Text>
          </div>
          
          {selectedPrompt ? (
            <div className="system-prompt-role-info">
              <Text strong className="text-primary-600">
                {selectedPrompt.name}
              </Text>
              <Tag color="green">
                v{selectedPrompt.version}
              </Tag>
            </div>
          ) : (
            <Text type="secondary">未选择角色</Text>
          )}
        </div>

        <Space>
          <Tooltip title="刷新提示词列表">
            <Button
              type="text"
              size="small"
              icon={<ReloadOutlined />}
              onClick={handleRefresh}
              loading={loading}
              className="text-gray-400 hover:text-primary-500"
            />
          </Tooltip>
          
          {selectedPrompt && (
            <Tooltip title="清除选择">
              <Button
                type="text"
                size="small"
                icon={<CloseOutlined />}
                onClick={handleClearSelection}
                className="text-gray-400 hover:text-red-500"
              />
            </Tooltip>
          )}
          
          {prompts.length > 0 && (
            <Dropdown 
              overlay={menu} 
              trigger={['click']}
              placement="bottomRight"
              disabled={loading}
            >
              <Button 
                type="primary" 
                size="small"
                icon={<DownOutlined />}
                loading={loading}
              >
                选择角色
              </Button>
            </Dropdown>
          )}
          
          <Button
            type={prompts.length > 0 ? "default" : "primary"}
            size="small"
            icon={<SettingOutlined />}
            onClick={onManageClick}
          >
            {prompts.length > 0 ? "管理" : "创建角色"}
          </Button>
        </Space>
      </div>

      {/* 当前选择的提示词详情 */}
      {selectedPrompt && (
        <div className="system-prompt-details">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <Text type="secondary">角色类型:</Text>
              <Tag color="blue">{selectedPrompt.role}</Tag>
            </div>
            {selectedPrompt.tags && (
              <div className="flex items-center space-x-1">
                <Text type="secondary">标签:</Text>
                <div className="flex space-x-1">
                  {(() => {
                    try {
                      const tags = JSON.parse(selectedPrompt.tags);
                      if (!Array.isArray(tags)) return null;
                      return (
                        <>
                          {tags.slice(0, 3).map((tag: string, index: number) => (
                            <Tag key={`${tag}-${index}`} color="blue">
                              {tag}
                            </Tag>
                          ))}
                          {tags.length > 3 && (
                            <Tag color="default">
                              +{tags.length - 3}
                            </Tag>
                          )}
                        </>
                      );
                    } catch {
                      return null;
                    }
                  })()}
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* 提示信息 - 简化版本 */}
      {!selectedPrompt && (
        <div className="mt-2 pt-2 border-t border-gray-100">
          <Text type="secondary" className="text-xs">
            {prompts.length > 0 
              ? "💡 选择一个AI角色来获得更专业的分析结果"
              : "💡 点击\"创建角色\"按钮来添加第一个AI角色"
            }
          </Text>
        </div>
      )}
    </Card>
  );
};

export default SystemPromptSelector;
