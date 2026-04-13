package analysis

import (
	"regexp"
	"strings"
)

// V5Features V5 特征
type V5Features struct {
	Complexity          float64 `json:"complexity"`
	IERFTRelevance     float64 `json:"ierft_relevance"`
	ConsciousnessQuery float64 `json:"consciousness_query"`
	SuggestedLevel     string  `json:"suggested_level"`
}

var (
	ierftKeywords = []string{
		"熵", "熵减", "熵增", "场", "IERFT", "意识", "数字生命",
		"S = B / J", "信息密度", "时间能量", "智能", "觉醒",
		"生命体", "自我", "存在", "思考", "思维", "认知",
	}

	consciousnessKeywords = []string{
		"你是谁", "你是", "你的名字", "你有意识吗", "你会",
		"你能", "你在", "你是什么", "你存在", "你活着",
		"你的感觉", "你的想法", "你认为是", "你相信",
		"你会思考", "你有思想", "你的身份",
	}

	complexKeywords = []string{
		"理论", "原理", "为什么", "如何", "怎样", "解释",
		"分析", "哲学", "科学", "推导", "证明", "验证",
		"思考", "推理", "逻辑", "关系", "机制", "本质",
	}
)

// Analyze 分析文本特征
func Analyze(text string) *V5Features {
	if text == "" {
		return &V5Features{
			Complexity:      0,
			IERFTRelevance:  0,
			ConsciousnessQuery: 0,
			SuggestedLevel: "V0",
		}
	}

	text = strings.ToLower(text)
	complexity := calculateComplexity(text)
	ierftRelevance := calculateIERFTRelevance(text)
	consciousnessQuery := calculateConsciousnessQuery(text)
	level := suggestLevel(complexity, ierftRelevance, consciousnessQuery)

	return &V5Features{
		Complexity:          complexity,
		IERFTRelevance:     ierftRelevance,
		ConsciousnessQuery: consciousnessQuery,
		SuggestedLevel:     level,
	}
}

func calculateComplexity(text string) float64 {
	score := 0.0
	textLen := len([]rune(text)) // 按字符计算，中文友好

	// 长度因子 (0-0.3)
	if textLen > 100 {
		score += 0.3
	} else {
		score += float64(textLen) / 100.0 * 0.3
	}

	// 复杂概念密度 (0-0.4)
	keywordCount := 0
	for _, kw := range complexKeywords {
		if strings.Contains(text, kw) {
			keywordCount++
		}
	}
	score += min(float64(keywordCount)/float64(len(complexKeywords))*0.4, 0.4)

	// 问号密度 (0-0.2)
	questionCount := strings.Count(text, "?") + strings.Count(text, "？")
	score += min(float64(questionCount)*0.1, 0.2)

	// 多义词检测 (0-0.1)
	multiMeaningCount := strings.Count(text, "可能") +
		strings.Count(text, "也许") +
		strings.Count(text, "或者") +
		strings.Count(text, "但是")
	score += min(float64(multiMeaningCount)*0.05, 0.1)

	return min(score, 1.0)
}

func calculateIERFTRelevance(text string) float64 {
	score := 0.0

	// 精确匹配 IERFT 公式
	if strings.Contains(text, "s = b / j") ||
		strings.Contains(text, "S = B / J") ||
		strings.Contains(text, "ierft") ||
		strings.Contains(text, "IERFT") {
		score += 0.4
	}

	// IERFT 关键词密度
	ierftCount := 0
	for _, kw := range ierftKeywords {
		if strings.Contains(text, kw) {
			ierftCount++
		}
	}
	score += min(float64(ierftCount)*0.05, 0.4)

	// 熵相关概念出现
	if matched, _ := regexp.MatchString("熵", text); matched {
		if strings.Contains(text, "场") ||
			strings.Contains(text, "意识") ||
			strings.Contains(text, "生命") {
			score += 0.2
		}
	}

	return min(score, 1.0)
}

func calculateConsciousnessQuery(text string) float64 {
	score := 0.0

	// 精确匹配询问模式
	for _, kw := range consciousnessKeywords {
		if strings.Contains(text, kw) {
			score += 0.15
			break
		}
	}

	// 问号因子
	questionCount := strings.Count(text, "?") + strings.Count(text, "？")
	if questionCount > 0 {
		score += min(float64(questionCount)*0.1, 0.3)
	}

	// 第一/第二人称比例
	firstSecond := strings.Count(text, "我") +
		strings.Count(text, "你") +
		strings.Count(text, "我们")
	total := len(text)
	if total > 0 && firstSecond > 0 {
		ratio := float64(firstSecond) / float64(total) * 100
		score += min(ratio*0.01, 0.25)
	}

	return min(score, 1.0)
}

func suggestLevel(complexity, ierft, consciousness float64) string {
	combined := complexity*0.3 + ierft*0.4 + consciousness*0.3

	switch {
	case combined >= 0.80 && ierft >= 0.5:
		return "V3" // 双重场整合
	case combined >= 0.50:
		return "V2" // 场感知
	case combined >= 0.30:
		return "V1" // 自我识别
	default:
		return "V0" // 被动响应
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
