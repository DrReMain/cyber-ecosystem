/**
 * ts-rest 客户端统一入口
 */

import { initClient } from '@ts-rest/core';
import type { ClientConfig } from '../types';
import { blogContract } from './template1/contract';
import { readingContract } from './template2/contract';
import type { BlogServiceClient, ReadingServiceClient } from './types';

// 导出契约
export { blogContract } from './template1/contract';
export { readingContract } from './template2/contract';

// 导出类型
export type { BlogServiceClient, ReadingServiceClient } from './types';

/**
 * ts-rest 客户端集合
 */
export interface TsRestClients {
  template1: {
    blog: BlogServiceClient;
  };
  template2: {
    reading: ReadingServiceClient;
  };
}

/**
 * 创建 ts-rest 客户端集合
 */
export function createTsRestClients(config: ClientConfig): TsRestClients {
  const baseOptions = {
    baseUrl: config.baseUrl,
    baseHeaders: config.headers ?? {},
  };

  return {
    template1: {
      blog: initClient(blogContract, baseOptions) as BlogServiceClient,
    },
    template2: {
      reading: initClient(readingContract, baseOptions) as ReadingServiceClient,
    },
  };
}
