import type { ErrorBody, ResponseMeta } from '../types';

/**
 * API 错误类，封装 ErrorBody 语义
 */
export class ApiError extends Error {
  public readonly reason: string;
  public readonly details: unknown[];
  public readonly errors?: ErrorBody[];
  public readonly meta?: ResponseMeta;

  constructor(errorBody: ErrorBody, meta?: ResponseMeta) {
    super(errorBody.message);
    this.name = 'ApiError';
    this.reason = errorBody.reason;
    this.details = errorBody.details;
    this.errors = errorBody.errors;
    this.meta = meta;
  }

  /**
   * 从 Connect error 解析 ApiError
   */
  static fromConnectError(error: unknown, meta?: ResponseMeta): ApiError {
    // Connect error 结构解析
    if (isConnectError(error)) {
      const connectError = error as ConnectErrorLike;
      const errorBody = extractErrorBodyFromConnect(connectError);
      return new ApiError(errorBody, meta);
    }

    // 通用错误处理
    return new ApiError(
      {
        reason: 'UNKNOWN_ERROR',
        message: error instanceof Error ? error.message : String(error),
        details: [],
      },
      meta
    );
  }

  /**
   * 从 HTTP 响应解析 ApiError
   */
  static fromHttpResponse(response: Response, body: unknown, meta?: ResponseMeta): ApiError {
    if (isErrorBody(body)) {
      return new ApiError(body, meta);
    }

    return new ApiError(
      {
        reason: `HTTP_${response.status}`,
        message: response.statusText || 'HTTP Error',
        details: [body],
      },
      meta
    );
  }
}

/**
 * Connect error 接口（简化版）
 */
interface ConnectErrorLike {
  code: number;
  message: string;
  details: Array<{ type: string; value: unknown }>;
  rawMessage: string;
}

/**
 * 检查是否为 Connect error
 */
function isConnectError(error: unknown): error is ConnectErrorLike {
  return (
    typeof error === 'object' &&
    error !== null &&
    'code' in error &&
    'message' in error &&
    'details' in error
  );
}

/**
 * 检查是否为 ErrorBody
 */
function isErrorBody(body: unknown): body is ErrorBody {
  return (
    typeof body === 'object' &&
    body !== null &&
    'reason' in body &&
    'message' in body &&
    typeof (body as ErrorBody).reason === 'string' &&
    typeof (body as ErrorBody).message === 'string'
  );
}

/**
 * 从 Connect error details 中提取 ErrorBody
 */
function extractErrorBodyFromConnect(error: ConnectErrorLike): ErrorBody {
  // 查找 ErrorBody 类型的 detail
  const errorBodyDetail = error.details.find(
    (d) => d.type === 'type.googleapis.com/common.ErrorBody'
  );

  if (errorBodyDetail && isErrorBody(errorBodyDetail.value)) {
    return errorBodyDetail.value;
  }

  // 回退到从 message 解析
  return {
    reason: `CODE_${error.code}`,
    message: error.message,
    details: error.details,
  };
}

/**
 * 从响应 headers 提取元信息
 */
export function extractResponseMeta(headers: Headers): ResponseMeta {
  return {
    traceId: headers.get('x-trace-id') ?? undefined,
    success: headers.get('x-response-success') === 'true',
    errorReason: headers.get('x-error-reason') ?? undefined,
  };
}

/**
 * 判断是否为 ApiError
 */
export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError;
}
