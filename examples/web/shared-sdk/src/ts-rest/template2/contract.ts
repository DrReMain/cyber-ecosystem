/**
 * Template2 ReadingService ts-rest 契约定义
 * 基于 OpenAPI 规范生成
 */

import { initContract } from '@ts-rest/core';
import { z } from 'zod';

const c = initContract();

// ============ 公共类型 ============

const PageResponseSchema = z.object({
  pageNo: z.number(),
  pageSize: z.number(),
  total: z.number(),
  more: z.boolean(),
});

// ============ Reading 相关 Schema ============

const BlogWithReadingSchema = z.object({
  id: z.string(),
  createdAt: z.string().datetime(),
  updatedAt: z.string().datetime(),
  title: z.string().optional().nullable(),
  content: z.string().optional().nullable(),
  publishedAt: z.string().datetime().optional().nullable(),
  readingCount: z.union([z.number(), z.string()]), // OpenAPI 生成为 string
});

const QueryBlogRequestSchema = z.object({
  'page.pageNo': z.coerce.number().int().positive().optional(),
  'page.pageSize': z.coerce.number().int().positive().max(100).optional(),
  'page.all': z.coerce.boolean().optional(),
  'page.createdAtA': z.string().datetime().optional(),
  'page.createdAtZ': z.string().datetime().optional(),
  'page.updatedAtA': z.string().datetime().optional(),
  'page.updatedAtZ': z.string().datetime().optional(),
  id: z.string().length(20).optional(),
  title: z.string().optional(),
  publishedAtA: z.string().datetime().optional(),
  publishedAtZ: z.string().datetime().optional(),
  orderBy: z.array(z.string()).optional(),
});

const QueryBlogResponseSchema = z.object({
  page: PageResponseSchema,
  list: z.array(BlogWithReadingSchema),
});

const GetBlogResponseSchema = z.object({
  id: z.string(),
  createdAt: z.string().datetime(),
  updatedAt: z.string().datetime(),
  title: z.string().optional().nullable(),
  content: z.string().optional().nullable(),
  publishedAt: z.string().datetime().optional().nullable(),
  readingCount: z.union([z.number(), z.string()]),
});

const RecordReadingRequestSchema = z.object({
  id: z.string().length(20),
});

const RecordReadingResponseSchema = z.object({
  readingCount: z.union([z.number(), z.string()]),
});

// ============ ReadingService 契约 ============

export const readingContract = c.router({
  queryBlog: {
    method: 'GET',
    path: '/api/v1/reading/blog',
    query: QueryBlogRequestSchema,
    responses: {
      200: QueryBlogResponseSchema,
    },
    summary: '查询博客列表（含阅读统计）',
  },
  getBlog: {
    method: 'GET',
    path: '/api/v1/reading/blog/:id',
    pathParams: z.object({
      id: z.string().length(20),
    }),
    responses: {
      200: GetBlogResponseSchema,
    },
    summary: '获取博客详情（含阅读统计）',
  },
  recordReading: {
    method: 'POST',
    path: '/api/v1/reading/blog/:id/read',
    pathParams: z.object({
      id: z.string().length(20),
    }),
    body: RecordReadingRequestSchema,
    responses: {
      200: RecordReadingResponseSchema,
    },
    summary: '记录阅读',
  },
});
