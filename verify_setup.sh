#!/bin/bash
# 快速验证脚本 - 检查所有必需文件和配置

echo "🔍 检查 pytest 自动化测试环境..."
echo ""

# 检查文件存在性
files=(
    "test/user/test_register_user.py"
    "test/__init__.py"
    "test/user/__init__.py"
    "test/check_service.py"
    "pytest.ini"
    "Makefile"
    "pyproject.toml"
)

echo "📁 检查必需文件..."
all_files_exist=true
for file in "${files[@]}"; do
    if [ -f "$file" ]; then
        echo "  ✓ $file"
    else
        echo "  ✗ $file (缺失)"
        all_files_exist=false
    fi
done
echo ""

# 检查 Python 环境
echo "🐍 检查 Python 环境..."
if command -v uv &> /dev/null; then
    echo "  ✓ uv 已安装"
    
    # 检查 pytest
    if uv run python -c "import pytest; print(f'  ✓ pytest {pytest.__version__} 已安装')" 2>/dev/null; then
        :
    else
        echo "  ✗ pytest 未安装，运行: uv sync"
    fi
else
    echo "  ✗ uv 未安装"
fi
echo ""

# 检查 Go 环境
echo "🔧 检查 Go 环境..."
if command -v go &> /dev/null; then
    echo "  ✓ Go $(go version | awk '{print $3}') 已安装"
else
    echo "  ✗ Go 未安装"
fi
echo ""

# 检查 Docker
echo "🐳 检查 Docker 环境..."
if command -v docker-compose &> /dev/null || command -v docker &> /dev/null; then
    echo "  ✓ Docker 可用"
else
    echo "  ⚠ Docker 未安装（测试需要 MySQL/Redis）"
fi
echo ""

# 测试发现
echo "🧪 测试发现..."
if uv run pytest --collect-only -q test/user/test_register_user.py 2>/dev/null | grep -q "4 tests"; then
    echo "  ✓ 发现 4 个测试用例"
else
    echo "  ⚠ 测试发现异常"
fi
echo ""

# 语法检查
echo "✅ Python 语法检查..."
if uv run python -m py_compile test/user/test_register_user.py 2>/dev/null; then
    echo "  ✓ 测试文件语法正确"
else
    echo "  ✗ 测试文件有语法错误"
fi
echo ""

# 总结
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ "$all_files_exist" = true ]; then
    echo "✨ 所有必需文件已就绪!"
    echo ""
    echo "📖 快速开始:"
    echo "  1. 启动 Docker: make docker-up"
    echo "  2. 运行测试: make test"
    echo "  或直接: pytest"
    echo ""
    echo "📚 更多信息:"
    echo "  - TESTING_SUMMARY.md  (完整总结)"
    echo "  - QUICKSTART.md       (快速指南)"
    echo "  - test/README.md      (测试文档)"
else
    echo "⚠ 部分文件缺失,请检查"
fi
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
