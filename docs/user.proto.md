```protobuf
syntax = "proto3";

package api.v1;

import "buf/validate/validate.proto";
import "common/common.proto";
import "google/api/annotations.proto";

option go_package = "github.com/DrReMain/cyber-ecosystem/kratos/system-service/gen/v1;v1";

// UserService 系统用户服务
// 提供用户的创建、更新、删除、查询等核心操作接口
service UserService {
  // CreateUser 创建用户接口
  // 用于新增系统用户，包含用户名、邮箱、年龄、密码等信息的校验
  // @summary 创建系统用户
  // @description 新增系统用户，需满足用户名、邮箱、年龄、密码等校验规则，密码需二次确认
  // @tags 系统用户管理
  // @accept json
  // @produce json
  // @param body body CreateUserRequest true "创建用户请求参数"
  // @success 200 {object} CreateUserResponse "创建成功响应"
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {
    option (google.api.http) = {
      post: "/api/v1/system/user" 
      body: "*"                   
    };
  }

  // UpdateUser 更新用户接口
  // 根据用户ID更新用户信息，支持用户名、邮箱、年龄、密码等字段的修改
  // @summary 更新系统用户
  // @description 根据用户ID更新用户信息，修改的字段需满足对应的校验规则
  // @tags 系统用户管理
  // @accept json
  // @produce json
  // @param id path string true "用户ID"
  // @param body body UpdateUserRequest true "更新用户请求参数"
  // @success 200 {object} UpdateUserResponse "更新成功响应"
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse) {
    option (google.api.http) = {
      put: "/api/v1/system/user/{id}"  
      body: "*"                        
    };
  }

  // DeleteBatchUser 批量删除用户接口
  // 一次性删除多个用户，需传入至少一个用户ID，且ID不能重复
  // @summary 批量删除系统用户
  // @description 批量删除指定ID的系统用户，支持单次删除多个用户
  // @tags 系统用户管理
  // @accept json
  // @produce json
  // @param body body DeleteBatchUserRequest true "批量删除用户请求参数"
  // @success 200 {object} DeleteBatchUserResponse "批量删除成功响应"
  rpc DeleteBatchUser(DeleteBatchUserRequest) returns (DeleteBatchUserResponse) {
    option (google.api.http) = {
      post: "/api/v1/system/user/batch-delete"  
      body: "*"                                 
    };
  }

  // DeleteUser 删除单个用户接口
  // 根据用户ID删除指定的系统用户
  // @summary 删除系统用户
  // @description 根据用户ID删除单个系统用户
  // @tags 系统用户管理
  // @accept json
  // @produce json
  // @param id path string true "用户ID"
  // @success 200 {object} DeleteUserResponse "删除成功响应"
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse) {
    option (google.api.http) = {
      delete: "/api/v1/system/user/{id}"  
    };
  }

  // GetUser 查询单个用户接口
  // 根据用户ID查询用户的详细信息
  // @summary 查询单个系统用户
  // @description 根据用户ID获取用户的完整信息（ID、创建时间、用户名、邮箱、年龄等）
  // @tags 系统用户管理
  // @accept json
  // @produce json
  // @param id path string true "用户ID"
  // @success 200 {object} GetUserResponse "查询成功响应，包含用户详细信息"
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {
    option (google.api.http) = {
      get: "/api/v1/system/user/{id}"  
    };
  }

  // QueryUser 分页查询用户列表接口
  // 支持分页和邮箱模糊查询，返回符合条件的用户列表及分页信息
  // @summary 分页查询系统用户列表
  // @description 分页查询系统用户，支持根据邮箱筛选，返回分页数据和用户列表
  // @tags 系统用户管理
  // @accept json
  // @produce json
  // @param page query api.common.PageRequest true "分页参数"
  // @param email query string false "邮箱筛选条件（可选）"
  // @success 200 {object} QueryUserResponse "分页查询成功响应，包含分页信息和用户列表"
  rpc QueryUser(QueryUserRequest) returns (QueryUserResponse) {
    option (google.api.http) = {
      get: "/api/v1/system/user" 
    };
  }

  // ListUser 查询所有用户接口
  // 返回系统中所有用户的完整列表，无分页和筛选条件
  // @summary 查询所有系统用户
  // @description 获取系统中所有用户的完整列表，包含所有用户的详细信息
  // @tags 系统用户管理
  // @accept json
  // @produce json
  // @success 200 {object} ListUserResponse "查询成功响应，包含所有用户列表"
  rpc ListUser(ListUserRequest) returns (ListUserResponse) {
    option (google.api.http) = {
      get: "/api/v1/system/user-all"
    };
  }
}

// UserEntity 用户实体模型
// 系统用户的核心数据结构，包含用户的基础信息
message UserEntity {
  string id = 1;          // 用户唯一标识ID
  int64 created_at = 2;   // 用户创建时间（时间戳，单位：秒）
  int64 updated_at = 3;   // 用户更新时间（时间戳，单位：秒）
  string username = 4;    // 用户名（唯一）
  string email = 5;       // 用户邮箱（公司内网邮箱）
  int32 age = 6;          // 用户年龄
}

// CreateUserRequest 创建用户请求参数
// 包含创建用户所需的所有字段及校验规则
message CreateUserRequest {
  // 用户名：不能包含"admin"关键词
  string username = 1 [(buf.validate.field).cel = {
    id: "CreateUserRequest.username"
    message: "用户名不能包含 admin"
    expression: "!this.contains('admin')"
  }];
  // 用户邮箱：必须以@cyber.com结尾（公司内网邮箱）
  string email = 2 [(buf.validate.field).cel = {
    id: "CreateUserRequest.email"
    message: "仅限公司内网邮箱注册 (@cyber.com)"
    expression: "this.endsWith('@cyber.com')"
  }];
  // 用户年龄：18 ≤ 年龄 < 150
  int32 age = 3 [(buf.validate.field).cel = {
    id: "CreateUserRequest.age"
    message: "年龄必须大于等于18，小于150"
    expression: "this >= 18 && this < 150"
  }];
  // 用户密码：长度至少8位
  string password = 4 [(buf.validate.field).cel = {
    id: "CreateUserRequest.password"
    message: "密码至少需要8位"
    expression: "size(this) >= 8"
  }];
  // 确认密码：需与密码字段一致
  string confirm_password = 5;  

  option (buf.validate.message).cel = {
    id: "CreateUserRequest.password-equal-confirm_password"
    message: "两次输入的密码不一致"
    expression: "this.password == this.confirm_password"
  };
}

// CreateUserResponse 创建用户响应结果
// 创建用户接口的返回数据结构（当前无返回字段，仅标识操作成功）
message CreateUserResponse {}

// UpdateUserRequest 更新用户请求参数
// 包含更新用户所需的所有字段及校验规则，ID为必传字段
message UpdateUserRequest {
  string id = 1 [(buf.validate.field).string.len = 20];
  
  // 用户名：更新时不能包含"admin"关键词
  string username = 2;
  // 用户邮箱：更新时必须以@cyber.com结尾
  string email = 3;
  // 用户年龄：更新时需满足18 ≤ 年龄 < 150
  int32 age = 4;
  // 用户密码：更新时长度至少8位
  string password = 5;
  // 确认密码：更新时需与密码字段一致
  string confirm_password = 6;  

  option (buf.validate.message).cel = {
    id: "UpdateUserRequest.username"
    message: "用户名不能包含 admin"
    expression: "!this.username.contains('admin')"
  };
  option (buf.validate.message).cel = {
    id: "UpdateUserRequest.email"
    message: "仅限公司内网邮箱注册 (@cyber.com)"
    expression: "this.email.endsWith('@cyber.com')"
  };
  option (buf.validate.message).cel = {
    id: "UpdateUserRequest.age"
    message: "年龄必须大于等于18，小于150"
    expression: "this.age >= 18 && this.age < 150"
  };
  option (buf.validate.message).cel = {
    id: "UpdateUserRequest.password"
    message: "密码至少需要8位"
    expression: "size(this.password) >= 8"
  };
  option (buf.validate.message).cel = {
    id: "UpdateUserRequest.password-equal-confirm_password"
    message: "两次输入的密码不一致"
    expression: "this.password == this.confirm_password"
  };
}

// UpdateUserResponse 更新用户响应结果
// 更新用户接口的返回数据结构（当前无返回字段，仅标识操作成功）
message UpdateUserResponse {}

// DeleteBatchUserRequest 批量删除用户请求参数
// 包含需要删除的用户ID列表，需满足至少1个ID且无重复
message DeleteBatchUserRequest {
  repeated string ids = 1 [(buf.validate.field).repeated = {
    min_items: 1  
    unique: true  
  }];
}

// DeleteBatchUserResponse 批量删除用户响应结果
// 批量删除用户接口的返回数据结构（当前无返回字段，仅标识操作成功）
message DeleteBatchUserResponse {}

// DeleteUserRequest 删除单个用户请求参数
// 包含需要删除的用户ID，固定长度20位
message DeleteUserRequest {
  string id = 1 [(buf.validate.field).string.len = 20];
}

// DeleteUserResponse 删除单个用户响应结果
// 删除单个用户接口的返回数据结构（当前无返回字段，仅标识操作成功）
message DeleteUserResponse {}

// GetUserRequest 查询单个用户请求参数
// 包含需要查询的用户ID，固定长度20位
message GetUserRequest {
  string id = 1 [(buf.validate.field).string.len = 20];
}

// GetUserResponse 查询单个用户响应结果
// 包含查询到的用户详细信息
message GetUserResponse {
  UserEntity user = 1;
}

// QueryUserRequest 分页查询用户列表请求参数
// 包含分页参数和可选的邮箱筛选条件
message QueryUserRequest {
  api.common.PageRequest page = 1;
  // 邮箱筛选条件：可选，用于模糊查询
  optional string email = 2;             
}

// QueryUserResponse 分页查询用户列表响应结果
// 包含分页信息和符合条件的用户列表
message QueryUserResponse {
  api.common.PageResponse page = 1; 
  repeated UserEntity list = 2;          
}

// ListUserRequest 查询所有用户请求参数
// 无请求参数，用于获取全量用户列表
message ListUserRequest {}

// ListUserResponse 查询所有用户响应结果
// 包含系统中所有用户的列表
message ListUserResponse {
  repeated UserEntity list = 1;
}
```
