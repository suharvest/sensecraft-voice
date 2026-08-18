import CryptoJS from 'crypto-js';

/**
 * 对字符串进行MD5加密
 * @param text 要加密的文本
 * @returns MD5加密后的字符串
 */
export const md5 = (text: string): string => {
  return CryptoJS.MD5(text).toString();
};

/**
 * 对密码进行MD5加密
 * @param password 原始密码
 * @returns MD5加密后的密码
 */
export const encryptPassword = (password: string): string => {
  return md5(password);
};
