package tcmd

import (
	"strings"
	"testing"
)

func TestCmdExec(t *testing.T) {
	tests := []struct {
		name    string
		cmd     string
		args    []string
		want    string
		wantErr bool
	}{
		{
			name: "echo command",
			cmd:  "echo",
			args: []string{"hello world"},
			want: "hello world",
		},
		{
			name: "pwd command",
			cmd:  "pwd",
			args: []string{},
			want: "", // 实际运行时会返回当前目录
		},
		{
			name:    "invalid command",
			cmd:     "invalidcommand123",
			args:    []string{},
			wantErr: true,
		},
		{
			name: "ls command",
			cmd:  "ls",
			args: []string{"-l"},
			want: "", // 实际运行时会返回目录列表
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CmdExec(tt.cmd, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("CmdExec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				switch tt.cmd {
				case "echo":
					if got != tt.want {
						t.Errorf("CmdExec() = %v, want %v", got, tt.want)
					}
				case "pwd":
					if !strings.Contains(got, "/") {
						t.Errorf("CmdExec() pwd result should contain path separator")
					}
				case "ls":
					if len(got) == 0 {
						t.Errorf("CmdExec() ls result should not be empty")
					}
				}
			}
		})
	}
}

func TestCmdExecWithPanic(t *testing.T) {
	// 测试命令执行过程中的panic恢复
	_, err := CmdExec("echo", strings.Repeat("a", 1<<30)) // 尝试创建一个非常大的字符串来触发panic
	if err == nil {
		t.Error("Expected error for large input, got nil")
	}
}
