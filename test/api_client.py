#!/usr/bin/env python3
"""API 客户端模块

提供统一的 HTTP 请求封装，处理公共逻辑如：
- 基础 URL 配置
- 请求头配置
- 响应统一处理
- 异常统一处理
- 日志记录
"""

from typing import Any, Dict, Optional

import requests
from loguru import logger

from test.models import CommonResponse


class ApiClient:
    """API 客户端，封装 HTTP 请求的公共逻辑
    
    使用 requests.Session() 维护连接池，提高性能并正确管理连接生命周期
    """

    def __init__(self, base_url: str = "http://localhost:8888", timeout: int = 5):
        """初始化 API 客户端

        Args:
            base_url: API 基础 URL
            timeout: 请求超时时间（秒）
        """
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout
        self.default_headers = {"Content-Type": "application/json"}
        self.session = requests.Session()  # 使用 Session 复用连接池
        self.session.headers.update(self.default_headers)

    def _make_request(
        self,
        method: str,
        url: str,
        data: Optional[Dict[str, Any]] = None,
        headers: Optional[Dict[str, str]] = None,
    ) -> CommonResponse:
        """发送 HTTP 请求的通用方法

        Args:
            method: HTTP 方法（GET/POST/PUT/DELETE等）
            url: 相对 URL 路径（如 /api/v1/users）
            data: 请求体数据（POST/PUT 时使用）
            headers: 额外的请求头（会与默认头合并）

        Returns:
            CommonResponse: 统一的响应对象
        """
        # 构建完整 URL
        full_url = f"{self.base_url}{url}"

        # 合并请求头
        request_headers = self.default_headers.copy()
        if headers:
            request_headers.update(headers)

        try:
            # 使用 Session 发送请求（复用连接池）
            response = self.session.request(
                method=method.upper(),
                url=full_url,
                json=data,
                headers=request_headers,
                timeout=self.timeout,
            )
        except requests.RequestException as e:
            logger.error(f"请求异常 [{method.upper()} {url}]: {str(e)}")
            return CommonResponse(status_code=0, response_time=0.0, data={})

        # 构建响应对象
        result = CommonResponse(
            status_code=response.status_code,
            response_time=response.elapsed.total_seconds() * 1000.0,  # 转换为毫秒
            data=None,
        )

        # 解析响应数据
        try:
            # 检查 Content-Type 是否包含 json（兼容 charset）
            content_type = response.headers.get("Content-Type", "").lower()
            if "application/json" in content_type:
                result.data = response.json()
            else:
                result.data = response.text
        except Exception:
            result.data = response.text

        # 记录日志
        logger.info(
            f"{method.upper()} {url} - status:{result.status_code}, "
            f"time:{result.response_time:.3f}ms"
        )

        return result

    def get(self, url: str, headers: Optional[Dict[str, str]] = None) -> CommonResponse:
        """发送 GET 请求

        Args:
            url: 相对 URL 路径
            headers: 额外的请求头

        Returns:
            CommonResponse: 响应对象

        Examples:
            >>> client = ApiClient()
            >>> response = client.get("/api/v1/users/testuser")
            >>> assert response.status_code == 200
        """
        return self._make_request("GET", url, headers=headers)

    def post(
        self,
        url: str,
        data: Optional[Dict[str, Any]] = None,
        headers: Optional[Dict[str, str]] = None,
    ) -> CommonResponse:
        """发送 POST 请求

        Args:
            url: 相对 URL 路径
            data: 请求体数据
            headers: 额外的请求头

        Returns:
            CommonResponse: 响应对象

        Examples:
            >>> client = ApiClient()
            >>> response = client.post("/api/v1/users/register", data=user_data)
            >>> assert response.status_code == 200
        """
        return self._make_request("POST", url, data=data, headers=headers)

    def put(
        self,
        url: str,
        data: Optional[Dict[str, Any]] = None,
        headers: Optional[Dict[str, str]] = None,
    ) -> CommonResponse:
        """发送 PUT 请求

        Args:
            url: 相对 URL 路径
            data: 请求体数据
            headers: 额外的请求头

        Returns:
            CommonResponse: 响应对象
        """
        return self._make_request("PUT", url, data=data, headers=headers)

    def delete(
        self, url: str, headers: Optional[Dict[str, str]] = None
    ) -> CommonResponse:
        """发送 DELETE 请求

        Args:
            url: 相对 URL 路径
            headers: 额外的请求头

        Returns:
            CommonResponse: 响应对象

        Examples:
            >>> client = ApiClient()
            >>> response = client.delete("/api/v1/users/testuser")
            >>> assert response.status_code == 200
        """
        return self._make_request("DELETE", url, headers=headers)

    def patch(
        self,
        url: str,
        data: Optional[Dict[str, Any]] = None,
        headers: Optional[Dict[str, str]] = None,
    ) -> CommonResponse:
        """发送 PATCH 请求

        Args:
            url: 相对 URL 路径
            data: 请求体数据
            headers: 额外的请求头

        Returns:
            CommonResponse: 响应对象
        """
        return self._make_request("PATCH", url, data=data, headers=headers)

    def close(self):
        """关闭 Session，释放连接池资源
        
        在测试结束后应该调用此方法，避免连接泄漏
        """
        if self.session:
            self.session.close()
            logger.debug("ApiClient Session 已关闭")

    def __enter__(self):
        """支持上下文管理器"""
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """退出上下文时自动关闭连接"""
        self.close()
