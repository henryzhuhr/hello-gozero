#!/usr/bin/env python3
"""测试配置的 Pydantic 模型定义"""

from pathlib import Path
from typing import Optional

import yaml
from pydantic import BaseModel, Field


# ============ 服务配置 ============
class ServiceConfig(BaseModel):
    """服务配置"""

    base_url: str = Field(..., description="API 服务地址")
    health_check_url: str = Field(..., description="健康检查地址")
    startup_timeout: int = Field(..., gt=0, description="服务启动超时时间(秒)")
    shutdown_timeout: int = Field(..., gt=0, description="服务关闭超时时间(秒)")


# ============ 数据库配置 ============
class DatabaseConfig(BaseModel):
    """数据库配置"""

    host: str = Field(..., description="数据库主机地址")
    port: int = Field(..., gt=0, le=65535, description="数据库端口")
    user: str = Field(..., description="数据库用户名")
    password: str = Field(..., description="数据库密码")
    database: str = Field(..., description="数据库名称")
    charset: str = Field(default="utf8mb4", description="字符集")


# ============ Redis 配置 ============
class RedisConfig(BaseModel):
    """Redis 配置"""

    host: str = Field(..., description="Redis 主机地址")
    port: int = Field(..., gt=0, le=65535, description="Redis 端口")
    db: int = Field(default=0, ge=0, description="Redis 数据库编号")
    password: Optional[str] = Field(default=None, description="Redis 密码")


# ============ 性能测试配置 ============


class RegisterUserPerfConfig(BaseModel):
    """用户注册接口性能配置"""

    response_time_baseline: int = Field(
        ..., gt=0, description="单次注册响应时间基准(ms)"
    )
    concurrent_success_rate: float = Field(
        ..., ge=0.0, le=1.0, description="并发注册成功率基准(百分比)"
    )
    bulk_success_rate: float = Field(
        ..., ge=0.0, le=1.0, description="批量注册成功率基准(百分比)"
    )
    performance_degradation_factor: float = Field(
        ..., gt=0.0, description="性能衰减系数"
    )
    concurrent_users: int = Field(..., gt=0, description="并发测试用户数")
    duplicate_username_concurrent: int = Field(
        ..., gt=0, description="重复用户名并发测试数"
    )
    bulk_count: int = Field(..., gt=0, description="批量注册压力测试用户数")


class GetUserPerfConfig(BaseModel):
    """获取用户接口性能配置"""

    response_time_baseline: int = Field(
        ..., gt=0, description="单次查询响应时间基准(ms)"
    )
    concurrent_count: int = Field(..., gt=0, description="并发查询数")
    hot_query_count: int = Field(..., gt=0, description="热查询次数")
    cache_performance_ratio: float = Field(
        ..., ge=0.0, le=1.0, description="缓存性能提升比例"
    )
    bulk_count: int = Field(..., gt=0, description="批量查询用户数")
    bulk_success_rate: float = Field(
        ..., ge=0.0, le=1.0, description="批量查询成功率基准(百分比)"
    )


class DeleteUserPerfConfig(BaseModel):
    """删除用户接口性能配置"""

    response_time_baseline: int = Field(
        ..., gt=0, description="单次删除响应时间基准(ms)"
    )
    concurrent_count: int = Field(..., gt=0, description="并发删除数")
    bulk_success_rate: float = Field(
        ..., ge=0.0, le=1.0, description="批量删除成功率基准(百分比)"
    )


class UserModulePerfConfig(BaseModel):
    """用户模块性能配置"""

    register_user: RegisterUserPerfConfig
    get_user: GetUserPerfConfig
    delete_user: DeleteUserPerfConfig


class PerformanceConfig(BaseModel):
    """性能测试总配置"""

    user: UserModulePerfConfig


# ============ 总配置 ============
class TestConfig(BaseModel):
    """测试总配置"""

    service: ServiceConfig
    database: DatabaseConfig
    redis: RedisConfig
    performance: PerformanceConfig

    @classmethod
    def load_from_yaml(cls, config_path: Path) -> "TestConfig":
        """从 YAML 文件加载配置并验证"""
        with open(config_path, "r", encoding="utf-8") as f:
            data = yaml.safe_load(f)
        return cls(**data)
