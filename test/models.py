#!/usr/bin/env python3
"""
测试数据模型
定义所有测试用例使用的共享数据结构
"""

from typing import Any, Optional

from pydantic import BaseModel, ConfigDict, Field


class CommonResponse(BaseModel):
    """统一的HTTP响应格式"""

    status_code: int = Field(..., description="HTTP状态码")
    response_time: float = Field(..., description="响应时间（毫秒）")
    data: Optional[Any] = Field(
        None, description="响应数据，可以是字典、字符串或其他类型"
    )

    model_config = ConfigDict(extra="allow")
