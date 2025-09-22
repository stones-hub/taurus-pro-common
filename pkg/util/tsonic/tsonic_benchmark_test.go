package tsonic

import (
	"encoding/json"
	"testing"

	jsoniter "github.com/json-iterator/go"
)

var (
	smallJSON = []byte(`{"name":"张三","age":25,"email":"zhangsan@example.com"}`)
	largeJSON = []byte(`{
		"id": 123456,
		"name": "张三",
		"age": 25,
		"email": "zhangsan@example.com",
		"phone": "13800138000",
		"address": {
			"province": "北京市",
			"city": "北京市",
			"district": "朝阳区",
			"street": "三里屯街道",
			"detail": "SOHO大厦"
		},
		"tags": ["Go", "Python", "JavaScript", "Docker", "Kubernetes"],
		"skills": [
			{"name": "Go", "level": 90, "years": 5},
			{"name": "Python", "level": 85, "years": 3},
			{"name": "JavaScript", "level": 80, "years": 4}
		],
		"education": [
			{
				"school": "北京大学",
				"major": "计算机科学",
				"degree": "硕士",
				"start_year": 2015,
				"end_year": 2018,
				"achievements": ["奖学金", "优秀毕业生"]
			},
			{
				"school": "清华大学",
				"major": "软件工程",
				"degree": "学士",
				"start_year": 2011,
				"end_year": 2015,
				"achievements": ["三好学生", "优秀毕业论文"]
			}
		],
		"work_experience": [
			{
				"company": "字节跳动",
				"position": "高级工程师",
				"department": "基础架构部",
				"start_date": "2018-07",
				"end_date": "至今",
				"projects": [
					{
						"name": "微服务框架优化",
						"description": "提升系统性能30%",
						"technologies": ["Go", "gRPC", "Kubernetes"]
					},
					{
						"name": "监控系统重构",
						"description": "降低系统延迟50%",
						"technologies": ["Python", "Prometheus", "Grafana"]
					}
				]
			},
			{
				"company": "腾讯",
				"position": "工程师",
				"department": "腾讯云",
				"start_date": "2015-07",
				"end_date": "2018-06",
				"projects": [
					{
						"name": "云存储系统",
						"description": "开发高可用存储服务",
						"technologies": ["Go", "Redis", "MySQL"]
					}
				]
			}
		],
		"certificates": [
			{
				"name": "AWS Certified Solutions Architect",
				"issue_date": "2019-01",
				"expire_date": "2022-01"
			},
			{
				"name": "Kubernetes Administrator",
				"issue_date": "2020-03",
				"expire_date": "2023-03"
			}
		],
		"languages": [
			{"name": "中文", "level": "母语"},
			{"name": "英语", "level": "专业八级"},
			{"name": "日语", "level": "N2"}
		],
		"hobbies": ["读书", "游泳", "摄影", "旅行"],
		"social_media": {
			"github": "https://github.com/zhangsan",
			"twitter": "https://twitter.com/zhangsan",
			"linkedin": "https://linkedin.com/in/zhangsan"
		},
		"updated_at": "2023-09-01T12:00:00Z"
	}`)
)

type SmallStruct struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

// 大结构体定义
type LargeStruct struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address struct {
		Province string `json:"province"`
		City     string `json:"city"`
		District string `json:"district"`
		Street   string `json:"street"`
		Detail   string `json:"detail"`
	} `json:"address"`
	Tags   []string `json:"tags"`
	Skills []struct {
		Name  string `json:"name"`
		Level int    `json:"level"`
		Years int    `json:"years"`
	} `json:"skills"`
	Education []struct {
		School       string   `json:"school"`
		Major        string   `json:"major"`
		Degree       string   `json:"degree"`
		StartYear    int      `json:"start_year"`
		EndYear      int      `json:"end_year"`
		Achievements []string `json:"achievements"`
	} `json:"education"`
	WorkExperience []struct {
		Company    string `json:"company"`
		Position   string `json:"position"`
		Department string `json:"department"`
		StartDate  string `json:"start_date"`
		EndDate    string `json:"end_date"`
		Projects   []struct {
			Name         string   `json:"name"`
			Description  string   `json:"description"`
			Technologies []string `json:"technologies"`
		} `json:"projects"`
	} `json:"work_experience"`
	Certificates []struct {
		Name       string `json:"name"`
		IssueDate  string `json:"issue_date"`
		ExpireDate string `json:"expire_date"`
	} `json:"certificates"`
	Languages []struct {
		Name  string `json:"name"`
		Level string `json:"level"`
	} `json:"languages"`
	Hobbies     []string `json:"hobbies"`
	SocialMedia struct {
		Github   string `json:"github"`
		Twitter  string `json:"twitter"`
		Linkedin string `json:"linkedin"`
	} `json:"social_media"`
	UpdatedAt string `json:"updated_at"`
}

// 小数据 Unmarshal 基准测试
func BenchmarkSmallUnmarshal(b *testing.B) {
	var data SmallStruct

	b.Run("Sonic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Default.Unmarshal(smallJSON, &data)
		}
	})

	b.Run("StandardLib", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = json.Unmarshal(smallJSON, &data)
		}
	})

	b.Run("JsonIter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = jsoniter.Unmarshal(smallJSON, &data)
		}
	})
}

// 大数据 Unmarshal 基准测试
func BenchmarkLargeUnmarshal(b *testing.B) {
	var data LargeStruct

	b.Run("Sonic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Default.Unmarshal(largeJSON, &data)
		}
	})

	b.Run("StandardLib", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = json.Unmarshal(largeJSON, &data)
		}
	})

	b.Run("JsonIter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = jsoniter.Unmarshal(largeJSON, &data)
		}
	})
}

// 小数据 Marshal 基准测试
func BenchmarkSmallMarshal(b *testing.B) {
	data := SmallStruct{
		Name:  "张三",
		Age:   25,
		Email: "zhangsan@example.com",
	}

	b.Run("Sonic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Default.Marshal(data)
		}
	})

	b.Run("StandardLib", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(data)
		}
	})

	b.Run("JsonIter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = jsoniter.Marshal(data)
		}
	})
}

// 大数据 Marshal 基准测试
func BenchmarkLargeMarshal(b *testing.B) {
	var data LargeStruct
	_ = json.Unmarshal(largeJSON, &data) // 先解析准备数据

	b.Run("Sonic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Default.Marshal(data)
		}
	})

	b.Run("StandardLib", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(data)
		}
	})

	b.Run("JsonIter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = jsoniter.Marshal(data)
		}
	})
}

// 验证性能测试
func BenchmarkValidate(b *testing.B) {
	b.Run("Sonic-Small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Default.Validate(smallJSON)
		}
	})

	b.Run("Sonic-Large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Default.Validate(largeJSON)
		}
	})

	b.Run("StandardLib-Small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = json.Valid(smallJSON)
		}
	})

	b.Run("StandardLib-Large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = json.Valid(largeJSON)
		}
	})
}
