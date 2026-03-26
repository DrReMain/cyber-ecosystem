/**
 * demo1 API 客户端配置
 * 使用 connect-web 协议对接 template1 和 template2
 */

import { createBlogServiceClient } from 'shared-sdk/connect-web';
import { createReadingServiceClient } from 'shared-sdk/connect-web';

// API 基础 URL，从环境变量获取
const TEMPLATE1_URL = import.meta.env.VITE_TEMPLATE1_URL || 'http://localhost:13000';
const TEMPLATE2_URL = import.meta.env.VITE_TEMPLATE2_URL || 'http://localhost:13001';

// 创建 connect-web 客户端
export const api = {
  template1: {
    blog: createBlogServiceClient({
      baseUrl: TEMPLATE1_URL,
      headers: {
        'Accept-Language': 'zh-Hans',
      },
    }),
  },
  template2: {
    reading: createReadingServiceClient({
      baseUrl: TEMPLATE2_URL,
      headers: {
        'Accept-Language': 'zh-Hans',
      },
    }),
  },
};

// 导出类型
export type { ApiError } from 'shared-sdk/errors';
