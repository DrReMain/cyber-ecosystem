/**
 * Template1 BlogService 类型定义
 * 基于 proto 契约生成
 */

import type { PageRequest, PageResponse } from '../../types';

// ============ Blog 相关类型 ============

export interface CreateBlogRequest {
  title?: string;
  content?: string;
  publishedAt?: string; // ISO 8601 datetime
}

export interface CreateBlogResponse {}

export interface UpdateBlogRequest {
  id: string;
  title?: string;
  content?: string;
  publishedAt?: string;
  fieldsMask?: string[];
}

export interface UpdateBlogResponse {}

export interface DeleteBlogRequest {
  id: string;
}

export interface DeleteBlogResponse {}

export interface DeleteBatchBlogRequest {
  ids: string[];
}

export interface DeleteBatchBlogResponse {
  count: number;
}

export interface GetBlogRequest {
  id: string;
}

export interface GetBlogResponse {
  id: string;
  createdAt: string;
  updatedAt: string;
  title?: string;
  content?: string;
  publishedAt?: string;
}

export interface QueryBlogRequest {
  page?: PageRequest;
  id?: string;
  title?: string;
  publishedAtA?: string;
  publishedAtZ?: string;
  orderBy?: string[];
}

export interface QueryBlogResponse {
  page: PageResponse;
  list: GetBlogResponse[];
}
