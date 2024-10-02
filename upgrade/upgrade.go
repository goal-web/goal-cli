package upgrade

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// UpdateGoalWebDependencies 更新 `go.mod` 中的 `github.com/goal-web` 依赖到指定 major.minor 版本的最新补丁版本
func UpdateGoalWebDependencies(goModFile string, targetVersion string) error {
	// 打开 go.mod 文件
	file, err := os.Open(goModFile)
	if err != nil {
		return fmt.Errorf("无法打开 go.mod 文件: %v", err)
	}
	defer file.Close()

	// 正则匹配 github.com/goal-web/[组件名] vX.Y.Z
	goalWebRegex := regexp.MustCompile(`(github.com/goal-web/[a-zA-Z0-9_-]+)\s+v([0-9]+\.[0-9]+\.[0-9]+)`)

	// 存储 `github.com/goal-web` 相关依赖
	dependencies := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// 只匹配 `require` 或 tab 缩进的依赖项
		if strings.HasPrefix(line, "require ") || strings.HasPrefix(line, "\t") {
			matches := goalWebRegex.FindStringSubmatch(line)
			if len(matches) == 3 {
				packageName := matches[1]
				version := matches[2]
				dependencies[packageName] = version
			}
		}
	}

	// 检查是否有解析错误
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取 go.mod 文件失败: %v", err)
	}

	// 如果没有指定 targetVersion，则自动使用 `github.com/goal-web/contracts` 的主版本号
	if targetVersion == "" {
		if contractVersion, exists := dependencies["github.com/goal-web/contracts"]; exists {
			targetVersion = extractMajorMinor(contractVersion)
			fmt.Printf("未指定版本号，自动使用 contracts 版本: %s\n", targetVersion)
		} else {
			return fmt.Errorf("未找到 github.com/goal-web/contracts 依赖，且未指定目标版本")
		}
	} else {
		fmt.Printf("使用指定的目标版本: %s\n", targetVersion)
	}

	// 存储需要更新的依赖项
	var updates []string

	// 查询最新版本并更新依赖
	for dep, currentVersion := range dependencies {
		// 提取当前依赖的主版本（vX.Y）
		currentMajorMinor := extractMajorMinor(currentVersion)

		// 如果当前版本与目标版本不同，则进行更新
		if currentMajorMinor != targetVersion {
			// 获取符合 `targetVersion` 的最新补丁版本号
			latestVersion, err := getLatestVersion(dep, targetVersion)
			if err != nil {
				fmt.Printf("获取 %s 最新版本失败: %v\n", dep, err)
				continue
			}

			// 如果当前版本和最新版本不一致，则更新
			if currentVersion != latestVersion {
				updates = append(updates, fmt.Sprintf("%s: %s -> %s", dep, currentVersion, latestVersion))
				// 使用 `go get` 更新指定的依赖
				cmd := exec.Command("go", "get", fmt.Sprintf("%s@%s", dep, latestVersion))
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					fmt.Printf("更新 %s 失败: %v\n", dep, err)
				} else {
					fmt.Printf("已更新 %s 从 %s 到 %s\n", dep, currentVersion, latestVersion)
				}
			}
		}
	}

	// 输出更新结果
	if len(updates) > 0 {
		fmt.Println("以下依赖已被更新:")
		for _, update := range updates {
			fmt.Println(update)
		}
	} else {
		fmt.Println("所有 github.com/goal-web 依赖都已是最新版本或不需要更改。")
	}

	return nil
}

// 获取指定包在 `targetVersion` 版本号下的最新补丁版本
func getLatestVersion(packageName string, targetVersion string) (string, error) {
	// 使用 `go list -m -versions` 来获取所有版本
	cmd := exec.Command("go", "list", "-m", "-versions", packageName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取版本列表失败: %v", err)
	}

	// 解析输出并提取最高版本
	versions := strings.Fields(string(output))
	latestVersion := ""
	for _, version := range versions {
		// 过滤符合 targetVersion 前缀的版本号
		if strings.HasPrefix(version, targetVersion) {
			latestVersion = version
		}
	}

	if latestVersion == "" {
		return "", fmt.Errorf("未找到 %s 对应的最新版本", targetVersion)
	}

	return latestVersion, nil
}

// 提取 `v0.4.3` 格式的主版本号（v0.4）
func extractMajorMinor(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1] // 返回 v0.4
	}
	return version
}
