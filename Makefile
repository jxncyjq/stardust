# 设置 Go 命令
GOCMD=go
SWAG=swag init
# 设置 Go 构建命令
GOBUILD=$(GOCMD) build
# 设置可执行文件输出的名称
BINARY_NAME=ad_server

SOURCE_DIR=.

# 默认的 make 命令目标
all: linux

# 构建 Linux 可执行文件的目标
linux:
	@echo "Building for Linux..."
	$(SWAG)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-linux-amd64 $(SOURCE_DIR)

# 构建 Windows 可执行文件的目标
windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-windows-amd64.exe $(SOURCE_DIR)

# 清理构建文件的目标
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)-linux-amd64
	rm -f $(BINARY_NAME)-windows-amd64.exe

# 这里的.PHONY 表示这些目标都是“伪目标”
.PHONY: all linux windows clean
