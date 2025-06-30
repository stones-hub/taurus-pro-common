package util

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	file = "/test/files/test.csv"
)

type Employee struct {
	Name     string    `json:"name,omitempty"`   // 姓名
	Position string    `json:"position"`         // 职位
	Salary   int64     `json:"salary,omitempty"` // 薪资
	JoinDate time.Time // 加入日期
	User     User      `json:"user"`
}

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
	Date  time.Time
}

func TestExcelWriter(t *testing.T) {

	basePath, err := os.Getwd()
	if err != nil {
		t.Errorf("get work dir error : %v\n", err.Error())
	}
	file = filepath.Dir(filepath.Dir(basePath)) + file

	employees := []Employee{
		{Name: "Alice", Position: "Software Engineer", Salary: 100000, JoinDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			User: User{Name: "Alice1", Email: "alice@example.com", Age: 30, Date: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)}},

		{Name: "Bob", Position: "Product Manager", Salary: 120000, JoinDate: time.Date(2019, 6, 1, 0, 0, 0, 0, time.UTC),
			User: User{Name: "Bob1", Email: "bob@example.com", Age: 35, Date: time.Date(2019, 6, 1, 0, 0, 0, 0, time.UTC)}},
	}

	empWriter, err := InitExcelWriter(file, []string{"Name", "Position", "Salary", "JoinDate", "User"})
	if err != nil {
		t.Errorf("init excel writer error : %v\n", err.Error())
		return
	}
	defer empWriter.Close()

	// 将切片转换为 []interface{},  语法上是不允许将一个有类型的slice 等同于 一个 []interface{}的类型的，尽管 interface{}可以代表任何类型
	// 这里就是把有类型的slice一个个遍历出来，然后写入到[]interface{} 的slice中
	employeeBatch := make([]interface{}, len(employees))
	for i, item := range employees {
		employeeBatch[i] = item
	}

	// 写入 Excel文件, 如果要分批写也很简单， 将employees slice切分成多个slice传入就可以了
	if err := empWriter.WriteBatch(employeeBatch); err != nil {
		t.Errorf("init excel writer data to file error : %v\n", err.Error())
		return
	}
}
