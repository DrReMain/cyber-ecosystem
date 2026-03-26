/**
 * Connect-Web 客户端统一入口
 */

import type { ClientConfig } from '../types';
import { createBlogServiceClient, type BlogServiceClient } from './template1/client';
import { createReadingServiceClient, type ReadingServiceClient } from './template2/client';

// 导出类型（避免命名冲突，使用显式导出）
export type { BlogServiceClient } from './template1/client';
export type { ReadingServiceClient } from './template2/client';

// 导出创建函数
export { createBlogServiceClient } from './template1/client';
export { createReadingServiceClient } from './template2/client';

/**
 * Connect-Web 客户端集合
 */
export interface ConnectWebClients {
  template1: {
    blog: BlogServiceClient;
  };
  template2: {
    reading: ReadingServiceClient;
  };
}

/**
 * 创建 Connect-Web 客户端集合
 */
export function createConnectWebClients(config: ClientConfig): ConnectWebClients {
  return {
    template1: {
      blog: createBlogServiceClient(config),
    },
    template2: {
      reading: createReadingServiceClient(config),
    },
  };
}
