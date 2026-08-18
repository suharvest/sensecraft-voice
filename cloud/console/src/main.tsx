import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import { antdTheme } from './theme/antd-theme';
import App from './App';
import './index.css';
// 导入API测试工具（开发环境）
import './utils/apiTest';
// 导入颜色工具测试（开发环境）
import './utils/colorTest';
// 导入搜索功能测试（开发环境）
import './utils/searchTest';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <ConfigProvider theme={antdTheme}>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </ConfigProvider>
);
