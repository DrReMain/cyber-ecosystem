/**
 * @cyber/sdk 统一入口
 *
 * 提供两种客户端实现：
 * - connect-web: 使用 Connect 协议，适合现代浏览器和 Node.js
 * - ts-rest: 使用 HTTP REST 协议，适合小程序、鸿蒙等环境
 */

// 类型导出
export * from './types';

// 错误处理
export * from './errors';

// connect-web 客户端
export { createConnectWebClients, type ConnectWebClients } from './connect-web';
export type { BlogServiceClient as ConnectWebBlogServiceClient } from './connect-web/template1';
export type { ReadingServiceClient as ConnectWebReadingServiceClient } from './connect-web/template2';

// ts-rest 客户端
export { createTsRestClients, type TsRestClients } from './ts-rest';
export { blogContract, readingContract } from './ts-rest';
export type { BlogServiceClient as TsRestBlogServiceClient, ReadingServiceClient as TsRestReadingServiceClient } from './ts-rest/types';
