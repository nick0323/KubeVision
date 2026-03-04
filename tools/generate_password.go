package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"

	"github.com/nick0323/K8sVision/model"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("使用方法:")
		fmt.Println("  go run tools/generate_password.go <密码>")
		fmt.Println("  go run tools/generate_password.go generate <长度>")
		fmt.Println("")
		fmt.Println("示例:")
		fmt.Println("  go run tools/generate_password.go admin123!")
		fmt.Println("  go run tools/generate_password.go generate 16")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "generate":
		length := 12
		if len(os.Args) > 2 {
			if l, err := strconv.Atoi(os.Args[2]); err == nil {
				length = l
			}
		}
		generateRandomPassword(length)
	default:
		hashPassword(command)
	}
}

func hashPassword(password string) {
	// 生成随机盐值
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		fmt.Printf("生成盐值失败: %v\n", err)
		os.Exit(1)
	}

	// 将密码与盐值结合
	passwordWithSalt := password + base64.URLEncoding.EncodeToString(salt)

	// 使用bcrypt哈希
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(passwordWithSalt), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("密码哈希失败: %v\n", err)
		os.Exit(1)
	}

	// 返回格式: base64(salt):bcrypt_hash
	hashedPassword := base64.URLEncoding.EncodeToString(salt) + ":" + string(hashedBytes)

	fmt.Println("=== 密码哈希生成结果 ===")
	fmt.Printf("原始密码: %s\n", password)
	fmt.Printf("哈希密码: %s\n", hashedPassword)
	fmt.Println("")
	fmt.Println("=== 配置方法 ===")
	fmt.Println("1. 在config.yaml中设置:")
	fmt.Printf("   auth:\n")
	fmt.Printf("     username: \"admin\"\n")
	fmt.Printf("     password: \"%s\"\n", hashedPassword)
	fmt.Println("")
	fmt.Println("2. 或通过环境变量设置:")
	fmt.Printf("   export K8SVISION_AUTH_USERNAME=\"admin\"\n")
	fmt.Printf("   export K8SVISION_AUTH_PASSWORD=\"%s\"\n", hashedPassword)
}

func generateRandomPassword(length int) {
	password := make([]byte, length)
	charsetBytes := []byte(model.PasswordCharset)

	for i := range password {
		randomIndex := make([]byte, 1)
		if _, err := rand.Read(randomIndex); err != nil {
			fmt.Printf("生成随机字符失败: %v\n", err)
			os.Exit(1)
		}
		password[i] = charsetBytes[randomIndex[0]%byte(len(charsetBytes))]
	}

	passwordStr := string(password)
	fmt.Printf("生成的随机密码: %s\n", passwordStr)
	fmt.Println("")

	// 同时生成哈希
	hashPassword(passwordStr)
}
