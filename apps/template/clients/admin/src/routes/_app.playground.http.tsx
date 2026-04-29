import { Filter } from "@shared/antd/filter"
import { TableToolbar, useTable } from "@shared/antd/table"
import {
  buildSearchPatch,
  pageNoField,
  pageSizeField,
  sortField,
  useFilter,
  usePagination,
  useSort,
  useUrlSearchStore,
} from "@shared/antd/use-search"
import { useMutation, useQuery } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import {
  App,
  Button,
  Card,
  Descriptions,
  Drawer,
  Flex,
  Form,
  Input,
  Modal,
  Space,
  Table,
  Typography,
} from "antd"
import {
  type Ref,
  useEffect,
  useImperativeHandle,
  useMemo,
  useRef,
  useState,
} from "react"
import { z } from "zod"
import { buildHTTPPage, buildOrderBy, defaultKV } from "#/lib/builder"
import { useUtils } from "#/lib/use-utils"
import { m } from "#/paraglide/messages"
import "#/services/http-client"
import {
  messageServiceCreateMessage,
  messageServiceDeleteMessage,
  messageServiceGetMessage,
  messageServiceQueryMessage,
  messageServiceUpdateMessage,
} from "#/services/openapi/sdk.gen"
import type { ApiTemplateV1GetMessageResponse } from "#/services/openapi/types.gen"

const schema = z.object({
  pageNo: pageNoField(),
  pageSize: pageSizeField(),
  sort: sortField(),
  createdAtA: z.number().optional(),
  createdAtZ: z.number().optional(),
  updatedAtA: z.number().optional(),
  updatedAtZ: z.number().optional(),
  title: z.string().optional(),
  status: z.string().nullable().default("draft"),
})

export const Route = createFileRoute("/_app/playground/http")({
  validateSearch: schema.parse,
  component: HttpPlayground,
})

function HttpPlayground() {
  const search = Route.useSearch()
  const navigate = Route.useNavigate()
  const initialData = Route.useLoaderData()

  const store = useUrlSearchStore(search, {
    onNavigate: (patch) =>
      navigate({
        search: (prev) => buildSearchPatch(prev, patch),
        replace: true,
      }),
  })
  const { values, onFilter, onReset } = useFilter(store, schema)
  const { onPageChange } = usePagination(store, schema)
  const { sort } = useSort(store, schema)
  const { tableSize, setTableSize } = useTable()

  const { fieldTimestamp, fieldAction, fieldCopy } = useUtils()
  const { message } = App.useApp()
  const detailRef = useRef<DetailModalRef>(null)
  const createRef = useRef<CreateDrawerRef>(null)
  const updateRef = useRef<UpdateDrawerRef>(null)

  const query = useMemo(
    () => ({
      ...buildHTTPPage(search),
      orderBy: buildOrderBy(sort),
      ...defaultKV("title", search.title),
      ...defaultKV("status", search.status),
    }),
    [search, sort],
  )

  const { data, isFetching, refetch } = useQuery({
    queryKey: ["queryMessage", query],
    queryFn: async () => {
      const res = await messageServiceQueryMessage({
        query,
        throwOnError: true,
      })
      return res.data
    },
    initialData,
  })

  const deleteMut = useMutation({
    mutationFn: (id: string) =>
      messageServiceDeleteMessage({ path: { id }, throwOnError: true }),
    onSuccess: () => Promise.all([refetch(), message.success("删除成功")]),
  })

  return (
    <Flex className="p-4" flex={1} gap={16} vertical>
      <Card>
        <Filter
          columns={4}
          initialValues={values}
          labels={{
            search: m.filter_search(),
            reset: m.filter_reset(),
            fold: m.filter_fold(),
            expand: m.filter_expand(),
          }}
          onFilter={onFilter}
          onReset={onReset}
          options={[
            { label: "标题", name: "title", placeholder: "模糊搜索" },
            {
              label: "状态",
              name: "status",
              type: "select",
              options: [
                { value: "draft", label: "未发布" },
                { value: "published", label: "已发布" },
                { value: "archived", label: "已归档" },
              ],
              placeholder: "不限",
            },
            {
              label: "创建时间",
              name: ["createdAtA", "createdAtZ"],
              type: "range-date",
            },
            {
              label: "修改时间",
              name: ["updatedAtA", "updatedAtZ"],
              type: "range-datetime",
            },
          ]}
        />
      </Card>

      <Card>
        <Flex gap={16} vertical>
          <TableToolbar
            extra={
              <Space>
                <Button
                  color="primary"
                  onClick={() => createRef.current?.open()}
                  variant="filled"
                >
                  新增
                </Button>
              </Space>
            }
            labels={{
              refresh: m.toolbar_refresh(),
              density: m.toolbar_density(),
              densityLarge: m.toolbar_density_large(),
              densityMiddle: m.toolbar_density_middle(),
              densitySmall: m.toolbar_density_small(),
            }}
            loading={isFetching}
            onRefresh={() => refetch()}
            onSizeChange={setTableSize}
            size={tableSize}
          />
          <Table<ApiTemplateV1GetMessageResponse>
            columns={[
              fieldCopy("id", { title: "ID", fixed: "left" }),
              { title: "标题", dataIndex: "title" },
              {
                title: "内容",
                dataIndex: "content",
                render: (v: string) => (
                  <Typography.Text
                    className="max-w-lg"
                    ellipsis={{ tooltip: v }}
                  >
                    {v}
                  </Typography.Text>
                ),
              },
              { title: "状态", dataIndex: "status" },
              fieldTimestamp("createdAt", { title: m.column_created_at() }),
              fieldTimestamp("updatedAt", { title: m.column_updated_at() }),
              fieldAction((_, { id }) => (
                <Space size="small">
                  <Button
                    color="primary"
                    onClick={() => detailRef.current?.open(id ?? "")}
                    size="small"
                    variant="filled"
                  >
                    查看
                  </Button>
                  <Button
                    color="primary"
                    onClick={() => updateRef.current?.open(id ?? "")}
                    size="small"
                    variant="text"
                  >
                    修改
                  </Button>
                  <Button
                    color="danger"
                    disabled={deleteMut.isPending}
                    onClick={() => deleteMut.mutate(id ?? "")}
                    size="small"
                    variant="text"
                  >
                    删除
                  </Button>
                </Space>
              )),
            ]}
            dataSource={data?.list}
            loading={isFetching}
            pagination={{
              current: data?.page?.pageNo,
              pageSize: data?.page?.pageSize,
              total: data?.page?.total,
              showSizeChanger: true,
              onChange: onPageChange,
            }}
            rowKey="id"
            scroll={{ x: "max-content" }}
            size={tableSize}
          />
        </Flex>
      </Card>

      <DetailModal ref={detailRef} />
      <CreateDrawer onSuccess={() => refetch()} ref={createRef} />
      <UpdateDrawer onSuccess={() => refetch()} ref={updateRef} />
    </Flex>
  )
}

interface CreateDrawerRef {
  open: () => void
}
const CreateDrawer = function CreateDrawer({
  ref,
  onSuccess,
}: Readonly<{ ref: Ref<CreateDrawerRef>; onSuccess: () => void }>) {
  const [open, setOpen] = useState(false)
  const [form] = Form.useForm()
  const { message } = App.useApp()

  useImperativeHandle(ref, () => ({
    open: () => {
      form.resetFields()
      setOpen(true)
    },
  }))

  const createMut = useMutation({
    mutationFn: (values: { title: string; content: string }) =>
      messageServiceCreateMessage({ body: values, throwOnError: true }),
    onSuccess: () => {
      setOpen(false)
      message.success("创建成功")
      onSuccess()
    },
  })

  return (
    <Drawer
      destroyOnHidden
      footer={
        <Button
          color="primary"
          loading={createMut.isPending}
          onClick={() => form.submit()}
          variant="filled"
        >
          提交
        </Button>
      }
      onClose={() => setOpen(false)}
      open={open}
      title="新增消息"
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={(values) => createMut.mutate(values)}
      >
        <Form.Item label="标题" name="title" rules={[{ required: true }]}>
          <Input placeholder="请输入标题" />
        </Form.Item>
        <Form.Item label="内容" name="content" rules={[{ required: true }]}>
          <Input.TextArea placeholder="请输入内容" rows={4} />
        </Form.Item>
      </Form>
    </Drawer>
  )
}

interface UpdateDrawerRef {
  open: (id: string) => void
}
const UpdateDrawer = function UpdateDrawer({
  ref,
  onSuccess,
}: Readonly<{ ref: Ref<UpdateDrawerRef>; onSuccess: () => void }>) {
  const [id, setId] = useState("")
  const [form] = Form.useForm()
  const { message } = App.useApp()

  useImperativeHandle(ref, () => ({
    open: (_) => {
      form.resetFields()
      setId(_)
    },
  }))

  const { data } = useQuery({
    queryKey: ["getMessage", id],
    queryFn: async () => {
      const res = await messageServiceGetMessage({
        path: { id },
        throwOnError: true,
      })
      return res.data
    },
    enabled: !!id,
  })

  useEffect(() => {
    if (data) {
      form.setFieldsValue({ title: data.title, content: data.content })
    }
  }, [data, form])

  const updateMut = useMutation({
    mutationFn: (values: { title: string; content: string }) =>
      messageServiceUpdateMessage({
        path: { id },
        body: { ...values, fieldsMask: ["title", "content"] },
        throwOnError: true,
      }),
    onSuccess: () => {
      setId("")
      message.success("修改成功")
      onSuccess()
    },
  })

  return (
    <Drawer
      destroyOnHidden
      footer={
        <Button
          color="primary"
          loading={updateMut.isPending}
          onClick={() => form.submit()}
          variant="filled"
        >
          提交
        </Button>
      }
      onClose={() => setId("")}
      open={!!id}
      title="修改消息"
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={(values) => updateMut.mutate(values)}
      >
        <Form.Item label="标题" name="title" rules={[{ required: true }]}>
          <Input placeholder="请输入标题" />
        </Form.Item>
        <Form.Item label="内容" name="content" rules={[{ required: true }]}>
          <Input.TextArea placeholder="请输入内容" rows={4} />
        </Form.Item>
      </Form>
    </Drawer>
  )
}

interface DetailModalRef {
  open: (id: string) => void
}
const DetailModal = function DetailModal({
  ref,
}: Readonly<{ ref: Ref<DetailModalRef> }>) {
  const [id, setId] = useState("")
  const { formatTime } = useUtils()

  useImperativeHandle(ref, () => ({
    open: (_) => {
      setId(_)
    },
  }))

  const { data, isFetching } = useQuery({
    queryKey: ["getMessage", id],
    queryFn: async () => {
      const res = await messageServiceGetMessage({
        path: { id },
        throwOnError: true,
      })
      return res.data
    },
    enabled: !!id,
  })

  return (
    <Modal
      destroyOnHidden
      footer={null}
      loading={isFetching}
      onCancel={() => setId("")}
      open={!!id}
      title="消息详情"
    >
      <Descriptions column={1}>
        <Descriptions.Item label="">{data?.id}</Descriptions.Item>
        <Descriptions.Item label="">{data?.title}</Descriptions.Item>
        <Descriptions.Item label="">{data?.content}</Descriptions.Item>
        <Descriptions.Item label="">{data?.status}</Descriptions.Item>
        <Descriptions.Item label="">
          {formatTime(data?.createdAt)}
        </Descriptions.Item>
        <Descriptions.Item label="">
          {formatTime(data?.updatedAt)}
        </Descriptions.Item>
      </Descriptions>
    </Modal>
  )
}
