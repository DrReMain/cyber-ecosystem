/**
 * Template1 BlogService Connect-Web 客户端
 */

import type { ClientConfig } from '../../types';
import { ApiError } from '../../errors';
import type {
  CreateBlogRequest,
  CreateBlogResponse,
  UpdateBlogRequest,
  UpdateBlogResponse,
  DeleteBlogRequest,
  DeleteBlogResponse,
  DeleteBatchBlogRequest,
  DeleteBatchBlogResponse,
  GetBlogRequest,
  GetBlogResponse,
  QueryBlogRequest,
  QueryBlogResponse,
} from './types';

/**
 * BlogService 客户端接口
 */
export interface BlogServiceClient {
  createBlog(request: CreateBlogRequest): Promise<CreateBlogResponse>;
  updateBlog(request: UpdateBlogRequest): Promise<UpdateBlogResponse>;
  deleteBlog(request: DeleteBlogRequest): Promise<DeleteBlogResponse>;
  deleteBatchBlog(request: DeleteBatchBlogRequest): Promise<DeleteBatchBlogResponse>;
  getBlog(request: GetBlogRequest): Promise<GetBlogResponse>;
  queryBlog(request: QueryBlogRequest): Promise<QueryBlogResponse>;
}

/**
 * 创建 BlogService 客户端
 */
export function createBlogServiceClient(config: ClientConfig): BlogServiceClient {
  return new BlogServiceClientImpl(config);
}

/**
 * BlogService 客户端实现
 * 使用 fetch 直接调用 Connect 协议
 */
class BlogServiceClientImpl implements BlogServiceClient {
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

  async createBlog(request: CreateBlogRequest): Promise<CreateBlogResponse> {
    return this.call<CreateBlogResponse>('/api.template1.v1.BlogService/CreateBlog', request);
  }

  async updateBlog(request: UpdateBlogRequest): Promise<UpdateBlogResponse> {
    return this.call<UpdateBlogResponse>('/api.template1.v1.BlogService/UpdateBlog', request);
  }

  async deleteBlog(request: DeleteBlogRequest): Promise<DeleteBlogResponse> {
    return this.call<DeleteBlogResponse>('/api.template1.v1.BlogService/DeleteBlog', request);
  }

  async deleteBatchBlog(request: DeleteBatchBlogRequest): Promise<DeleteBatchBlogResponse> {
    return this.call<DeleteBatchBlogResponse>('/api.template1.v1.BlogService/DeleteBatchBlog', request);
  }

  async getBlog(request: GetBlogRequest): Promise<GetBlogResponse> {
    return this.call<GetBlogResponse>('/api.template1.v1.BlogService/GetBlog', request);
  }

  async queryBlog(request: QueryBlogRequest): Promise<QueryBlogResponse> {
    return this.call<QueryBlogResponse>('/api.template1.v1.BlogService/QueryBlog', request);
  }
}
