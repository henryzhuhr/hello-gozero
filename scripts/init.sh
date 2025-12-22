#!/bin/bash


# Install Go tools
export GOPROXY=https://goproxy.cn,direct

# go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
# for Go latest
go install golang.org/x/tools/gopls@latest
go install github.com/cweill/gotests/...@latest
go install github.com/fatih/gomodifytags@latest
go install github.com/josharian/impl@latest
go install github.com/haya14busa/goplay/cmd/goplay@latest
go install github.com/go-delve/delve/cmd/dlv@latest
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest


# go-zero toolkit
go install github.com/zeromicro/go-zero/tools/goctl@latest
go mod tidy

# Init Python environment 
uv sync #--active


echo ": $(date +%s):0;uv run debug/navs/test_navs_v3.py" >> "$HOME"/.zsh_history
echo ": $(date +%s):0;go run app/portal/service/api/portal.go" >> "$HOME"/.zsh_history
