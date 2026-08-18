// 颜色工具函数测试工具
// 在浏览器控制台中运行: testColorUtils()

import { normalizeColorValue, isValidColor, getColorName } from './colorUtils';

// 手动测试函数
export const testColorUtils = () => {
  console.log('🧪 测试颜色工具函数...');
  
  // 测试字符串输入
  console.log('测试字符串输入:');
  console.log('  #ff4d4f ->', normalizeColorValue('#ff4d4f'));
  console.log('  #52c41a ->', normalizeColorValue('#52c41a'));
  
  // 测试对象输入
  console.log('测试对象输入:');
  const mockColor = { toHexString: () => '#1890ff' };
  console.log('  mockColor ->', normalizeColorValue(mockColor));
  
  // 测试无效输入
  console.log('测试无效输入:');
  console.log('  null ->', normalizeColorValue(null));
  console.log('  undefined ->', normalizeColorValue(undefined));
  
  // 测试颜色验证
  console.log('测试颜色验证:');
  console.log('  #ff4d4f 有效:', isValidColor('#ff4d4f'));
  console.log('  #gggggg 有效:', isValidColor('#gggggg'));
  
  // 测试颜色名称
  console.log('测试颜色名称:');
  console.log('  #ff4d4f ->', getColorName('#ff4d4f'));
  console.log('  #52c41a ->', getColorName('#52c41a'));
  
  console.log('✅ 颜色工具函数测试完成！');
};

// 挂载到window对象
if (typeof window !== 'undefined') {
  (window as any).testColorUtils = testColorUtils;
  console.log('🔧 颜色工具测试已加载，在控制台运行: testColorUtils()');
}
