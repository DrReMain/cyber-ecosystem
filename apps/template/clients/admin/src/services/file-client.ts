import ky from "ky"
import { env, resolveApiBaseUrl } from "#/env"
import { getLocale } from "#/paraglide/runtime"

const baseUrl = resolveApiBaseUrl(env.HTTP_API_URL, "/http")

export interface FileInfo {
  id: string
  name: string
  contentType: string
  size: number
  status: string
  createdAt?: string
  updatedAt?: string
}

const fileApi = ky.create({
  prefix: `${baseUrl}/api/v1`,
  headers: { "Accept-Language": getLocale() },
})

function mapFileInfo(raw: Record<string, unknown>): FileInfo {
  return {
    id: raw.ID as string,
    name: raw.Name as string,
    contentType: raw.ContentType as string,
    size: raw.Size as number,
    status: raw.Status as string,
    createdAt: raw.CreatedAt as string | undefined,
    updatedAt: raw.UpdatedAt as string | undefined,
  }
}

export async function uploadFile(file: File): Promise<FileInfo> {
  const form = new FormData()
  form.append("file", file)
  const res = await fileApi.post("files", { body: form })
  return mapFileInfo(await res.json<Record<string, unknown>>())
}

export async function downloadFile(id: string): Promise<Blob> {
  const res = await fileApi.get(`files/${id}`)
  return res.blob()
}

export function getDownloadUrl(id: string): string {
  return `${baseUrl}/api/v1/files/${id}`
}
