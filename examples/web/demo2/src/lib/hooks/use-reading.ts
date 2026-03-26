/**
 * Reading 相关的 TanStack Query hooks
 * 使用 ts-rest 客户端对接 template2
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
    queryFn: async () => {
      const { status, body } = await api.template2.reading.queryBlog({
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
      throw new Error(`Failed to query reading blog: ${status}`);
    },
  });
}

/**
 * 获取博客详情（含阅读统计）
 */
export function useReadingBlog(id: string) {
  return useQuery({
    queryKey: readingKeys.detail(id),
    queryFn: async () => {
      const { status, body } = await api.template2.reading.getBlog({
        params: { id },
      });
      if (status === 200) {
        return body;
      }
      throw new Error(`Failed to get reading blog: ${status}`);
    },
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
    mutationFn: async (id: string) => {
      const { status, body } = await api.template2.reading.recordReading({
        params: { id },
        body: { id },
      });
      if (status === 200) {
        return body;
      }
      throw new Error(`Failed to record reading: ${status}`);
    },
    onSuccess: (_, id) => {
      // 刷新详情和列表
      queryClient.invalidateQueries({ queryKey: readingKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: readingKeys.lists() });
    },
  });
}
