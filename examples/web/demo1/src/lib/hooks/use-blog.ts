/**
 * Blog 相关的 TanStack Query hooks
 * 使用 connect-web 客户端
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
    queryFn: () =>
      api.template1.blog.queryBlog({
        page: {
          pageNo: params?.pageNo,
          pageSize: params?.pageSize,
        },
        title: params?.title,
        orderBy: params?.orderBy,
      }),
  });
}

/**
 * 获取博客详情
 */
export function useBlog(id: string) {
  return useQuery({
    queryKey: blogKeys.detail(id),
    queryFn: () => api.template1.blog.getBlog({ id }),
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
    mutationFn: (request: Parameters<typeof api.template1.blog.createBlog>[0]) =>
      api.template1.blog.createBlog(request),
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
    mutationFn: (request: Parameters<typeof api.template1.blog.updateBlog>[0]) =>
      api.template1.blog.updateBlog(request),
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
    mutationFn: (request: Parameters<typeof api.template1.blog.deleteBlog>[0]) =>
      api.template1.blog.deleteBlog(request),
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
    mutationFn: (request: Parameters<typeof api.template1.blog.deleteBatchBlog>[0]) =>
      api.template1.blog.deleteBatchBlog(request),
    onSuccess: () => {
      // 刷新列表
      queryClient.invalidateQueries({ queryKey: blogKeys.lists() });
    },
  });
}
