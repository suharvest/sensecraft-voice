// 颜色工具函数
// 确保颜色值始终是字符串格式

export const normalizeColorValue = (color: any): string => {
  if (typeof color === 'string') {
    return color;
  }
  
  if (color && typeof color === 'object') {
    // 处理ColorPicker返回的对象
    if (typeof color.toHexString === 'function') {
      return color.toHexString();
    }
    
    if (color.hex) {
      return color.hex;
    }
    
    if (color.value) {
      return color.value;
    }
  }
  
  // 默认返回红色
  return '#ff4d4f';
};

// 验证颜色格式
export const isValidColor = (color: string): boolean => {
  const hexColorRegex = /^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$/;
  return hexColorRegex.test(color);
};

// 获取颜色显示名称
export const getColorName = (color: string): string => {
  const colorMap: { [key: string]: string } = {
    '#ff4d4f': '红色',
    '#52c41a': '绿色', 
    '#1890ff': '蓝色',
    '#fa8c16': '橙色',
    '#722ed1': '紫色',
    '#13c2c2': '青色',
    '#eb2f96': '粉色',
    '#faad14': '黄色',
  };
  
  return colorMap[color] || '自定义颜色';
};
