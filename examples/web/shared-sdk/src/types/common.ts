/**
 * 分页请求参数
 */
export interface PageRequest {
  pageNo?: number;
  pageSize?: number;
  all?: boolean;
  createdAtA?: string;
  createdAtZ?: string;
  updatedAtA?: string;
  updatedAtZ?: string;
}

/**
 * 分页响应信息
 */
export interface PageResponse {
  pageNo: number;
  pageSize: number;
  total: number;
  more: boolean;
}

/**
 * 错误详情结构
 * 符合项目契约 common.ErrorBody
 */
export interface ErrorBody {
  reason: string;
  message: string;
  details: unknown[];
  errors?: ErrorBody[];
}

/**
 * API 响应元信息（从 header 获取）
 */
export interface ResponseMeta {
  traceId?: string;
  success: boolean;
  errorReason?: string;
}

/**
 * 客户端配置基类
 */
export interface ClientConfig {
  baseUrl: string;
  headers?: Record<string, string>;
  timeout?: number;
}
