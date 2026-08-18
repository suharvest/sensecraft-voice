// 格式化时间戳为可读日期
export const formatDate = (timestamp: number): string => {
  if (!timestamp) return '-';
  
  const date = new Date(timestamp);
  const now = new Date();
  const diffInHours = (now.getTime() - date.getTime()) / (1000 * 60 * 60);
  
  // 如果是今天，显示时间
  if (diffInHours < 24 && date.toDateString() === now.toDateString()) {
    return date.toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  }
  
  // 如果是昨天，显示"昨天 + 时间"
  const yesterday = new Date(now);
  yesterday.setDate(yesterday.getDate() - 1);
  if (date.toDateString() === yesterday.toDateString()) {
    return `昨天 ${date.toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit',
    })}`;
  }
  
  // 其他情况显示完整日期
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
};

// 格式化时间戳为相对时间
export const formatRelativeTime = (timestamp: number): string => {
  if (!timestamp) return '-';
  
  const now = new Date().getTime();
  const diff = now - timestamp;
  
  const seconds = Math.floor(diff / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);
  
  if (days > 0) {
    return `${days}天前`;
  } else if (hours > 0) {
    return `${hours}小时前`;
  } else if (minutes > 0) {
    return `${minutes}分钟前`;
  } else {
    return '刚刚';
  }
};

// 智能时间格式化 - 根据时间差显示最合适的格式（仪表盘专用，不显示"刚刚"）
export const formatSmartTime = (timestamp: number): string => {
  if (!timestamp) return '-';
  
  // 检查时间戳格式：如果是秒级时间戳（10位数），转换为毫秒
  const actualTimestamp = timestamp.toString().length === 10 ? timestamp * 1000 : timestamp;
  
  const date = new Date(actualTimestamp);
  const now = new Date();
  const diff = now.getTime() - actualTimestamp;
  
  // 调试信息
  console.log('formatSmartTime debug:', {
    originalTimestamp: timestamp,
    actualTimestamp,
    date: date.toISOString(),
    now: now.toISOString(),
    diff,
    diffMinutes: Math.floor(diff / 1000 / 60)
  });
  
  const seconds = Math.floor(diff / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);
  
  // 1小时内显示分钟（包括1分钟内的情况）
  if (hours < 1) {
    // 如果少于1分钟，显示具体时间而不是"刚刚"
    if (minutes < 1) {
      return date.toLocaleTimeString('zh-CN', {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      });
    }
    return `${minutes}分钟前`;
  }
  
  // 1天内显示具体时间
  if (days < 1 && date.toDateString() === now.toDateString()) {
    return date.toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit',
    });
  }
  
  // 昨天显示"昨天 + 时间"
  const yesterday = new Date(now);
  yesterday.setDate(yesterday.getDate() - 1);
  if (date.toDateString() === yesterday.toDateString()) {
    return `昨天 ${date.toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit',
    })}`;
  }
  
  // 一周内显示星期
  if (days < 7) {
    const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];
    return `${weekdays[date.getDay()]} ${date.toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit',
    })}`;
  }
  
  // 超过一周显示日期
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
};

// 格式化时间戳为ISO字符串
export const formatISO = (timestamp: number): string => {
  if (!timestamp) return '-';
  return new Date(timestamp).toISOString();
};

// 获取时间范围
export const getTimeRange = (days: number) => {
  const end = new Date();
  const start = new Date();
  start.setDate(start.getDate() - days);
  
  return {
    start: start.getTime(),
    end: end.getTime(),
  };
};

// 判断设备是否在线
// 如果updated_at与当前时间比较超过了5分钟，则判断为离线
export const isDeviceOnline = (updatedAt: number): boolean => {
  if (!updatedAt) return false;
  
  const now = new Date().getTime();
  const diffInMinutes = (now - updatedAt) / (1000 * 60);
  
  // 超过5分钟判断为离线
  return diffInMinutes <= 5;
};

// 获取设备在线状态文本
export const getDeviceStatusText = (updatedAt: number): string => {
  return isDeviceOnline(updatedAt) ? '在线' : '离线';
};

// 获取设备在线状态颜色
export const getDeviceStatusColor = (updatedAt: number): string => {
  return isDeviceOnline(updatedAt) ? 'success' : 'error';
};

// 格式化完整时间 - 专门用于录音记录显示完整的 yyyy/mm/dd HH:mm:ss 格式
export const formatFullDateTime = (timestamp: number): string => {
  if (!timestamp) return '-';
  
  // 检查时间戳格式：如果是秒级时间戳（10位数），转换为毫秒
  const actualTimestamp = timestamp.toString().length === 10 ? timestamp * 1000 : timestamp;
  
  const date = new Date(actualTimestamp);
  
  // 格式化为 yyyy/mm/dd HH:mm:ss
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  const seconds = String(date.getSeconds()).padStart(2, '0');
  
  return `${year}/${month}/${day} ${hours}:${minutes}:${seconds}`;
};

// 格式化批量分析专用的压缩时间 - 今天显示HH:mm:ss，其他显示MM/dd HH:mm:ss
export const formatCompactDateTime = (timestamp: number): string => {
  if (!timestamp) return '-';
  
  // 检查时间戳格式：如果是秒级时间戳（10位数），转换为毫秒
  const actualTimestamp = timestamp.toString().length === 10 ? timestamp * 1000 : timestamp;
  
  const date = new Date(actualTimestamp);
  const now = new Date();
  
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  const seconds = String(date.getSeconds()).padStart(2, '0');
  
  // 如果是今天，只显示时间
  if (date.toDateString() === now.toDateString()) {
    return `${hours}:${minutes}:${seconds}`;
  }
  
  // 其他情况显示MM/dd HH:mm:ss
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  
  return `${month}/${day} ${hours}:${minutes}:${seconds}`;
};
