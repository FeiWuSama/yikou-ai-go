#!/bin/bash

# 进入前端目录
cd yikou-ai-feiwu-front

# 安装依赖
npm install

# 构建前端
npm run build

# 检查构建是否成功
if [ $? -eq 0 ]; then
    echo "前端构建成功"
else
    echo "前端构建失败"
    exit 1
fi
