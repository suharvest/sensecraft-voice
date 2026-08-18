import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import AppLayout from './components/Layout/AppLayout';

import DashboardPage from './pages/Dashboard/DashboardPage';
import LoginPage from './pages/Auth/LoginPage';
import StoresPage from './pages/Stores/StoresPage';
import SettingsPage from './pages/Settings/SettingsPage';
import AIAnalysisPage from './pages/AIAnalysis/AIAnalysisPage';
import { useAuthStore } from './stores/authStore';

// 受保护的路由组件
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAuthStore();
  
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  
  return <>{children}</>;
};

const App: React.FC = () => {
  return (
    <Routes>
      {/* 登录页面 */}
      <Route path="/login" element={<LoginPage />} />
      
      {/* 受保护的路由 */}
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <AppLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<DashboardPage />} />
        <Route path="ai-analysis" element={<AIAnalysisPage />} />
        <Route path="stores" element={<StoresPage />} />
        <Route path="settings" element={<SettingsPage />} />
      </Route>
      
      {/* 默认重定向到首页 */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
};

export default App;
