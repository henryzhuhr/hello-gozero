"""
数据库辅助模块
提供测试中使用的数据库连接和清理功能
"""

from typing import Optional

import pymysql
from loguru import logger
from pymysql.connections import Connection


class DatabaseHelper:
    """数据库辅助类，用于测试中的数据库操作"""

    def __init__(
        self,
        host: str,
        port: int,
        user: str,
        password: str,
        database: str,
        charset: str = "utf8mb4",
    ):
        """
        初始化数据库连接配置

        Args:
            host: 数据库主机
            port: 数据库端口
            user: 数据库用户名
            password: 数据库密码
            database: 数据库名称
            charset: 字符集，默认 utf8mb4
        """
        self.config = {
            "host": host,
            "port": port,
            "user": user,
            "password": password,
            "database": database,
            "charset": charset,
            "autocommit": True,  # 自动提交事务
        }
        self._connection: Optional[Connection] = None

    def connect(self) -> Connection:
        """
        创建数据库连接

        Returns:
            数据库连接对象
        """
        if self._connection is None or not self._connection.open:
            self._connection = pymysql.connect(**self.config)
            logger.debug(
                f"数据库连接成功: {self.config['user']}@{self.config['host']}:{self.config['port']}/{self.config['database']}"
            )
        return self._connection

    def close(self):
        """关闭数据库连接"""
        if self._connection is not None and self._connection.open:
            self._connection.close()
            logger.debug("数据库连接已关闭")
            self._connection = None

    def execute(self, sql: str, params=None) -> int:
        """
        执行 SQL 语句

        Args:
            sql: SQL 语句
            params: SQL 参数（可选）

        Returns:
            影响的行数
        """
        conn = self.connect()
        with conn.cursor() as cursor:
            affected_rows = cursor.execute(sql, params)
            conn.commit()
            return affected_rows

    def execute_many(self, sql: str, params_list) -> int:
        """
        批量执行 SQL 语句

        Args:
            sql: SQL 语句
            params_list: SQL 参数列表

        Returns:
            影响的行数
        """
        conn = self.connect()
        with conn.cursor() as cursor:
            affected_rows = cursor.executemany(sql, params_list)
            conn.commit()
            return affected_rows

    def query(self, sql: str, params=None):
        """
        查询数据

        Args:
            sql: SQL 语句
            params: SQL 参数（可选）

        Returns:
            查询结果列表
        """
        conn = self.connect()
        with conn.cursor(pymysql.cursors.DictCursor) as cursor:
            cursor.execute(sql, params)
            return cursor.fetchall()

    def query_one(self, sql: str, params=None):
        """
        查询单条数据

        Args:
            sql: SQL 语句
            params: SQL 参数（可选）

        Returns:
            查询结果（单条记录）
        """
        conn = self.connect()
        with conn.cursor(pymysql.cursors.DictCursor) as cursor:
            cursor.execute(sql, params)
            return cursor.fetchone()

    def truncate_table(self, table_name: str):
        """
        清空指定表的数据（保留表结构）

        Args:
            table_name: 表名
        """
        try:
            # 先禁用外键检查
            self.execute("SET FOREIGN_KEY_CHECKS = 0")
            # 清空表
            self.execute(f"TRUNCATE TABLE {table_name}")
            # 恢复外键检查
            self.execute("SET FOREIGN_KEY_CHECKS = 1")
            logger.info(f"表 {table_name} 已清空")
        except Exception as e:
            logger.error(f"清空表 {table_name} 失败: {e}")
            raise

    def delete_all(self, table_name: str, where: Optional[str] = None):
        """
        删除表中的所有数据（软删除兼容）

        Args:
            table_name: 表名
            where: WHERE 条件（可选），例如 "deleted_at IS NULL"
        """
        try:
            sql = f"DELETE FROM {table_name}"
            if where:
                sql += f" WHERE {where}"
            affected = self.execute(sql)
            logger.info(f"表 {table_name} 删除了 {affected} 条记录")
            return affected
        except Exception as e:
            logger.error(f"删除表 {table_name} 数据失败: {e}")
            raise

    def truncate_all_tables(self, exclude_tables: Optional[list] = None):
        """
        清空数据库中所有表的数据

        Args:
            exclude_tables: 要排除的表名列表（可选）
        """
        exclude_tables = exclude_tables or []

        # 查询所有表名
        tables = self.query(
            f"SELECT table_name FROM information_schema.tables "
            f"WHERE table_schema = %s AND table_type = 'BASE TABLE'",
            (self.config["database"],),
        )

        # 禁用外键检查
        self.execute("SET FOREIGN_KEY_CHECKS = 0")

        try:
            for table_info in tables:
                table_name = table_info["table_name"]
                if table_name not in exclude_tables:
                    self.execute(f"TRUNCATE TABLE {table_name}")
                    logger.debug(f"已清空表: {table_name}")
        finally:
            # 恢复外键检查
            self.execute("SET FOREIGN_KEY_CHECKS = 1")

        logger.success(f"已清空数据库 {self.config['database']} 的所有表")

    def __enter__(self):
        """支持上下文管理器（with 语句）"""
        self.connect()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """上下文管理器退出时关闭连接"""
        self.close()

    def __del__(self):
        """析构函数，确保连接关闭"""
        self.close()
