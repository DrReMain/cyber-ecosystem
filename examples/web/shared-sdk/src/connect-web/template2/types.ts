/**
 * Template2 ReadingService 类型定义
 * 基于 proto 契约生成
 */

import type { PageRequest, PageResponse } from '../../types';

// ============ Reading 相关类型 ============

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
  list: BlogWithReading[];
}

export interface BlogWithReading {
  id: string;
  createdAt: string;
  updatedAt: string;
  title?: string;
  content?: string;
  publishedAt?: string;
  readingCount: number | string; // OpenAPI 生成为 string
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
  readingCount: number | string;
}

export interface RecordReadingRequest {
  id: string;
}

export interface RecordReadingResponse {
  readingCount: number | string;
}
