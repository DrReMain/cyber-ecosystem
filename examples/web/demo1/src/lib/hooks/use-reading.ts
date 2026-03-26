/**
 * Reading 相关的 TanStack Query hooks
 * 使用 connect-web 客户端对接 template2
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api.ts';

// ============ Query Keys ============

export const readingKeys = {
  all: ['reading'] as const,
  lists: () => [...readingKeys.all, 'list'] as const,
  list: (params?: Record<string, unknown>) => [...readingKeys.lists(), params] as const,
  details: () => [...readingKeys.all, 'detail'] as const,
  detail: (id: string) => [...readingKeys.details(), id] as const,
};

// ============ Query Hooks ============

/**
 * 查询博客列表（含阅读统计）
 */
export function useReadingBlogList(params?: {
  pageNo?: number;
  pageSize?: number;
  title?: string;
  orderBy?: string[];
}) {
  return useQuery({
    queryKey: readingKeys.list(params),
    queryFn: () =>
      api.template2.reading.queryBlog({
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
 * 获取博客详情（含阅读统计）
 */
export function useReadingBlog(id: string) {
  return useQuery({
    queryKey: readingKeys.detail(id),
    queryFn: () => api.template2.reading.getBlog({ id }),
    enabled: !!id,
  });
}

// ============ Mutation Hooks ============

/**
 * 记录阅读
 */
export function useRecordReading() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: Parameters<typeof api.template2.reading.recordReading>[0]) =>
      api.template2.reading.recordReading(request),
    onSuccess: (_, variables) => {
      // 刷新详情和列表
      queryClient.invalidateQueries({ queryKey: readingKeys.detail(variables.id) });
      queryClient.invalidateQueries({ queryKey: readingKeys.lists() });
    },
  });
}
