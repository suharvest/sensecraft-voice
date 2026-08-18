// API集成测试工具
// 在浏览器控制台中运行: window.testKeywordsAPI()

import { keywordService } from '@/services/keywords';

export const testKeywordsAPI = async () => {
  console.log('🧪 开始测试关键词API集成...');
  console.log('📋 API文档版本: v1.3.0');
  console.log('🔗 基础路径: /api/v1/keywords');
  
  const results = {
    getKeywords: false,
    createKeyword: false,
    updateKeyword: false,
    deleteKeyword: false,
    batchDelete: false
  };

  try {
    // 测试1: 获取关键词列表
    console.log('\n📋 测试1: 获取关键词列表...');
    try {
      const listResult = await keywordService.getKeywords({ offset: 0, limit: 5 });
      console.log('✅ 获取关键词列表成功:', listResult);
      results.getKeywords = true;
    } catch (error: any) {
      console.log('❌ 获取关键词列表失败:', error.message);
      if (error.message.includes('404') || error.message.includes('not found')) {
        console.log('💡 这是预期的，因为后端API可能还未实现');
      }
    }

    // 测试2: 创建关键词
    console.log('\n➕ 测试2: 创建关键词...');
    try {
      const createResult = await keywordService.createKeyword({
        keyword: 'API集成测试',
        synonyms: '测试, 验证, 检查, 集成',
        mark_color: '#52c41a'
      });
      console.log('✅ 创建关键词成功:', createResult);
      results.createKeyword = true;
    } catch (error: any) {
      console.log('❌ 创建关键词失败:', error.message);
      if (error.message.includes('404') || error.message.includes('not found')) {
        console.log('💡 这是预期的，因为后端API可能还未实现');
      }
    }

    // 测试3: 更新关键词
    console.log('\n✏️ 测试3: 更新关键词...');
    try {
      const updateResult = await keywordService.updateKeyword(1, {
        keyword: '更新后的关键词',
        synonyms: '更新, 修改, 编辑',
        mark_color: '#1890ff'
      });
      console.log('✅ 更新关键词成功:', updateResult);
      results.updateKeyword = true;
    } catch (error: any) {
      console.log('❌ 更新关键词失败:', error.message);
      if (error.message.includes('404') || error.message.includes('not found')) {
        console.log('💡 这是预期的，因为后端API可能还未实现');
      }
    }

    // 测试4: 删除关键词
    console.log('\n🗑️ 测试4: 删除关键词...');
    try {
      await keywordService.deleteKeyword(999);
      console.log('✅ 删除关键词成功');
      results.deleteKeyword = true;
    } catch (error: any) {
      console.log('❌ 删除关键词失败:', error.message);
      if (error.message.includes('404') || error.message.includes('not found')) {
        console.log('💡 这是预期的，因为后端API可能还未实现');
      }
    }

    // 测试5: 批量删除关键词
    console.log('\n🗑️ 测试5: 批量删除关键词...');
    try {
      const batchResult = await keywordService.batchDeleteKeywords([999, 998, 997]);
      console.log('✅ 批量删除关键词成功:', batchResult);
      results.batchDelete = true;
    } catch (error: any) {
      console.log('❌ 批量删除关键词失败:', error.message);
      if (error.message.includes('404') || error.message.includes('not found')) {
        console.log('💡 这是预期的，因为后端API可能还未实现');
      }
    }

  } catch (error) {
    console.error('❌ 测试过程中发生错误:', error);
  }

  // 输出测试结果摘要
  console.log('\n📊 测试结果摘要:');
  console.log('==================');
  Object.entries(results).forEach(([test, passed]) => {
    console.log(`${passed ? '✅' : '❌'} ${test}: ${passed ? '通过' : '失败'}`);
  });

  const passedCount = Object.values(results).filter(Boolean).length;
  const totalCount = Object.keys(results).length;
  
  console.log(`\n🎯 总体结果: ${passedCount}/${totalCount} 个测试通过`);
  
  if (passedCount === 0) {
    console.log('💡 所有测试都失败，这通常意味着后端API还未实现');
    console.log('📖 请参考 docs/keywords_api.md 文档实现后端API');
  } else if (passedCount === totalCount) {
    console.log('🎉 所有API测试都通过！集成成功！');
  } else {
    console.log('⚠️ 部分API测试通过，请检查失败的接口');
  }

  return results;
};

// 将测试函数挂载到window对象，方便在控制台调用
if (typeof window !== 'undefined') {
  (window as any).testKeywordsAPI = testKeywordsAPI;
  console.log('🔧 API测试工具已加载，在控制台运行: testKeywordsAPI()');
}
