# Slice 6: Storage Service

> **闭环标准**: 文件上传/下载/删除正常，通过 thin client auth 校验身份。

## 目标

构建独立的文件存储服务，管理文件元数据（PostgreSQL）和物理存储（本地文件系统/MinIO/S3）。业务服务通过 thin client 或直接 gRPC/HTTP 调用。

此 Slice 验证 thin client 模块（Slice 5）在真实服务中的集成效果。

## 前置条件

- Slice 5 完成（thin client 模块可用）
- IAM 服务运行中（提供认证）
- PostgreSQL 和 Redis 在本地可用

## 端口

| Transport | Address |
|-----------|---------|
| HTTP | `0.0.0.0:11002` |
| gRPC | `0.0.0.0:12002` |
| ConnectRPC | `0.0.0.0:13002` |
| Ops | `0.0.0.0:14002` |

---

## Step 1: 目录结构 + Nx 配置

```bash
mkdir -p apps/yggdrasil/services/storage/{cmd/app,configs,internal/{conf,data/ent/schema,biz,server/{locales},service,pkg/storage}}
```

**文件**: `apps/yggdrasil/services/storage/project.json`

```json
{
  "name": "yggdrasil_storage",
  "$schema": "../../../../node_modules/nx/schemas/project-schema.json",
  "implicitDependencies": ["yggdrasil_api", "shared-go"],
  "targets": {
    "proto:conf": {
      "executor": "nx:run-commands",
      "options": {
        "command": "buf generate --template apps/yggdrasil/services/storage/buf.gen.conf.yaml --path apps/yggdrasil/services/storage/internal/conf"
      }
    },
    "generate": {
      "dependsOn": ["yggdrasil_api:proto:api", "proto:conf"],
      "executor": "nx:run-commands",
      "options": {
        "command": "cd apps/yggdrasil/services/storage && wire ./cmd/app/..."
      }
    },
    "dev": {
      "executor": "nx:run-commands",
      "options": {
        "command": "cd apps/yggdrasil/services/storage && go run ./cmd/app/... -conf ./configs"
      }
    },
    "build": {
      "executor": "nx:run-commands",
      "options": {
        "command": "cd apps/yggdrasil/services/storage && go build -o ./bin/storage ./cmd/app/..."
      }
    },
    "ent:new": {
      "executor": "nx:run-commands",
      "options": {
        "command": "cd apps/yggdrasil/services/storage && go run -mod=mod entgo.io/ent/cmd/ent new --target internal/data/ent/schema"
      }
    }
  }
}
```

`buf.gen.conf.yaml` 与 Slice 1 相同结构。

---

## Step 2: API Proto 定义

**文件**: `apps/yggdrasil/api/v1/storage_file.proto`

```protobuf
syntax = "proto3";

// buf:lint:ignore PACKAGE_DIRECTORY_MATCH
package api.yggdrasil.v1;

import "buf/validate/validate.proto";
import "common/page.proto";
import "desc/desc.proto";
import "google/api/annotations.proto";
import "google/protobuf/wrappers.proto";

option go_package = "cyber-ecosystem/apps/yggdrasil/gen/go/v1;yggdrasilV1";

// StorageService
service StorageService {
  option (desc.service_comment) = "文件存储服务";

  // UploadFile 上传文件（multipart）
  rpc UploadFile(UploadFileRequest) returns (UploadFileResponse) {
    option (desc.method_comment) = "上传文件";
    option (google.api.http) = {post: "/storage/files" body: "*"};
  }

  // GetFile 获取文件元数据
  rpc GetFile(GetFileRequest) returns (GetFileResponse) {
    option (desc.method_comment) = "获取文件信息";
    option (google.api.http) = {get: "/storage/files/{id}"};
  }

  // DownloadFile 下载文件
  rpc DownloadFile(DownloadFileRequest) returns (DownloadFileResponse) {
    option (desc.method_comment) = "下载文件";
    option (google.api.http) = {get: "/storage/files/{id}/download"};
  }

  // DeleteFile 删除文件
  rpc DeleteFile(DeleteFileRequest) returns (DeleteFileResponse) {
    option (desc.method_comment) = "删除文件";
    option (google.api.http) = {delete: "/storage/files/{id}"};
  }

  // QueryFiles 查询文件列表
  rpc QueryFiles(QueryFilesRequest) returns (QueryFilesResponse) {
    option (desc.method_comment) = "查询文件列表";
    option (google.api.http) = {post: "/storage/files/query" body: "*"};
  }
}

message UploadFileRequest {
  string name = 1 [(buf.validate.field).string.min_len = 1];
  string content_type = 2;
  string bucket = 3;
  bytes data = 4;
}

message UploadFileResponse {
  string id = 1;
  string name = 2;
  string url = 3;
}

message GetFileRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
}

message GetFileResponse {
  string id = 1;
  string name = 2;
  string content_type = 3;
  int64 size = 4;
  string bucket = 5;
  string url = 6;
  string uploader_id = 7;
  string created_at = 8;
}

message DownloadFileRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
}

message DownloadFileResponse {
  string name = 1;
  string content_type = 2;
  bytes data = 3;
}

message DeleteFileRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
}

message DeleteFileResponse {}

message QueryFilesRequest {
  common.PageRequest page = 1;
  optional string bucket = 2;
  optional string uploader_id = 3;
  repeated string order_by = 100;
}

message QueryFilesResponse {
  common.PageResponse page = 1;
  repeated GetFileResponse list = 2;
}
```

生成:

```bash
buf lint apps/yggdrasil/api/v1/storage_file.proto
./nx run yggdrasil_api:proto:api
```

提交:
```bash
git add apps/yggdrasil/api/v1/storage_file.proto apps/yggdrasil/gen/
git commit -m "feat(yggdrasil): add storage service API proto definition"
```

---

## Step 3: Conf Proto + 生成

与 Slice 1 结构相同，额外添加 Storage 配置:

```protobuf
message Storage {
  message Local {
    string path = 1;  // e.g. "./uploads"
  }
  message Minio {
    string endpoint = 1;
    string access_key = 2;
    string secret_key = 3;
    bool secure = 4;
  }
  string backend = 1;  // "local" or "minio"
  Local local = 2;
  Minio minio = 3;
  int64 max_file_size = 4; // bytes, 0 = unlimited
}

// Bootstrap 追加:
message Bootstrap {
  // ... 同 Slice 1
  Storage storage = 7;
}
```

---

## Step 4: Pkg 层 — Storage Backend 接口

**文件**: `apps/yggdrasil/services/storage/internal/pkg/storage/backend.go`

```go
package storage

import (
	"context"
	"io"
)

// Backend is the interface for file storage backends.
type Backend interface {
	// Save stores a file and returns the storage path/key.
	Save(ctx context.Context, bucket, name string, data []byte) (string, error)
	// Load retrieves a file by its storage path/key.
	Load(ctx context.Context, bucket, key string) ([]byte, error)
	// Delete removes a file by its storage path/key.
	Delete(ctx context.Context, bucket, key string) error
	// URL returns a URL for accessing the file.
	URL(bucket, key string) string
}
```

**文件**: `apps/yggdrasil/services/storage/internal/pkg/storage/local.go`

```go
package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type LocalBackend struct {
	basePath string
	baseURL  string
}

func NewLocalBackend(basePath, baseURL string) *LocalBackend {
	return &LocalBackend{basePath: basePath, baseURL: baseURL}
}

func (b *LocalBackend) Save(ctx context.Context, bucket, name string, data []byte) (string, error) {
	dir := filepath.Join(b.basePath, bucket)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	key := fmt.Sprintf("%s/%s", bucket, name)
	fullPath := filepath.Join(b.basePath, key)
	if err := os.WriteFile(fullPath, data, 0o644); err != nil {
		return "", err
	}
	return key, nil
}

func (b *LocalBackend) Load(ctx context.Context, bucket, key string) ([]byte, error) {
	fullPath := filepath.Join(b.basePath, key)
	return os.ReadFile(fullPath)
}

func (b *LocalBackend) Delete(ctx context.Context, bucket, key string) error {
	fullPath := filepath.Join(b.basePath, key)
	return os.Remove(fullPath)
}

func (b *LocalBackend) URL(bucket, key string) string {
	return fmt.Sprintf("%s/%s", b.baseURL, key)
}
```

> MinIO 后端实现类似，使用 `minio-go` SDK。

---

## Step 5: Ent Schema

**文件**: `apps/yggdrasil/services/storage/internal/data/ent/schema/file.go`

```go
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

type File struct {
	ent.Schema
}

func (File) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MaxLen(500),
		field.String("content_type").Default("").MaxLen(200),
		field.Int64("size").Default(0),
		field.String("bucket").Default("default").MaxLen(100),
		field.String("storage_key").NotEmpty().MaxLen(500),
		field.String("url").Default("").MaxLen(1000),
		field.String("uploader_id").Default("").MaxLen(20),
		field.Bool("deleted").Default(false),
	}
}

func (File) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (File) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("bucket"),
		index.Fields("uploader_id"),
		index.Fields("deleted"),
	}
}

func (File) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "file"},
	}
}
```

生成:
```bash
cd apps/yggdrasil/services/storage && go generate ./internal/data/ent/...
```

---

## Step 6: Biz 层

**文件**: `apps/yggdrasil/services/storage/internal/biz/biz.go`

```go
package biz

import (
	"context"

	"github.com/google/wire"
	"github.com/go-kratos/kratos/v2/log"
)

type Transaction interface {
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type UC struct {
	log *log.Helper
	tm  Transaction
}

var ProviderSet = wire.NewSet(
	NewFileUC,
)
```

**文件**: `apps/yggdrasil/services/storage/internal/biz/uc_file.go`

```go
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/contracts/go/common"
	"cyber-ecosystem/shared-go/utils"
)

type File struct {
	ID          *string
	Name        *string
	ContentType *string
	Size        *int64
	Bucket      *string
	StorageKey  *string
	URL         *string
	UploaderID  *string
	Deleted     *bool
	CreatedAt   *time.Time
}

type FileQueryIn struct {
	*common.PageRequest
	OrderBy    []*utils.OrderBy
	Bucket     *string
	UploaderID *string
}

type FileQueryOut struct {
	*common.PageResponse
	List []*File
}

type FileRP interface {
	Create(ctx context.Context, f *File) (*File, error)
	Get(ctx context.Context, id string) (*File, error)
	Delete(ctx context.Context, id string) (*File, error)
	Query(ctx context.Context, in *FileQueryIn) (*FileQueryOut, error)
}

type StorageBackend interface {
	Save(ctx context.Context, bucket, name string, data []byte) (string, error)
	Load(ctx context.Context, bucket, key string) ([]byte, error)
	Delete(ctx context.Context, bucket, key string) error
	URL(bucket, key string) string
}

type FileUC struct {
	UC
	fileRP  FileRP
	backend StorageBackend
}

func NewFileUC(logger log.Logger, tm Transaction, fileRP FileRP, backend StorageBackend) *FileUC {
	return &FileUC{
		UC:      UC{log: log.NewHelper(log.With(logger, "module", "biz/uc_file")), tm: tm},
		fileRP:  fileRP,
		backend: backend,
	}
}

func (uc *FileUC) Upload(ctx context.Context, name, contentType, bucket string, data []byte) (*File, error) {
	key, err := uc.backend.Save(ctx, bucket, name, data)
	if err != nil {
		return nil, err
	}
	url := uc.backend.URL(bucket, key)
	file := &File{
		Name:        &name,
		ContentType: &contentType,
		Size:        ptrInt64(int64(len(data))),
		Bucket:      &bucket,
		StorageKey:  &key,
		URL:         &url,
	}
	return uc.fileRP.Create(ctx, file)
}

func (uc *FileUC) Download(ctx context.Context, id string) ([]byte, *File, error) {
	file, err := uc.fileRP.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	data, err := uc.backend.Load(ctx, *file.Bucket, *file.StorageKey)
	if err != nil {
		return nil, nil, err
	}
	return data, file, nil
}

func (uc *FileUC) Delete(ctx context.Context, id string) error {
	file, err := uc.fileRP.Delete(ctx, id)
	if err != nil {
		return err
	}
	// Best-effort physical deletion
	_ = uc.backend.Delete(ctx, *file.Bucket, *file.StorageKey)
	return nil
}

func (uc *FileUC) Get(ctx context.Context, id string) (*File, error) {
	return uc.fileRP.Get(ctx, id)
}

func (uc *FileUC) Query(ctx context.Context, in *FileQueryIn) (*FileQueryOut, error) {
	return uc.fileRP.Query(ctx, in)
}
```

---

## Step 7: Data 层 + Service 层 + Server 层

按 Slice 1 模式创建:

**Data 层:**
- `data/store.go`, `data/store_ent.go`, `data/store_cache.go` — 通用模式
- `data/rp_file.go` — CRUD + Query（与 Slice 1 的 rp_audit_log.go 模式相同）
- DB 名称: `cyber_ecosystem_yggdrasil_storage`
- `data/data.go` ProviderSet 包含 NewFileRP

**Service 层:**
- `service/service.go` — RegistrarList 包含 StorageService
- `service/storage.go` — 实现 StorageServiceServer
  - UploadFile: 调用 fileUC.Upload
  - DownloadFile: 调用 fileUC.Download，返回 bytes
  - GetFile / DeleteFile / QueryFiles: 标准 CRUD

**Server 层:**
- `server/server.go` — buildMiddlewares 使用 thin client auth + rbac
  ```go
  import (
      capauth "cyber-ecosystem/shared-go/capabilities/auth"
      caprbac "cyber-ecosystem/shared-go/capabilities/rbac"
  )
  ```
- 中间件链: `i18n → recovery → ratelimit → metrics → tracing → metadata → logging → selector(auth → rbac) → validate`
- grpc.go / http.go / connect.go / ops.go / i18n.go + locales — 与 Slice 1 结构相同

**Wire:**
- `cmd/app/wire.go` — `server.ProviderSet, service.ProviderSet, biz.ProviderSet, data.ProviderSet, newApp`
- `cmd/app/main.go` — 标准模式

**Config:**
- `configs/config.yaml` — HTTP `0.0.0.0:11002`, gRPC `0.0.0.0:12002`, ConnectRPC `0.0.0.0:13002`, Ops `0.0.0.0:14002`

---

## Step 8: 编译闭环

```bash
cd apps/yggdrasil/services/storage && go mod tidy
./nx run yggdrasil_storage:generate
./nx run yggdrasil_storage:build
```

提交:
```bash
git add apps/yggdrasil/services/storage/
git commit -m "feat(yggdrasil): storage service — first full build"
```

---

## Step 9: 集成验证

### 9a: 创建数据库 + 启动

```bash
psql -h localhost -U postgres -c "CREATE DATABASE cyber_ecosystem_yggdrasil_storage;"
./nx run yggdrasil_storage:dev
```

### 9b: 获取 token

```bash
TOKEN=$(curl -s -X POST http://localhost:11000/iam/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "admin123"}' | jq -r '.access_token')
```

### 9c: 上传文件

```bash
curl -X POST http://localhost:11002/storage/files \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "test.txt", "content_type": "text/plain", "bucket": "default", "data": "SGVsbG8gV29ybGQ="}'
```

预期: 返回 `{id, name, url}`

### 9d: 查询文件列表

```bash
curl -X POST http://localhost:11002/storage/files/query \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

### 9e: 下载文件

```bash
FILE_ID="<id_from_9c>"
curl http://localhost:11002/storage/files/$FILE_ID/download \
  -H "Authorization: Bearer $TOKEN"
```

### 9f: 删除文件

```bash
curl -X DELETE http://localhost:11002/storage/files/$FILE_ID \
  -H "Authorization: Bearer $TOKEN"
```

### 9g: 验证无 token 被拒绝

```bash
curl -X POST http://localhost:11002/storage/files/query \
  -H "Content-Type: application/json" \
  -d '{}'
```

预期: 401 Unauthorized

### 9h: 提交

```bash
git add apps/yggdrasil/
git commit -m "feat(yggdrasil): storage service passes integration verification"
```

---

## 完成标准

- [x] `./nx run yggdrasil_storage:build` 编译通过
- [x] 服务可启动，3 个传输层正常监听
- [x] 文件上传返回 ID 和 URL
- [x] 文件下载返回正确内容
- [x] 文件查询分页正常
- [x] 文件删除正常
- [x] 无 token 请求被 auth 中间件拒绝
- [x] Storage backend 接口可切换（local / minio）
- [x] 变更已提交
