/**
 * demo2 API 客户端配置
 * 使用 ts-rest 协议对接 template1 和 template2
 */

import { createTsRestClients } from 'shared-sdk/ts-rest';

// API 基础 URL，从环境变量获取
// ts-rest 使用 HTTP 协议，所以使用 HTTP 端口
const TEMPLATE1_URL = import.meta.env.VITE_TEMPLATE1_URL || 'http://localhost:11000';
const TEMPLATE2_URL = import.meta.env.VITE_TEMPLATE2_URL || 'http://localhost:11001';

// 创建 ts-rest 客户端
export const api = {
  template1: createTsRestClients({
    baseUrl: TEMPLATE1_URL,
    headers: {
      'Accept-Language': 'zh-Hans',
    },
  }).template1,
  template2: createTsRestClients({
    baseUrl: TEMPLATE2_URL,
    headers: {
      'Accept-Language': 'zh-Hans',
    },
  }).template2,
};

// 导出类型
export type { ApiError } from 'shared-sdk/errors';
