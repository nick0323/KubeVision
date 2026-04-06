package main

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// 密码生成常量
const (
	DefaultPasswordLength = 16
	MinPasswordLength     = 8
	MaxPasswordLength     = 64
	BcryptCost            = 12 // 推荐值 10-14，生产环境建议 12-14
)

// 密码字符集
const (
	UppercaseChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	LowercaseChars = "abcdefghijklmnopqrstuvwxyz"
	DigitChars     = "0123456789"
	SpecialChars   = "!@#$%^&*()-_=+[]{};:,.<>?"
	AllChars       = UppercaseChars + LowercaseChars + DigitChars + SpecialChars
)

// 密码强度等级
type PasswordStrength int

const (
	StrengthVeryWeak PasswordStrength = iota
	StrengthWeak
	StrengthMedium
	StrengthStrong
	StrengthVeryStrong
)

// 密码强度描述
var strengthDescriptions = map[PasswordStrength]string{
	StrengthVeryWeak:   "非常弱 (极易被破解)",
	StrengthWeak:       "弱 (容易被破解)",
	StrengthMedium:     "中等 (有一定防护能力)",
	StrengthStrong:     "强 (推荐用于生产环境)",
	StrengthVeryStrong: "非常强 (高安全性场景)",
}

// 输出消息常量
const (
	UsageMessage = `K8sVision 密码管理工具 - 生产级最佳实践

使用方法:
  generate_password [命令] [参数]

命令:
  hash <密码>              生成指定密码的 bcrypt 哈希
  generate [长度]          生成随机密码 (默认 16 位)
  check <密码>             检查密码强度
  verify <密码> <哈希>     验证密码是否匹配哈希
  env                      生成环境变量设置命令

示例:
  generate_password hash "MyP@ssw0rd"
  generate_password generate 20
  generate_password check "MyP@ssw0rd123"
  generate_password verify "MyP@ssw0rd" "$2a$12$..."
  generate_password env

安全建议:
  - 生产环境密码长度至少 16 位
  - 包含大小写字母、数字和特殊字符
  - 定期更换密码
  - 使用环境变量或密钥管理系统存储密码`
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println(UsageMessage)
		os.Exit(0)
	}

	command := os.Args[1]

	switch command {
	case "generate":
		length := DefaultPasswordLength
		if len(os.Args) > 2 {
			if l, err := strconv.Atoi(os.Args[2]); err == nil && l > 0 {
				length = l
			}
		}
		generateSecurePassword(length)

	case "hash":
		if len(os.Args) < 3 {
			fmt.Println("错误：请提供要哈希的密码")
			fmt.Println("用法：generate_password hash <密码>")
			os.Exit(1)
		}
		hashPassword(os.Args[2])

	case "check":
		if len(os.Args) < 3 {
			fmt.Println("错误：请提供要检查的密码")
			fmt.Println("用法：generate_password check <密码>")
			os.Exit(1)
		}
		checkPasswordStrength(os.Args[2])

	case "verify":
		if len(os.Args) < 4 {
			fmt.Println("错误：请提供密码和哈希值")
			fmt.Println("用法：generate_password verify <密码> <哈希>")
			os.Exit(1)
		}
		verifyPassword(os.Args[2], os.Args[3])

	case "env":
		generateEnvCommands()

	default:
		fmt.Println(UsageMessage)
		os.Exit(1)
	}
}

// generateSecurePassword 生成安全的随机密码
func generateSecurePassword(length int) {
	if length < MinPasswordLength {
		fmt.Printf("警告：密码长度 %d 小于推荐的最小值 %d，已调整为 %d\n\n", length, MinPasswordLength, MinPasswordLength)
		length = MinPasswordLength
	}
	if length > MaxPasswordLength {
		fmt.Printf("警告：密码长度 %d 超过最大值 %d，已调整为 %d\n\n", length, MaxPasswordLength, MaxPasswordLength)
		length = MaxPasswordLength
	}

	// 确保密码包含各类字符（生产级要求）
	password := make([]byte, 0, length)

	// 至少包含一个大写字母
	if char, err := randomChar(UppercaseChars); err == nil {
		password = append(password, char)
	}

	// 至少包含一个小写字母
	if char, err := randomChar(LowercaseChars); err == nil {
		password = append(password, char)
	}

	// 至少包含一个数字
	if char, err := randomChar(DigitChars); err == nil {
		password = append(password, char)
	}

	// 至少包含一个特殊字符
	if char, err := randomChar(SpecialChars); err == nil {
		password = append(password, char)
	}

	// 填充剩余长度
	for len(password) < length {
		if char, err := randomChar(AllChars); err == nil {
			password = append(password, char)
		}
	}

	// 打乱密码顺序
	shuffledPassword := shuffleBytes(password)

	passwordStr := string(shuffledPassword)

	fmt.Println("=== 安全密码生成结果 ===")
	fmt.Println()
	fmt.Printf("生成的密码：%s\n", passwordStr)
	fmt.Printf("密码长度：%d\n", length)
	fmt.Printf("密码熵值：%.2f bits\n", calculateEntropy(passwordStr))
	fmt.Println()

	// 检查密码强度
	strength := checkPasswordStrengthInternal(passwordStr)
	fmt.Printf("密码强度：%s\n", getStrengthLabel(strength))
	fmt.Printf("强度说明：%s\n", strengthDescriptions[strength])
	fmt.Println()

	// 生成哈希
	generateHashOutput(passwordStr)

	// 输出安全建议
	printSecurityTips()
}

// randomChar 从字符集中随机选择一个字符
func randomChar(charset string) (byte, error) {
	if len(charset) == 0 {
		return 0, fmt.Errorf("字符集为空")
	}

	max := big.NewInt(int64(len(charset)))
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, fmt.Errorf("生成随机数失败：%w", err)
	}

	return charset[n.Int64()], nil
}

// shuffleBytes 使用 Fisher-Yates 算法打乱字节切片
func shuffleBytes(data []byte) []byte {
	result := make([]byte, len(data))
	copy(result, data)

	for i := len(result) - 1; i > 0; i-- {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		j := int(n.Int64())
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// calculateEntropy 计算密码熵值（bits）
func calculateEntropy(password string) float64 {
	if len(password) == 0 {
		return 0
	}

	// 计算字符集大小
	charsetSize := 0
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, r := range password {
		if unicode.IsUpper(r) && !hasUpper {
			hasUpper = true
			charsetSize += len(UppercaseChars)
		}
		if unicode.IsLower(r) && !hasLower {
			hasLower = true
			charsetSize += len(LowercaseChars)
		}
		if unicode.IsDigit(r) && !hasDigit {
			hasDigit = true
			charsetSize += len(DigitChars)
		}
		if !hasSpecial && strings.ContainsRune(SpecialChars, r) {
			hasSpecial = true
			charsetSize += len(SpecialChars)
		}
	}

	if charsetSize == 0 {
		return 0
	}

	// 熵 = 长度 * log2(字符集大小)
	entropy := float64(len(password)) * math.Log2(float64(charsetSize))
	return entropy
}

// checkPasswordStrength 检查密码强度并输出详细报告
func checkPasswordStrength(password string) {
	strength := checkPasswordStrengthInternal(password)

	fmt.Println("=== 密码强度分析报告 ===")
	fmt.Println()
	fmt.Printf("密码：%s\n", maskPassword(password))
	fmt.Printf("长度：%d\n", len(password))
	fmt.Println()

	// 详细检查项
	fmt.Println("检查项:")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// 长度检查
	lengthStatus := "✓"
	if len(password) < 8 {
		lengthStatus = "✗"
	}
	fmt.Fprintf(w, "  %s 长度至少 8 位\n", lengthStatus)

	// 大写字母检查
	hasUpper := false
	for _, r := range password {
		if unicode.IsUpper(r) {
			hasUpper = true
			break
		}
	}
	upperStatus := "✓"
	if !hasUpper {
		upperStatus = "✗"
	}
	fmt.Fprintf(w, "  %s 包含大写字母 (A-Z)\n", upperStatus)

	// 小写字母检查
	hasLower := false
	for _, r := range password {
		if unicode.IsLower(r) {
			hasLower = true
			break
		}
	}
	lowerStatus := "✓"
	if !hasLower {
		lowerStatus = "✗"
	}
	fmt.Fprintf(w, "  %s 包含小写字母 (a-z)\n", lowerStatus)

	// 数字检查
	hasDigit := false
	for _, r := range password {
		if unicode.IsDigit(r) {
			hasDigit = true
			break
		}
	}
	digitStatus := "✓"
	if !hasDigit {
		digitStatus = "✗"
	}
	fmt.Fprintf(w, "  %s 包含数字 (0-9)\n", digitStatus)

	// 特殊字符检查
	hasSpecial := false
	for _, r := range password {
		if strings.ContainsRune(SpecialChars, r) {
			hasSpecial = true
			break
		}
	}
	specialStatus := "✓"
	if !hasSpecial {
		specialStatus = "✗"
	}
	fmt.Fprintf(w, "  %s 包含特殊字符 (!@#$...)\n", specialStatus)
	w.Flush()

	fmt.Println()
	fmt.Printf("综合强度：%s\n", getStrengthLabel(strength))
	fmt.Printf("强度说明：%s\n", strengthDescriptions[strength])
	fmt.Println()

	// 改进建议
	if strength < StrengthStrong {
		fmt.Println("改进建议:")
		if len(password) < 12 {
			fmt.Println("  • 增加密码长度至 12 位以上")
		}
		if !hasUpper || !hasLower {
			fmt.Println("  • 同时使用大写和小写字母")
		}
		if !hasDigit {
			fmt.Println("  • 添加数字")
		}
		if !hasSpecial {
			fmt.Println("  • 添加特殊字符 (!@#$%^&* 等)")
		}
		fmt.Println()
	}

	// 生成哈希选项
	fmt.Println("需要生成此密码的哈希吗？")
	fmt.Println("运行：generate_password hash \"" + password + "\"")
}

// checkPasswordStrengthInternal 内部密码强度检查
func checkPasswordStrengthInternal(password string) PasswordStrength {
	score := 0

	// 长度评分
	if len(password) >= 8 {
		score++
	}
	if len(password) >= 12 {
		score++
	}
	if len(password) >= 16 {
		score++
	}
	if len(password) >= 20 {
		score++
	}

	// 字符多样性评分
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, r := range password {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
		if strings.ContainsRune(SpecialChars, r) {
			hasSpecial = true
		}
	}

	charTypes := 0
	if hasUpper {
		charTypes++
	}
	if hasLower {
		charTypes++
	}
	if hasDigit {
		charTypes++
	}
	if hasSpecial {
		charTypes++
	}

	score += charTypes

	// 检查常见弱密码模式
	if isCommonPattern(password) {
		score = max(0, score-2)
	}

	// 根据分数评定强度
	switch {
	case score >= 7:
		return StrengthVeryStrong
	case score >= 6:
		return StrengthStrong
	case score >= 4:
		return StrengthMedium
	case score >= 2:
		return StrengthWeak
	default:
		return StrengthVeryWeak
	}
}

// isCommonPattern 检查是否为常见弱密码模式
func isCommonPattern(password string) bool {
	lowerPwd := strings.ToLower(password)

	// 常见弱密码模式
	commonPatterns := []string{
		"password", "admin", "123456", "qwerty",
		"abc123", "111111", "123123",
		"iloveyou", "sunshine", "princess",
		"welcome", "monkey", "dragon",
		"k8s", "kubernetes", "k8svision",
	}

	for _, pattern := range commonPatterns {
		if strings.Contains(lowerPwd, pattern) {
			return true
		}
	}

	// 检查连续数字
	if matched, _ := regexp.MatchString(`(012|123|234|345|456|567|678|789)`, password); matched {
		return true
	}

	// 检查连续字母
	if matched, _ := regexp.MatchString(`(abc|bcd|cde|def|efg|fgh|ghi|hij|ijk|jkl|klm|lmn|mno|nop|opq|pqr|qrs|rst|stu|tuv|uvw|vwx|wxy|xyz)`, lowerPwd); matched {
		return true
	}

	return false
}

// hashPassword 生成密码哈希
func hashPassword(password string) {
	// 验证密码强度
	strength := checkPasswordStrengthInternal(password)
	if strength < StrengthMedium {
		fmt.Println("⚠️  警告：密码强度较弱，建议更换强密码")
		fmt.Println()
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		fmt.Printf("错误：密码哈希失败：%v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== 密码哈希生成结果 ===")
	fmt.Println()
	fmt.Printf("原始密码：%s\n", password)
	fmt.Printf("哈希密码：%s\n", string(hashedBytes))
	fmt.Printf("bcrypt cost：%d\n", BcryptCost)
	fmt.Println()

	// 配置方法
	fmt.Println("配置方法:")
	fmt.Println()
	fmt.Println("方法 1: 修改 config.yaml")
	fmt.Println("  auth:")
	fmt.Println("    username: \"admin\"")
	fmt.Printf("    password: \"%s\"\n", string(hashedBytes))
	fmt.Println()
	fmt.Println("方法 2: 使用环境变量")
	fmt.Printf("  export K8SVISION_AUTH_USERNAME=\"admin\"\n")
	fmt.Printf("  export K8SVISION_AUTH_PASSWORD='%s'\n", string(hashedBytes))
	fmt.Println()

	// 验证方法
	fmt.Println("验证哈希:")
	fmt.Printf("  generate_password verify \"%s\" \"%s\"\n", password, string(hashedBytes))
}

// verifyPassword 验证密码是否匹配哈希
func verifyPassword(password, hashedPassword string) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

	fmt.Println("=== 密码验证结果 ===")
	fmt.Println()

	if err == nil {
		fmt.Println("✓ 密码匹配")
		fmt.Println()
		fmt.Println("此哈希可用于配置文件或环境变量")
	} else {
		fmt.Println("✗ 密码不匹配")
		fmt.Printf("错误：%v\n", err)
		os.Exit(1)
	}
}

// generateEnvCommands 生成环境变量设置命令
func generateEnvCommands() {
	fmt.Println("=== 环境变量设置命令 ===")
	fmt.Println()
	fmt.Println("# 生成新密码并设置环境变量")
	fmt.Println("export K8SVISION_AUTH_USERNAME=\"admin\"")
	fmt.Println("export K8SVISION_AUTH_PASSWORD='$(generate_password generate 16 | grep \"生成的密码\" | cut -d\" \" -f3)'")
	fmt.Println()
	fmt.Println("# 或者手动设置")
	fmt.Println("export K8SVISION_AUTH_USERNAME=\"admin\"")
	fmt.Println("export K8SVISION_AUTH_PASSWORD=\"<你的密码哈希>\"")
	fmt.Println()
	fmt.Println("# Windows PowerShell")
	fmt.Println("$env:K8SVISION_AUTH_USERNAME=\"admin\"")
	fmt.Println("$env:K8SVISION_AUTH_PASSWORD=\"<你的密码哈希>\"")
	fmt.Println()
	fmt.Println("# Windows CMD")
	fmt.Println("set K8SVISION_AUTH_USERNAME=admin")
	fmt.Println("set K8SVISION_AUTH_PASSWORD=<你的密码哈希>")
}

// generateHashOutput 生成哈希输出（用于 generate 命令）
func generateHashOutput(password string) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		fmt.Printf("警告：密码哈希失败：%v\n", err)
		return
	}

	fmt.Println("=== 密码哈希（可直接使用） ===")
	fmt.Println()
	fmt.Printf("bcrypt 哈希：%s\n", string(hashedBytes))
	fmt.Println()

	fmt.Println("配置方法:")
	fmt.Println()
	fmt.Println("方法 1: 修改 config.yaml")
	fmt.Println("  auth:")
	fmt.Println("    username: \"admin\"")
	fmt.Printf("    password: \"%s\"\n", string(hashedBytes))
	fmt.Println()
	fmt.Println("方法 2: 使用环境变量")
	fmt.Printf("  export K8SVISION_AUTH_PASSWORD='%s'\n", string(hashedBytes))
}

// printSecurityTips 打印安全提示
func printSecurityTips() {
	fmt.Println("=== 安全提示 ===")
	fmt.Println()
	fmt.Println("✓ 请安全保存此密码（建议使用密码管理器）")
	fmt.Println("✓ 不要将密码提交到版本控制系统")
	fmt.Println("✓ 生产环境建议使用至少 16 位密码")
	fmt.Println("✓ 定期更换密码（建议每 90 天）")
	fmt.Println("✓ 使用环境变量或密钥管理系统存储密码哈希")
	fmt.Println()
	fmt.Println("相关命令:")
	fmt.Println("  generate_password verify <密码> <哈希>  # 验证密码")
	fmt.Println("  generate_password check <密码>         # 检查密码强度")
}

// getStrengthLabel 获取强度标签
func getStrengthLabel(strength PasswordStrength) string {
	labels := map[PasswordStrength]string{
		StrengthVeryWeak:   "非常弱",
		StrengthWeak:       "弱",
		StrengthMedium:     "中等",
		StrengthStrong:     "强",
		StrengthVeryStrong: "非常强",
	}
	return labels[strength]
}

// maskPassword 掩码密码（用于显示）
func maskPassword(password string) string {
	if len(password) <= 4 {
		return strings.Repeat("*", len(password))
	}
	return password[:2] + strings.Repeat("*", len(password)-4) + password[len(password)-2:]
}

// max 返回最大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
