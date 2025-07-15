package util

import "testing"

func TestPrintTable(t *testing.T) {
	// 示例数据
	data := TableData{
		{"Name", "Age", "Hobbies", "Details"},
		{"Alice", 25, []interface{}{"Reading", "Coding"}, map[string]interface{}{"City": "New York", "Job": "Engineer"}},
		{"Bob", 30, []interface{}{"Gaming", "Traveling"}, map[string]interface{}{"City": "Los Angeles", "Job": "Designer"}},
		{"Charlie", 22, []interface{}{"Cooking", "Painting"}, map[string]interface{}{"City": "San Francisco", "Job": "Artist"}},
	}
	PrintTable(data)
}

func TestTableAlignment(t *testing.T) {
	// 测试中文字符对齐
	headers := []string{"姓名", "年龄", "职业", "城市"}
	data := [][]interface{}{
		{"张三", 25, "工程师", "北京"},
		{"李四", 30, "设计师", "上海"},
		{"王五", 28, "产品经理", "深圳"},
		{"赵六", 35, "架构师", "杭州"},
	}

	t.Log("测试终端表格对齐：")
	RenderTable(headers, data, "terminal")

	// 如果没有 panic，测试就通过了
	t.Log("表格对齐测试通过")
}

func TestRenderTableMarkdownAndHTML(t *testing.T) {
	headers := []string{"姓名", "年龄", "职业", "城市"}
	data := [][]interface{}{
		{"张三", 25, "工程师", "北京"},
		{"李四", 30, "设计师", "上海"},
		{"王五", 28, "产品经理", "深圳"},
		{"赵六", 35, "架构师", "杭州"},
	}

	t.Log("Markdown 格式表格:")
	RenderTable(headers, data, "markdown")

	t.Log("HTML 格式表格:")
	RenderTable(headers, data, "html")
}
