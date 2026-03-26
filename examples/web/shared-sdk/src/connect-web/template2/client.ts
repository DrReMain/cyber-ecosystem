/**
 * Template2 ReadingService Connect-Web 客户端
 */

import type { ClientConfig } from '../../types';
import { ApiError } from '../../errors';
import type {
  QueryBlogRequest,
  QueryBlogResponse,
  GetBlogRequest,
  GetBlogResponse,
  RecordReadingRequest,
  RecordReadingResponse,
} from './types';

/**
 * ReadingService 客户端接口
 */
export interface ReadingServiceClient {
  queryBlog(request: QueryBlogRequest): Promise<QueryBlogResponse>;
  getBlog(request: GetBlogRequest): Promise<GetBlogResponse>;
  recordReading(request: RecordReadingRequest): Promise<RecordReadingResponse>;
}

/**
 * 创建 ReadingService 客户端
 */
export function createReadingServiceClient(config: ClientConfig): ReadingServiceClient {
  return new ReadingServiceClientImpl(config);
}

/**
 * ReadingService 客户端实现
 * 使用 fetch 直接调用 Connect 协议
 */
class ReadingServiceClientImpl implements ReadingServiceClient {
  private readonly baseUrl: string;
  private readonly headers: Record<string, string>;

  constructor(config: ClientConfig) {
    this.baseUrl = config.baseUrl.replace(/\/$/, '');
    this.headers = config.headers ?? {};
  }

  private async call<T>(path: string, request: unknown): Promise<T> {
    const url = `${this.baseUrl}${path}`;

    try {
      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
          ...this.headers,
        },
        body: JSON.stringify(request),
      });

      if (!response.ok) {
        const errorBody = await response.json().catch(() => ({}));
        throw ApiError.fromHttpResponse(response, errorBody);
      }

      return (await response.json()) as T;
    } catch (error) {
      if (error instanceof ApiError) {
        throw error;
      }
      throw ApiError.fromConnectError(error);
    }
  }

  async queryBlog(request: QueryBlogRequest): Promise<QueryBlogResponse> {
    return this.call<QueryBlogResponse>('/api.template2.v1.ReadingService/QueryBlog', request);
  }

  async getBlog(request: GetBlogRequest): Promise<GetBlogResponse> {
    return this.call<GetBlogResponse>('/api.template2.v1.ReadingService/GetBlog', request);
  }

  async recordReading(request: RecordReadingRequest): Promise<RecordReadingResponse> {
    return this.call<RecordReadingResponse>('/api.template2.v1.ReadingService/RecordReading', request);
  }
}
