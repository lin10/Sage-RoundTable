package utils

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

// ParseMarkdownConfig 解析带有 Frontmatter 的 Markdown 文件
// 它将 --- 包裹的 YAML 部分反序列化到 out 中，并返回去除 Frontmatter 后的纯文本 Markdown 正文
func ParseMarkdownConfig(content []byte, out interface{}) (string, error) {
	// 确保内容有前导的 ---
	if !bytes.HasPrefix(content, []byte("---\n")) && !bytes.HasPrefix(content, []byte("---\r\n")) {
		return "", fmt.Errorf("invalid markdown config: missing frontmatter start '---'")
	}

	// 找到第二个 ---
	// 跳过开头的 ---
	startIndex := bytes.IndexByte(content, '\n')
	if startIndex == -1 {
		return "", fmt.Errorf("invalid markdown config: missing frontmatter content")
	}
	startIndex++

	// 查找结束的 ---
	endIndex := bytes.Index(content[startIndex:], []byte("\n---"))
	if endIndex == -1 {
		return "", fmt.Errorf("invalid markdown config: missing frontmatter end '---'")
	}

	// 提取 YAML 内容
	yamlContent := content[startIndex : startIndex+endIndex]

	// 提取正文内容 (跳过 \n---\n)
	bodyStart := startIndex + endIndex + 4
	if bodyStart < len(content) && (content[bodyStart] == '\n' || content[bodyStart] == '\r') {
		if content[bodyStart] == '\r' && bodyStart+1 < len(content) && content[bodyStart+1] == '\n' {
			bodyStart += 2
		} else {
			bodyStart++
		}
	}

	var body string
	if bodyStart < len(content) {
		body = string(bytes.TrimSpace(content[bodyStart:]))
	}

	// 反序列化 YAML
	if err := yaml.Unmarshal(yamlContent, out); err != nil {
		return "", fmt.Errorf("failed to unmarshal frontmatter yaml: %w", err)
	}

	return body, nil
}
