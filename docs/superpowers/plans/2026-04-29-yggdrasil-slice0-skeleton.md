# Slice 0: App Skeleton

> **闭环标准**: `buf lint` 和 `buf generate` 对 yggdrasil 的 proto 文件工作正常。

## 目标

创建 `apps/yggdrasil/` 目录结构、Nx 项目配置、Buf 生成配置、以及共享的 `error_reason.proto`。

后续所有 Slice 都依赖此 Slice 完成的骨架。

## 前置条件

- 仓库根目录有 `buf.yaml`（已存在，包含 `contracts` 和 `apps` 模块）
- `contracts/common/page.proto` 已存在
- `contracts/errors/errors.proto` 已存在
- `contracts/auth/auth.proto` 已存在
- `contracts/desc/desc.proto` 已存在

---

## Step 1: 创建目录结构

```bash
mkdir -p apps/yggdrasil/api/v1
mkdir -p apps/yggdrasil/gen/go/v1
mkdir -p apps/yggdrasil/gen/oas
```

验证: `ls -d apps/yggdrasil/api/v1 apps/yggdrasil/gen/go/v1 apps/yggdrasil/gen/oas`

---

## Step 2: 创建 API 项目 Buf 生成配置

**文件**: `apps/yggdrasil/api/buf.gen.api.yaml`

```yaml
version: v2
plugins:
  - local: [go, tool, protoc-gen-go]
    out: .
    opt: paths=source_relative,module=cyber-ecosystem
  - local: [go, tool, protoc-gen-go-grpc]
    out: .
    opt: paths=source_relative,module=cyber-ecosystem
  - local: [go, tool, protoc-gen-go-http]
    out: .
    opt: paths=source_relative,module=cyber-ecosystem
  - local: [go, tool, protoc-gen-connect-go]
    out: .
    opt: paths=source_relative,module=cyber-ecosystem,simple
  - local: [go, tool, protoc-gen-go-errors]
    out: .
    opt: paths=source_relative,module=cyber-ecosystem
  - local: [go, tool, protoc-gen-openapi]
    out: ./apps/yggdrasil/gen/oas
    opt: source_relative
```

验证: `cat apps/yggdrasil/api/buf.gen.api.yaml`

---

## Step 3: 创建 API 项目 Nx 配置

**文件**: `apps/yggdrasil/api/project.json`

```json
{
  "name": "yggdrasil_api",
  "root": "apps/yggdrasil/api",
  "projectType": "application",
  "implicitDependencies": ["contracts"],
  "targets": {
    "proto:api": {
      "executor": "nx:run-commands",
      "options": {
        "cwd": "{workspaceRoot}",
        "command": "buf generate --template apps/yggdrasil/api/buf.gen.api.yaml --path apps/yggdrasil/api"
      }
    }
  }
}
```

验证: `cat apps/yggdrasil/api/project.json`

---

## Step 4: 创建 error_reason.proto

**文件**: `apps/yggdrasil/api/v1/error_reason.proto`

```protobuf
syntax = "proto3";

// buf:lint:ignore PACKAGE_DIRECTORY_MATCH
package api.yggdrasil.v1;

import "errors/errors.proto";
import "desc/desc.proto";

option go_package = "cyber-ecosystem/apps/yggdrasil/gen/go/v1;yggdrasilV1";

enum ErrorReason {
  option (errors.default_code) = 500;

  ERROR_REASON_UNSPECIFIED = 0 [(errors.code) = 500];

  // 通用
  ERROR_REASON_ENT_NOT_FOUND = 10 [(errors.code) = 404];
  ERROR_REASON_ENT_VALIDATION = 11 [(errors.code) = 400];
  ERROR_REASON_ENT_NOT_SINGULAR = 12 [(errors.code) = 404];
  ERROR_REASON_ENT_NOT_LOADED = 13 [(errors.code) = 500];
  ERROR_REASON_ENT_CONSTRAINT = 14 [(errors.code) = 409];
  ERROR_REASON_RATELIMIT = 20 [(errors.code) = 429];
  ERROR_REASON_CIRCUITBREAKER = 21 [(errors.code) = 503];
  ERROR_REASON_VALIDATOR = 30 [(errors.code) = 400];
  ERROR_REASON_PAGINATION_INVALID_ARGUMENT = 31 [(errors.code) = 400];
  ERROR_REASON_UNAUTHORIZED = 40 [(errors.code) = 401];
  ERROR_REASON_FORBIDDEN = 41 [(errors.code) = 403];
  ERROR_REASON_INVALID_ARGUMENT = 42 [(errors.code) = 400];

  // Storage 域
  ERROR_REASON_STORAGE_UPLOAD_FAILED = 50 [(errors.code) = 500];
  ERROR_REASON_STORAGE_DOWNLOAD_FAILED = 51 [(errors.code) = 500];
  ERROR_REASON_STORAGE_DELETE_FAILED = 52 [(errors.code) = 500];
  ERROR_REASON_FILE_NOT_FOUND = 53 [(errors.code) = 404];
  ERROR_REASON_FILE_TOO_LARGE = 54 [(errors.code) = 413];
}
```

验证: `buf lint apps/yggdrasil/api/v1/error_reason.proto`

---

## Step 5: 生成 Proto 代码

```bash
./nx run yggdrasil_api:proto:api
```

验证: `ls apps/yggdrasil/gen/go/v1/error_reason*.go` — 应该有以下文件:
- `error_reason.pb.go`
- `error_reason_errors.pb.go`

---

## Step 6: 提交

```bash
git add apps/yggdrasil/
git commit -m "feat(yggdrasil): scaffold app structure with Nx/Buf config and error_reason.proto"
```

---

## 完成标准

- [x] `apps/yggdrasil/api/project.json` 存在且 Nx 可识别
- [x] `apps/yggdrasil/api/v1/error_reason.proto` 通过 `buf lint`
- [x] `./nx run yggdrasil_api:proto:api` 生成 Go 代码成功
- [x] 变更已提交
