// 获取应用版本号
export const getAppVersion = (): string => {
  // 从 package.json 中读取版本号
  // 在构建时会被 Vite 处理
  return import.meta.env.VITE_APP_VERSION || '1.0.0';
};

// 获取应用名称
export const getAppName = (): string => {
  return import.meta.env.VITE_APP_NAME || 'SenseCraft Voice';
};
