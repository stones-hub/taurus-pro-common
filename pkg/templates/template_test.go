package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestTemplateManager(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "template-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试目录结构
	baseDir := filepath.Join(tempDir, "templates")
	mainDir := filepath.Join(baseDir, "main")
	subDir := filepath.Join(baseDir, "sub")
	mustCreateDir(t, mainDir)
	mustCreateDir(t, subDir)

	// 创建测试模板文件
	mainTemplate := `
{{define "layout"}}
<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>{{template "content" .}}</body>
</html>
{{end}}

{{define "content"}}
<h1>{{.Message}}</h1>
{{end}}`

	subTemplate := `
{{define "sub/footer"}}
<footer>
    <p>{{.Copyright}}</p>
    {{if .ShowContact}}
    <p>联系我们: {{.Contact}}</p>
    {{end}}
</footer>
{{end}}`

	mustCreateFile(t, filepath.Join(mainDir, "main.html"), mainTemplate)
	mustCreateFile(t, filepath.Join(subDir, "footer.html"), subTemplate)

	t.Run("初始化加载模板", func(t *testing.T) {
		// 1. 测试初始化加载
		manager, cleanup, err := New(
			TemplateOptions{
				Name: "main",
				Path: mainDir,
			},
			TemplateOptions{
				Name: "sub",
				Path: subDir,
			},
		)
		if err != nil {
			t.Fatalf("初始化加载模板失败: %v", err)
		}
		defer cleanup()

		// 测试渲染主模板
		data := map[string]interface{}{
			"Title":   "测试页面",
			"Message": "欢迎访问",
		}
		result, err := manager.Render("main", "layout", data)
		if err != nil {
			t.Fatalf("渲染主模板失败: %v", err)
		}
		if !contains(result, "测试页面") || !contains(result, "欢迎访问") {
			t.Errorf("渲染结果不包含预期内容: %s", result)
		}

		// 测试渲染子模板
		footerData := map[string]interface{}{
			"Copyright":   "© 2025 Taurus",
			"ShowContact": true,
			"Contact":     "contact@example.com",
		}
		result, err = manager.Render("sub", "sub/footer", footerData)
		if err != nil {
			t.Fatalf("渲染子模板失败: %v", err)
		}
		if !contains(result, "© 2025 Taurus") || !contains(result, "contact@example.com") {
			t.Errorf("渲染结果不包含预期内容: %s", result)
		}
	})

	t.Run("动态加载模板", func(t *testing.T) {
		manager, cleanup, err := New()
		if err != nil {
			t.Fatalf("创建模板管理器失败: %v", err)
		}
		defer cleanup()

		// 1. 动态添加基础布局
		err = manager.AddTemplate("dynamic", "base", `
			{{define "base"}}
			<div class="container">
				<nav>{{template "nav" .}}</nav>
				<main>{{template "main" .}}</main>
			</div>
			{{end}}
		`)
		if err != nil {
			t.Fatalf("添加基础布局失败: %v", err)
		}

		// 2. 动态添加导航模板
		err = manager.AddTemplate("dynamic", "nav", `
			{{define "nav"}}
			<ul>
				{{range .NavItems}}
				<li><a href="{{.URL}}">{{.Text}}</a></li>
				{{end}}
			</ul>
			{{end}}
		`)
		if err != nil {
			t.Fatalf("添加导航模板失败: %v", err)
		}

		// 3. 动态添加主内容模板
		err = manager.AddTemplate("dynamic", "main", `
			{{define "main"}}
			<article>
				<h1>{{.Title}}</h1>
				<p>{{.Content}}</p>
			</article>
			{{end}}
		`)
		if err != nil {
			t.Fatalf("添加主内容模板失败: %v", err)
		}

		// 4. 渲染完整页面
		data := map[string]interface{}{
			"Title":   "动态模板测试",
			"Content": "这是一个动态加载的模板示例",
			"NavItems": []struct {
				URL  string
				Text string
			}{
				{"/home", "首页"},
				{"/about", "关于"},
			},
		}

		result, err := manager.Render("dynamic", "base", data)
		if err != nil {
			t.Fatalf("渲染动态模板失败: %v", err)
		}

		// 验证渲染结果
		expectedContents := []string{
			"动态模板测试",
			"这是一个动态加载的模板示例",
			"首页",
			"关于",
			"/home",
			"/about",
		}
		for _, expected := range expectedContents {
			if !contains(result, expected) {
				t.Errorf("渲染结果不包含预期内容 %q: %s", expected, result)
			}
		}
	})

	t.Run("并发测试", func(t *testing.T) {
		manager, cleanup, err := New()
		if err != nil {
			t.Fatalf("创建模板管理器失败: %v", err)
		}
		defer cleanup()

		var wg sync.WaitGroup

		// 并发动态添加新模板
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				name := fmt.Sprintf("dynamic_template_%d", idx)
				content := fmt.Sprintf(`<p>Template {{.Text}} - %d</p>`, idx)
				err := manager.AddTemplate("concurrent", name, content)
				if err != nil {
					t.Errorf("并发添加模板失败: %v", err)
					return
				}
			}(i)
		}

		wg.Wait()

		// 等待所有模板添加完成后再并发渲染
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				name := fmt.Sprintf("dynamic_template_%d", idx)
				result, err := manager.Render("concurrent", name, map[string]interface{}{
					"Text": "动态内容",
				})
				if err != nil {
					t.Errorf("渲染模板失败: %v", err)
					return
				}
				expected := fmt.Sprintf("动态内容 - %d", idx)
				if !contains(result, expected) {
					t.Errorf("渲染结果不包含预期内容 %q: %s", expected, result)
				}
			}(i)
		}

		wg.Wait()
	})
}

// Helper functions
func mustCreateDir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("创建目录失败 %s: %v", dir, err)
	}
}

func mustCreateFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("创建文件失败 %s: %v", path, err)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
