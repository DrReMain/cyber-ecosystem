/**
 * Blog 相关的 TanStack Query hooks
 * 使用 ts-rest 客户端对接 template1
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api.ts';

// ============ Query Keys ============

export const blogKeys = {
  all: ['blog'] as const,
  lists: () => [...blogKeys.all, 'list'] as const,
  list: (params?: Record<string, unknown>) => [...blogKeys.lists(), params] as const,
  details: () => [...blogKeys.all, 'detail'] as const,
  detail: (id: string) => [...blogKeys.details(), id] as const,
};

// ============ Query Hooks ============

/**
 * 查询博客列表
 */
export function useBlogList(params?: {
  pageNo?: number;
  pageSize?: number;
  title?: string;
  orderBy?: string[];
}) {
  return useQuery({
    queryKey: blogKeys.list(params),
    queryFn: async () => {
      const { status, body } = await api.template1.blog.queryBlog({
        query: {
          'page.pageNo': params?.pageNo,
          'page.pageSize': params?.pageSize,
          title: params?.title,
          orderBy: params?.orderBy,
        },
      });
      if (status === 200) {
        return body;
      }
      throw new Error(`Failed to query blog: ${status}`);
    },
  });
}

/**
 * 获取博客详情
 */
export function useBlog(id: string) {
  return useQuery({
    queryKey: blogKeys.detail(id),
    queryFn: async () => {
      const { status, body } = await api.template1.blog.getBlog({
        params: { id },
      });
      if (status === 200) {
        return body;
      }
      throw new Error(`Failed to get blog: ${status}`);
    },
    enabled: !!id,
  });
}

// ============ Mutation Hooks ============

/**
 * 创建博客
 */
export function useCreateBlog() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: { title?: string; content?: string; publishedAt?: string }) => {
      const { status, body } = await api.template1.blog.createBlog({
        body: data,
      });
      if (status === 200) {
        return body;
      }
      throw new Error(`Failed to create blog: ${status}`);
    },
    onSuccess: () => {
      // 刷新列表
      queryClient.invalidateQueries({ queryKey: blogKeys.lists() });
    },
  });
}

/**
 * 更新博客
 */
export function useUpdateBlog() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: {
      id: string;
      title?: string;
      content?: string;
      publishedAt?: string;
      fieldsMask?: string[];
    }) => {
      const { status, body } = await api.template1.blog.updateBlog({
        params: { id: data.id },
        body: data,
      });
      if (status === 200) {
        return body;
      }
      throw new Error(`Failed to update blog: ${status}`);
    },
    onSuccess: (_, variables) => {
      // 刷新详情和列表
      queryClient.invalidateQueries({ queryKey: blogKeys.detail(variables.id) });
      queryClient.invalidateQueries({ queryKey: blogKeys.lists() });
    },
  });
}

/**
 * 删除博客
 */
export function useDeleteBlog() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: string) => {
      const { status, body } = await api.template1.blog.deleteBlog({
        params: { id },
      });
      if (status === 200) {
        return body;
      }
      throw new Error(`Failed to delete blog: ${status}`);
    },
    onSuccess: () => {
      // 刷新列表
      queryClient.invalidateQueries({ queryKey: blogKeys.lists() });
    },
  });
}

/**
 * 批量删除博客
 */
export function useDeleteBatchBlog() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (ids: string[]) => {
      const { status, body } = await api.template1.blog.deleteBatchBlog({
        body: { ids },
      });
      if (status === 200) {
        return body;
      }
      throw new Error(`Failed to batch delete blog: ${status}`);
    },
    onSuccess: () => {
      // 刷新列表
      queryClient.invalidateQueries({ queryKey: blogKeys.lists() });
    },
  });
}
