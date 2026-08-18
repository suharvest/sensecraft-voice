import React, { useState } from 'react';
import { Layout, Menu, Button, Avatar, Dropdown, Space } from 'antd';
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  HomeOutlined,
  SoundOutlined,
  ShopOutlined,
  UserOutlined,
  LogoutOutlined,
  SettingOutlined,
  RobotOutlined,
} from '@ant-design/icons';
import { useNavigate, useLocation, Outlet } from 'react-router-dom';
import { useAuthStore } from '@/stores/authStore';
import { getAppVersion } from '@/utils/version';

const { Header, Sider, Content } = Layout;

const AppLayout: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuthStore();

  const menuItems = [
    {
      key: '/',
      icon: <HomeOutlined />,
      label: '仪表盘',
    },
    {
      key: '/ai-analysis',
      icon: <RobotOutlined />,
      label: 'AI分析',
    },
    {
      key: '/stores',
      icon: <ShopOutlined />,
      label: '门店管理',
    },
    {
      key: '/settings',
      icon: <SettingOutlined />,
      label: '后台配置',
    },
  ];

  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人资料',
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '系统设置',
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: () => {
        logout();
        navigate('/login');
      },
    },
  ];

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key);
  };

  const selectedKey = menuItems.find(item => 
    location.pathname === item.key || location.pathname.startsWith(item.key + '/')
  )?.key || '/';

  return (
    <Layout className="h-screen">
      <Sider 
        trigger={null} 
        collapsible 
        collapsed={collapsed}
        className="bg-white border-r border-primary-200 shadow-lg relative"
        width={256}
      >
        {/* Logo区域 */}
        <div className="p-6 border-b border-primary-100">
          <div className="flex items-center space-x-3">
            <div className="w-10 h-10 bg-primary-500 rounded-lg flex items-center justify-center">
              <SoundOutlined className="text-white text-xl" />
            </div>
            {!collapsed && (
              <h1 className="text-xl font-bold text-primary-700">
                SenseCraft Voice
              </h1>
            )}
          </div>
        </div>

        {/* 导航菜单 */}
        <Menu
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
          onClick={handleMenuClick}
          className="border-0 bg-transparent"
          style={{ 
            backgroundColor: 'transparent',
            border: 'none',
          }}
        />
        
        {/* 版本号角标 */}
        <div className="absolute bottom-2 right-2">
          <div className="text-xs text-gray-400 bg-gray-100 px-2 py-1 rounded-full">
            v{getAppVersion()}
          </div>
        </div>
      </Sider>

      <Layout>
        <Header className="bg-white border-b border-primary-200 px-6 flex items-center justify-between shadow-sm">
          <div className="flex items-center">
            <Button
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setCollapsed(!collapsed)}
              className="text-primary-600 hover:text-primary-700 hover:bg-primary-50"
            />
            <div className="ml-4">
              <h2 className="text-lg font-semibold text-primary-700">
                {menuItems.find(item => item.key === selectedKey)?.label || '仪表盘'}
              </h2>
            </div>
          </div>

          <div className="flex items-center space-x-4">
            <Dropdown
              menu={{ items: userMenuItems }}
              placement="bottomRight"
              arrow
            >
              <Space className="cursor-pointer hover:bg-primary-50 px-3 py-2 rounded-lg transition-colors">
                <Avatar 
                  icon={<UserOutlined />} 
                  className="bg-primary-500"
                />
                <span className="text-primary-700 font-medium">
                  {user?.username || '用户'}
                </span>
              </Space>
            </Dropdown>
          </div>
        </Header>

        <Content className="bg-gradient-to-br from-primary-50 to-white h-full overflow-y-auto">
          <div className="p-6">
            <Outlet />
          </div>
        </Content>
      </Layout>
    </Layout>
  );
};

export default AppLayout;
