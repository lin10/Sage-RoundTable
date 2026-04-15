package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LoadEnvFile 加载 .env 文件中的环境变量
// 支持 KEY=VALUE 格式，忽略注释和空行
func LoadEnvFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open env file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 解析 KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			fmt.Printf("Warning: skipping invalid line %d in env file: %s\n", lineNum, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// 移除值两侧的引号（如果有）
		value = strings.Trim(value, `"'`)

		// 只有当环境变量未设置时才设置
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading env file: %w", err)
	}

	return nil
}

// LoadDefaultEnvFile 尝试加载项目根目录的 .env 文件
// 如果文件不存在则返回 nil（不报错）
func LoadDefaultEnvFile() error {
	envPath := ".env"
	
	// 检查文件是否存在
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return nil // 文件不存在，静默忽略
	}

	return LoadEnvFile(envPath)
}
