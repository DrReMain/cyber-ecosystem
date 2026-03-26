/**
 * Template1 BlogService ts-rest 契约定义
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

// ============ Blog 相关 Schema ============

const CreateBlogRequestSchema = z.object({
  title: z.string().optional(),
  content: z.string().optional(),
  publishedAt: z.string().datetime().optional().nullable(),
});

const CreateBlogResponseSchema = z.object({});

const UpdateBlogRequestSchema = z.object({
  id: z.string().length(20),
  title: z.string().optional(),
  content: z.string().optional(),
  publishedAt: z.string().datetime().optional().nullable(),
  fieldsMask: z.array(z.string()).min(1).optional(),
});

const UpdateBlogResponseSchema = z.object({});

const DeleteBlogRequestSchema = z.object({
  id: z.string().length(20),
});

const DeleteBlogResponseSchema = z.object({});

const DeleteBatchBlogRequestSchema = z.object({
  ids: z.array(z.string()).min(1),
});

const DeleteBatchBlogResponseSchema = z.object({
  count: z.number(),
});

const GetBlogResponseSchema = z.object({
  id: z.string(),
  createdAt: z.string().datetime(),
  updatedAt: z.string().datetime(),
  title: z.string().optional().nullable(),
  content: z.string().optional().nullable(),
  publishedAt: z.string().datetime().optional().nullable(),
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
  list: z.array(GetBlogResponseSchema),
});

// ============ BlogService 契约 ============

export const blogContract = c.router({
  createBlog: {
    method: 'POST',
    path: '/api/v1/blog',
    body: CreateBlogRequestSchema,
    responses: {
      200: CreateBlogResponseSchema,
    },
    summary: '创建博客',
  },
  updateBlog: {
    method: 'PUT',
    path: '/api/v1/blog/:id',
    pathParams: z.object({
      id: z.string().length(20),
    }),
    body: UpdateBlogRequestSchema,
    responses: {
      200: UpdateBlogResponseSchema,
    },
    summary: '更新博客',
  },
  deleteBlog: {
    method: 'DELETE',
    path: '/api/v1/blog/:id',
    pathParams: z.object({
      id: z.string().length(20),
    }),
    responses: {
      200: DeleteBlogResponseSchema,
    },
    summary: '删除博客',
  },
  deleteBatchBlog: {
    method: 'POST',
    path: '/api/v1/blog/delete-batch',
    body: DeleteBatchBlogRequestSchema,
    responses: {
      200: DeleteBatchBlogResponseSchema,
    },
    summary: '批量删除博客',
  },
  getBlog: {
    method: 'GET',
    path: '/api/v1/blog/:id',
    pathParams: z.object({
      id: z.string().length(20),
    }),
    responses: {
      200: GetBlogResponseSchema,
    },
    summary: '获取博客详情',
  },
  queryBlog: {
    method: 'GET',
    path: '/api/v1/blog',
    query: QueryBlogRequestSchema,
    responses: {
      200: QueryBlogResponseSchema,
    },
    summary: '查询博客列表',
  },
});
