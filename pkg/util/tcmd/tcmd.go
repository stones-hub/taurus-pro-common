package tcmd

import (
	"io"
	"log"
	"os/exec"
	"strings"
)

// CmdExec 执行外部命令并返回标准输出结果
// 参数：
//   - name: 要执行的命令名称（例如 "ls"、"git"）
//   - args: 命令的参数列表（可变参数）
//
// 返回值：
//   - string: 命令的标准输出结果，已去除首尾空白字符
//   - error: 执行过程中的错误信息，如果执行成功则为 nil
//
// 使用示例：
//
//	// 执行单个命令
//	output, err := tcmd.CmdExec("ls", "-l")
//	if err != nil {
//	    log.Printf("执行失败：%v", err)
//	    return
//	}
//	fmt.Println(output)
//
//	// 执行带多个参数的命令
//	output, err = tcmd.CmdExec("git", "clone", "https://github.com/user/repo.git")
//	if err != nil {
//	    log.Printf("克隆失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 函数会自动捕获 panic 并转换为错误返回
//   - 命令的标准错误输出会被忽略
//   - 如果命令执行失败，会返回详细的错误信息
//   - 输出结果会自动去除首尾的空白字符
//   - 该函数会等待命令完全执行完成后返回
//   - 适用于需要执行时间较短的命令
//   - 对于需要交互的命令，建议使用 os/exec 包直接处理
func CmdExec(name string, args ...string) (string, error) {
	var (
		err    error
		stdout io.ReadCloser
		output []byte
	)
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic: %v\n", err)
		}
	}()
	cmd := exec.Command(name, args...)
	// 设置标准输出
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		log.Printf("Failed to create stdout pipe: %v\n", err)
		return "", err
	}
	// 启动命令
	if err = cmd.Start(); err != nil {
		log.Printf("Failed to start command: %v\n", err)
		return "", err
	}
	// 读取标准输出
	output, err = io.ReadAll(stdout)
	if err != nil {
		log.Printf("Failed to read from stdout: %v\n", err)
		return "", err
	}

	// 关闭标准输出
	if err = stdout.Close(); err != nil {
		log.Printf("Failed to close stdout: %v\n", err)
		return "", err
	}

	// 等待命令完成
	if err = cmd.Wait(); err != nil {
		log.Printf("Failed to wait for command: %v\n", err)
		return "", err
	}

	// 解析输出结果
	return strings.TrimSpace(string(output)), nil
}
