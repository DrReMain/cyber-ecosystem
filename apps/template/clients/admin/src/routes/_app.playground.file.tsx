import { useMutation } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { App, Button, Card, Flex, Input, Tag, Upload } from "antd"
import { useState } from "react"
import { getDownloadUrl, uploadFile } from "#/services/file-client"

export const Route = createFileRoute("/_app/playground/file")({
  component: FilePlayground,
})

function FilePlayground() {
  return (
    <Flex className="p-4" flex={1} gap={16}>
      <UploadCard />
      <DownloadCard />
    </Flex>
  )
}

function UploadCard() {
  const { message } = App.useApp()

  const mutation = useMutation({
    mutationFn: uploadFile,
    onSuccess: () => message.success("上传成功"),
    onError: () => message.error("上传失败"),
  })

  return (
    <Card className="w-full">
      <Flex gap="small" vertical>
        <Upload.Dragger
          beforeUpload={(file) => {
            mutation.mutate(file)
            return false
          }}
          showUploadList={false}
        >
          <p className="text-antd-text-secondary text-sm">
            点击或拖拽文件到此区域上传
          </p>
        </Upload.Dragger>
        {mutation.error != null && (
          <Tag color="red">
            {mutation.error instanceof Error
              ? mutation.error.message
              : String(mutation.error)}
          </Tag>
        )}
        {mutation.data != null && (
          <pre className="max-h-48 overflow-auto rounded bg-antd-fill-quaternary p-2 text-xs">
            {JSON.stringify(mutation.data, null, 2)}
          </pre>
        )}
      </Flex>
    </Card>
  )
}

function DownloadCard() {
  const [id, setId] = useState("")

  return (
    <Card className="w-full">
      <Flex gap="small" vertical>
        <Input
          onChange={(e) => setId(e.target.value)}
          placeholder="输入文件 ID"
          value={id}
        />
        <Button
          color="primary"
          href={id ? getDownloadUrl(id.trim()) : undefined}
          target="_blank"
          variant="filled"
        >
          直链下载
        </Button>
      </Flex>
    </Card>
  )
}
