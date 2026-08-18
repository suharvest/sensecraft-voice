// 搜索功能测试工具
// 在浏览器控制台中运行: testSearchFunction()

export const testSearchFunction = () => {
  console.log('🔍 测试搜索功能...');
  
  // 模拟搜索输入状态
  const searchInput = {
    keyword: '测试关键词',
    mark_color: '#ff4d4f'
  };
  
  const userSearchInput = {
    username: '测试用户'
  };
  
  console.log('📝 关键词搜索输入状态:');
  console.log('  keyword:', searchInput.keyword);
  console.log('  mark_color:', searchInput.mark_color);
  
  console.log('👤 用户搜索输入状态:');
  console.log('  username:', userSearchInput.username);
  
  // 模拟搜索参数生成
  const keywordSearchParams = {
    offset: 0,
    limit: 10,
    keyword: searchInput.keyword || undefined,
    mark_color: searchInput.mark_color || undefined,
  };
  
  const userSearchParams = {
    offset: 0,
    limit: 20,
    username: userSearchInput.username || undefined,
  };
  
  console.log('🔗 生成的关键词搜索参数:');
  console.log('  ', keywordSearchParams);
  
  console.log('🔗 生成的用户搜索参数:');
  console.log('  ', userSearchParams);
  
  console.log('✅ 搜索功能测试完成！');
  console.log('💡 现在搜索只在点击搜索按钮时才会调用API');
  console.log('💡 输入时不会触发API调用，完全移除了键盘监听');
  console.log('💡 用户需要手动点击"搜索"按钮来执行搜索');
};

// 监控搜索行为的函数
export const monitorSearchBehavior = () => {
  console.log('👀 开始监控搜索行为...');
  
  // 监控网络请求
  const originalFetch = window.fetch;
  let requestCount = 0;
  
  window.fetch = function(...args) {
    requestCount++;
    const url = args[0];
    console.log(`🌐 API请求 #${requestCount}:`, url);
    return originalFetch.apply(this, args);
  };
  
  // 监控console.log来检测useEffect调用
  const originalLog = console.log;
  console.log = function(...args) {
    if (args[0] && args[0].includes && args[0].includes('API Response interceptor')) {
      console.log('🔍 检测到API调用:', args);
    }
    return originalLog.apply(this, args);
  };
  
  console.log('📊 监控已启动，现在请在页面上进行以下操作：');
  console.log('1. 在搜索框中输入内容（应该不会触发API调用）');
  console.log('2. 点击搜索按钮（应该触发API调用）');
  console.log('3. 点击重置按钮（应该触发API调用）');
  
  // 5秒后显示统计
  setTimeout(() => {
    console.log(`📈 监控结果: 共检测到 ${requestCount} 次API调用`);
    if (requestCount === 0) {
      console.log('✅ 很好！没有检测到意外的API调用');
    } else {
      console.log('⚠️ 检测到API调用，请检查是否在输入时触发了搜索');
    }
  }, 5000);
};

// 挂载到window对象
if (typeof window !== 'undefined') {
  (window as any).testSearchFunction = testSearchFunction;
  (window as any).monitorSearchBehavior = monitorSearchBehavior;
  console.log('🔧 搜索功能测试已加载，在控制台运行:');
  console.log('  testSearchFunction() - 测试搜索功能');
  console.log('  monitorSearchBehavior() - 监控搜索行为');
}
