package tfs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriteAndReadLine(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "test_write_line_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// 测试写入
	testData := []byte("test line")
	err = WriteLine(tmpFile.Name(), testData)
	if err != nil {
		t.Errorf("WriteLine() error = %v", err)
		return
	}

	// 测试读取
	lines, err := ReadLine(tmpFile.Name())
	if err != nil {
		t.Errorf("ReadLine() error = %v", err)
		return
	}

	if len(lines) != 1 {
		t.Errorf("ReadLine() got %d lines, want 1", len(lines))
		return
	}

	if lines[0].(string) != string(testData) {
		t.Errorf("ReadLine() got = %v, want %v", lines[0], string(testData))
	}
}

func TestReadAll(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "test_read_all_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// 写入测试数据
	testContent := "test content\nline 2"
	err = os.WriteFile(tmpFile.Name(), []byte(testContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// 测试读取
	content, err := ReadAll(tmpFile.Name())
	if err != nil {
		t.Errorf("ReadAll() error = %v", err)
		return
	}

	if content != testContent {
		t.Errorf("ReadAll() got = %v, want %v", content, testContent)
	}

	// 测试空文件
	emptyFile, err := os.CreateTemp("", "test_empty_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(emptyFile.Name())

	_, err = ReadAll(emptyFile.Name())
	if err == nil {
		t.Error("ReadAll() expected error for empty file")
	}
}

func TestFetchAllDir(t *testing.T) {
	// 创建临时目录结构
	tmpDir, err := os.MkdirTemp("", "test_fetch_dir_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建测试文件
	testFiles := []string{
		"file1.txt",
		"subdir/file2.txt",
		"subdir/subsubdir/file3.txt",
	}

	for _, file := range testFiles {
		path := filepath.Join(tmpDir, file)
		err := os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(path, []byte("test"), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// 测试遍历
	files, err := FetchAllDir(tmpDir)
	if err != nil {
		t.Errorf("FetchAllDir() error = %v", err)
		return
	}

	if len(files) != len(testFiles) {
		t.Errorf("FetchAllDir() got %d files, want %d", len(files), len(testFiles))
	}
}

func TestWalkDir(t *testing.T) {
	// 创建临时目录结构
	tmpDir, err := os.MkdirTemp("", "test_walk_dir_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建测试文件
	testFiles := []string{
		"file1.txt",
		"subdir/file2.txt",
		"subdir/subsubdir/file3.txt",
	}

	for _, file := range testFiles {
		path := filepath.Join(tmpDir, file)
		err := os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(path, []byte("test"), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// 测试不同深度
	tests := []struct {
		name    string
		depth   int
		wantLen int
	}{
		{"depth 0", 0, 1},
		{"depth 1", 1, 2},
		{"depth 2", 2, 3},
		{"depth 3", 3, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := WalkDir(tmpDir, tt.depth)
			if err != nil {
				t.Errorf("WalkDir() error = %v", err)
				return
			}

			if len(files) != tt.wantLen {
				t.Errorf("WalkDir() got %d files, want %d", len(files), tt.wantLen)
			}
		})
	}
}

func TestFileMove(t *testing.T) {
	// 创建临时目录和文件
	tmpDir, err := os.MkdirTemp("", "test_file_move_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "subdir/dest.txt")

	// 创建源文件
	err = os.WriteFile(srcFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// 测试移动
	err = FileMove(srcFile, dstFile)
	if err != nil {
		t.Errorf("FileMove() error = %v", err)
		return
	}

	// 验证源文件不存在
	if FileExist(srcFile) {
		t.Error("FileMove() source file still exists")
	}

	// 验证目标文件存在
	if !FileExist(dstFile) {
		t.Error("FileMove() destination file does not exist")
	}
}

func TestGenCSV(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "test_csv_*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// 测试数据
	data := []map[string]string{
		{
			"name":  "John",
			"age":   "30",
			"email": "john@example.com",
		},
		{
			"name":  "Jane",
			"age":   "25",
			"email": "jane@example.com",
		},
	}
	headers := []string{"name", "age", "email"}

	// 测试生成CSV
	err = GenCSV(tmpFile.Name(), data, headers)
	if err != nil {
		t.Errorf("GenCSV() error = %v", err)
		return
	}

	// 验证文件存在
	if !FileExist(tmpFile.Name()) {
		t.Error("GenCSV() file was not created")
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want string
	}{
		{"bytes", 500, "500 B"},
		{"kilobytes", 1024, "1.00 KB"},
		{"megabytes", 1024 * 1024, "1.00 MB"},
		{"gigabytes", 1024 * 1024 * 1024, "1.00 GB"},
		{"terabytes", 1024 * 1024 * 1024 * 1024, "1.00 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatFileSize(tt.size); got != tt.want {
				t.Errorf("FormatFileSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileWatcher(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "test_watch_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// 写入初始内容
	err = os.WriteFile(tmpFile.Name(), []byte("initial content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// 创建通道用于同步测试
	changed := make(chan bool)

	// 创建文件监控器
	watcher, err := NewFileWatcher(tmpFile.Name(), 100*time.Millisecond, func() {
		changed <- true
	})
	if err != nil {
		t.Fatalf("NewFileWatcher() error = %v", err)
	}

	// 启动监控
	watcher.Start()
	defer watcher.Stop()

	// 修改文件
	time.Sleep(200 * time.Millisecond)
	err = os.WriteFile(tmpFile.Name(), []byte("new content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// 等待变更通知
	select {
	case <-changed:
		// 成功检测到变更
	case <-time.After(time.Second):
		t.Error("FileWatcher did not detect file change")
	}
}

func TestPathOperations(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		segments []string
		depth    int
	}{
		{
			name:     "simple path",
			path:     "/path/to/file.txt",
			segments: []string{"path", "to", "file.txt"},
			depth:    3,
		},
		{
			name:     "path with trailing slash",
			path:     "/path/to/dir/",
			segments: []string{"path", "to", "dir"},
			depth:    3,
		},
		{
			name:     "single file",
			path:     "file.txt",
			segments: []string{"file.txt"},
			depth:    1,
		},
		{
			name:     "empty path",
			path:     "",
			segments: []string{},
			depth:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试GetPathSegments
			segments := GetPathSegments(tt.path)
			if len(segments) != len(tt.segments) {
				t.Errorf("GetPathSegments() got %v, want %v", segments, tt.segments)
			}

			// 测试GetPathDepth
			depth := GetPathDepth(tt.path)
			if depth != tt.depth {
				t.Errorf("GetPathDepth() got %v, want %v", depth, tt.depth)
			}

			// 测试JoinPathSegments
			if len(tt.segments) > 0 {
				joined := JoinPathSegments(tt.segments)
				if joined != strings.Join(tt.segments, "/") {
					t.Errorf("JoinPathSegments() got %v, want %v", joined, strings.Join(tt.segments, "/"))
				}
			}
		})
	}
}

func TestCopyOperations(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "test_copy_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 测试文件复制
	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "dest.txt")
	content := []byte("test content")

	err = os.WriteFile(srcFile, content, 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = CopyFile(srcFile, dstFile)
	if err != nil {
		t.Errorf("CopyFile() error = %v", err)
	}

	// 验证文件内容
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(dstContent) != string(content) {
		t.Errorf("CopyFile() content mismatch, got %v, want %v", string(dstContent), string(content))
	}

	// 测试目录复制
	srcDir := filepath.Join(tmpDir, "src")
	dstDir := filepath.Join(tmpDir, "dst")
	err = os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	// 在源目录中创建一些文件
	files := []string{
		"file1.txt",
		"subdir/file2.txt",
	}
	for _, f := range files {
		err = os.WriteFile(filepath.Join(srcDir, f), content, 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	err = CopyDir(srcDir, dstDir)
	if err != nil {
		t.Errorf("CopyDir() error = %v", err)
	}

	// 验证目录结构
	for _, f := range files {
		if !FileExist(filepath.Join(dstDir, f)) {
			t.Errorf("CopyDir() file %s not copied", f)
		}
	}
}
