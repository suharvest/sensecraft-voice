// 环境配置管理
export interface EnvConfig {
  API_BASE_URL: string;
  APP_TITLE: string;
  APP_ENV: string;
  TIMEOUT: number;
}

// 获取环境变量
const getEnvVar = (key: string, defaultValue: string = ''): string => {
  return import.meta.env[key] || defaultValue;
};

// 环境配置
export const envConfig: EnvConfig = {
  API_BASE_URL: getEnvVar('VITE_API_BASE_URL', 'http://localhost:8080'),
  APP_TITLE: getEnvVar('VITE_APP_TITLE', 'SenseCraft Voice'),
  APP_ENV: getEnvVar('VITE_APP_ENV', 'development'),
  TIMEOUT: parseInt(getEnvVar('VITE_API_TIMEOUT', '10000')),
};

// 判断是否为开发环境
export const isDev = envConfig.APP_ENV === 'development';

// 判断是否为生产环境
export const isProd = envConfig.APP_ENV === 'production';

// 判断是否为测试环境
export const isTest = envConfig.APP_ENV === 'test';

// 导出配置
export default envConfig;

