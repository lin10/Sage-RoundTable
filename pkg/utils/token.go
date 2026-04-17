package utils

// EstimateTokens 粗略估算文本的 Token 数量
// 在没有引入复杂 BPE 分词器的情况下，一个简单经验法则是：
// 英文单词约 0.75 个 token，中文字符大约 1~1.5 个 token。
// 为了简单和安全起见，这里按 1 个中文字符=1.5 Token，1 个英文单词=1 Token 粗略估算。
func EstimateTokens(text string) int {
	var tokens float64
	for _, runeValue := range text {
		// 简单区分：如果字符在 ASCII 范围内，计为 0.25 个 Token（相当于 4 个字母一个词）
		if runeValue < 128 {
			tokens += 0.25
		} else {
			// 非 ASCII 字符（如中文），保守估计每个计为 1.5 个 Token
			tokens += 1.5
		}
	}
	return int(tokens)
}

// 提取实际需要的精确计算可能需要 tiktoken-go 等库，这里提供轻量级方案
