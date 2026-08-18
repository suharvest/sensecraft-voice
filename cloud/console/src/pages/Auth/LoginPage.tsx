import React, { useState } from 'react';
import { Form, Input, Button, Card, message, Checkbox } from 'antd';
import { UserOutlined, LockOutlined, SoundOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { userService } from '@/services/users';
import { useAuthStore } from '@/stores/authStore';

interface LoginForm {
  username: string;
  password: string;
  remember: boolean;
}

const LoginPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { login } = useAuthStore();

  const onFinish = async (values: LoginForm) => {
    try {
      setLoading(true);
      const response = await userService.login({
        username: values.username,
        password: values.password,
      });
      
      login(response.user, response.token);
      message.success('登录成功！');
      navigate('/');
    } catch (error) {
      message.error('登录失败，请检查用户名和密码');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-50 via-white to-primary-100 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        {/* Logo和标题 */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-20 h-20 bg-primary-500 rounded-full mb-4 shadow-lg">
            <SoundOutlined className="text-white text-4xl" />
          </div>
          <h1 className="text-3xl font-bold text-primary-700 mb-2">
            SenseCraft Voice
          </h1>
          <p className="text-lg text-primary-600">
            语音管理系统
          </p>
        </div>

        {/* 登录表单 */}
        <Card className="card-primary shadow-xl">
          <div className="p-6">
            <h2 className="text-2xl font-semibold text-center text-primary-700 mb-6">
              用户登录
            </h2>
            
            <Form
              name="login"
              initialValues={{ remember: true }}
              onFinish={onFinish}
              size="large"
              layout="vertical"
            >
              <Form.Item
                name="username"
                label="用户名"
                rules={[
                  { required: true, message: '请输入用户名' },
                  { min: 3, message: '用户名至少3个字符' },
                  { max: 20, message: '用户名最多20个字符' },
                ]}
              >
                <Input
                  prefix={<UserOutlined className="text-primary-400" />}
                  placeholder="请输入用户名"
                  className="input-primary h-12"
                />
              </Form.Item>

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
                  prefix={<LockOutlined className="text-primary-400" />}
                  placeholder="请输入密码"
                  className="input-primary h-12"
                />
              </Form.Item>

              <Form.Item>
                <div className="flex items-center justify-between">
                  <Form.Item name="remember" valuePropName="checked" noStyle>
                    <Checkbox className="text-primary-600">
                      记住我
                    </Checkbox>
                  </Form.Item>
                  <a 
                    href="#" 
                    className="text-primary-600 hover:text-primary-700 text-sm"
                  >
                    忘记密码？
                  </a>
                </div>
              </Form.Item>

              <Form.Item>
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={loading}
                  className="w-full h-12 bg-primary-500 hover:bg-primary-600 border-primary-500 hover:border-primary-600 text-lg font-medium"
                >
                  {loading ? '登录中...' : '登录'}
                </Button>
              </Form.Item>
            </Form>


          </div>
        </Card>

        {/* 页脚 */}
        <div className="text-center mt-6">
          <p className="text-sm text-primary-500">
            © 2025 SenseCraft Voice. All rights reserved.
          </p>
        </div>
      </div>
    </div>
  );
};

export default LoginPage;
