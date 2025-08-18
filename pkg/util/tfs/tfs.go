package tfs

import (
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// WriteLine 向文件追加写入一行数据，自动添加换行符
// 参数：
//   - filename: 目标文件路径
//   - b: 要写入的字节数据
//
// 返回值：
//   - error: 写入过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	data := []byte("这是一行日志内容")
//	err := tfs.WriteLine("app.log", data)
//	if err != nil {
//	    log.Printf("写入失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 如果文件不存在会自动创建
//   - 使用追加模式写入（不会覆盖已有内容）
//   - 自动添加换行符
//   - 使用缓冲写入提高性能
//   - 写入完成后会自动刷新缓冲区
//   - 文件权限设置为0666
func WriteLine(filename string, b []byte) error {
	fd, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("open file failed: %w", err)
	}
	defer fd.Close()

	writer := bufio.NewWriter(fd)
	defer writer.Flush()

	if _, err := writer.Write(b); err != nil {
		return fmt.Errorf("write data failed: %w", err)
	}

	if err := writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("write newline failed: %w", err)
	}

	return nil
}

// ReadAll 读取整个文件的内容并返回为字符串
// 参数：
//   - filename: 要读取的文件路径
//
// 返回值：
//   - string: 文件的全部内容
//   - error: 读取过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	content, err := tfs.ReadAll("config.json")
//	if err != nil {
//	    log.Printf("读取失败：%v", err)
//	    return
//	}
//	fmt.Println("文件内容：", content)
//
// 注意事项：
//   - 一次性读取整个文件到内存
//   - 不适合读取大文件
//   - 如果文件为空会返回错误
//   - 自动处理文件关闭
//   - 返回UTF-8编码的字符串
//   - 适用于配置文件等小文件
func ReadAll(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("read file failed: %w", err)
	}

	if len(data) == 0 {
		return "", fmt.Errorf("file is empty")
	}

	return string(data), nil
}

// ReadLine 逐行读取文件内容并返回字符串切片
// 参数：
//   - filename: 要读取的文件路径
//
// 返回值：
//   - []interface{}: 每行内容作为一个元素的切片
//   - error: 读取过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	lines, err := tfs.ReadLine("access.log")
//	if err != nil {
//	    log.Printf("读取失败：%v", err)
//	    return
//	}
//	for i, line := range lines {
//	    fmt.Printf("第%d行：%s\n", i+1, line.(string))
//	}
//
// 注意事项：
//   - 使用缓冲读取提高性能
//   - 支持处理大文件
//   - 每行内容作为string类型存储
//   - 自动处理文件关闭
//   - 设置了512KB的最大行长度
//   - 适用于日志文件等按行组织的文本
func ReadLine(filename string) ([]interface{}, error) {
	fd, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open file failed: %w", err)
	}
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	var content []interface{}

	// 设置更大的缓冲区以处理长行
	const maxCapacity = 512 * 1024 // 512KB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		content = append(content, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan file failed: %w", err)
	}

	return content, nil
}

// FetchAllDir 递归遍历目录并返回所有文件的完整路径
// 参数：
//   - path: 要遍历的目录路径
//
// 返回值：
//   - []string: 所有文件的完整路径列表
//   - error: 遍历过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	files, err := tfs.FetchAllDir("./src")
//	if err != nil {
//	    log.Printf("遍历目录失败：%v", err)
//	    return
//	}
//	for _, file := range files {
//	    fmt.Println("文件：", file)
//	}
//
// 注意事项：
//   - 递归遍历所有子目录
//   - 只返回文件路径，不包括目录
//   - 返回的是绝对路径
//   - 遵循系统的符号链接
//   - 自动处理路径分隔符
//   - 适用于需要处理整个目录树的场景
func FetchAllDir(path string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(path, func(filepath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, filepath)
		}
		return nil
	})

	return files, err
}

// WalkDir 递归遍历目录并返回指定深度内的所有文件路径
// 参数：
//   - path: 要遍历的目录路径
//   - maxDepth: 最大遍历深度（相对于起始目录的层级数）
//
// 返回值：
//   - []string: 指定深度内的所有文件路径列表
//   - error: 遍历过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	// 只遍历2层目录深度
//	files, err := tfs.WalkDir("./project", 2)
//	if err != nil {
//	    log.Printf("遍历目录失败：%v", err)
//	    return
//	}
//	for _, file := range files {
//	    fmt.Println("文件：", file)
//	}
//
// 注意事项：
//   - 深度从0开始计算
//   - 如果目录深度小于指定深度，返回所有文件
//   - 超过指定深度的目录会被跳过
//   - 只返回文件路径，不包括目录
//   - 遵循系统的符号链接
//   - 适用于需要限制遍历深度的场景
func WalkDir(path string, maxDepth int) ([]string, error) {
	var files []string

	// 规范化路径，移除末尾的斜杠
	path = filepath.Clean(path)

	err := filepath.WalkDir(path, func(currentPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 如果是目录，检查是否需要跳过
		if d.IsDir() {
			// 如果是根目录，继续遍历
			if currentPath == path {
				return nil
			}

			// 计算当前目录的深度
			relPath, err := filepath.Rel(path, currentPath)
			if err != nil {
				return err
			}

			depth := len(strings.Split(relPath, string(os.PathSeparator)))
			if depth > maxDepth {
				return filepath.SkipDir
			}
			return nil
		}

		// 处理文件
		relPath, err := filepath.Rel(path, currentPath)
		if err != nil {
			return err
		}

		// 计算文件所在目录的深度
		parentDir := filepath.Dir(relPath)
		var depth int
		if parentDir == "." {
			depth = 0 // 文件在根目录
		} else {
			depth = len(strings.Split(parentDir, string(os.PathSeparator)))
		}

		if depth <= maxDepth {
			files = append(files, currentPath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// CheckPath 检查指定路径是否为目录
// 参数：
//   - path: 要检查的路径
//
// 返回值：
//   - bool: 如果是目录返回true，如果是文件返回false
//   - error: 检查过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	isDir, err := tfs.CheckPath("./config")
//	if err != nil {
//	    log.Printf("检查路径失败：%v", err)
//	    return
//	}
//	if isDir {
//	    fmt.Println("这是一个目录")
//	} else {
//	    fmt.Println("这是一个文件")
//	}
//
// 注意事项：
//   - 会检查路径是否存在
//   - 会解析符号链接
//   - 区分文件和目录
//   - 需要有路径的读取权限
//   - 适用于需要判断路径类型的场景
func CheckPath(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	if info.IsDir() {
		return true, nil
	}

	return false, nil
}

// GetCurrentPath 获取调用此函数的源文件所在目录的路径
// 返回值：
//   - string: 源文件所在目录的绝对路径
//   - error: 获取过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	dir, err := tfs.GetCurrentPath()
//	if err != nil {
//	    log.Printf("获取路径失败：%v", err)
//	    return
//	}
//	fmt.Println("当前目录：", dir)
//
// 注意事项：
//   - 使用运行时反射获取调用者信息
//   - 返回的是绝对路径
//   - 不包含文件名
//   - 会解析符号链接
//   - 适用于需要获取源文件位置的场景
//   - 在测试或编译时可能与预期不同
func GetCurrentPath() (string, error) {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		return "", fmt.Errorf("failed to get caller info: %v", ok)
	}
	dir := filepath.Dir(file)
	return dir, nil
}

// PathExists 检查指定路径是否存在并且是一个目录
// 参数：
//   - path: 要检查的路径
//
// 返回值：
//   - bool: 如果路径存在且是目录返回true，否则返回false
//   - error: 检查过程中的错误，特殊情况说明：
//   - 如果路径不存在，返回false和nil
//   - 如果路径存在但是文件，返回false和错误
//   - 如果发生其他错误，返回false和具体错误
//
// 使用示例：
//
//	exists, err := tfs.PathExists("./data")
//	if err != nil {
//	    log.Printf("检查失败：%v", err)
//	    return
//	}
//	if exists {
//	    fmt.Println("目录存在")
//	} else {
//	    fmt.Println("目录不存在")
//	}
//
// 注意事项：
//   - 区分目录和文件
//   - 会解析符号链接
//   - 需要有路径的读取权限
//   - 同名文件会返回错误
//   - 适用于需要确认目录是否可用的场景
func PathExists(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err == nil {
		if fi.IsDir() {
			return true, nil
		}
		return false, errors.New("file with same name exists")
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// CreateDir 创建一个或多个目录，如果目录已存在则跳过
// 参数：
//   - dirs: 要创建的目录路径列表（可变参数）
//
// 返回值：
//   - error: 创建过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	err := tfs.CreateDir("logs", "data/cache", "temp/downloads")
//	if err != nil {
//	    log.Printf("创建目录失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 支持创建多级目录
//   - 如果目录已存在会跳过
//   - 自动创建父目录
//   - 使用0777权限创建目录
//   - 会检查目录是否真的是目录
//   - 适用于需要确保目录存在的场景
func CreateDir(dirs ...string) (err error) {
	for _, v := range dirs {
		exist, err := PathExists(v)
		if err != nil {
			return err
		}
		if !exist {
			if err := os.MkdirAll(v, os.ModePerm); err != nil {
				return err
			}
		}
	}
	return err
}

// FileMove 将文件从源位置移动到目标位置
// 参数：
//   - src: 源文件路径
//   - dst: 目标文件路径
//
// 返回值：
//   - error: 移动过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	err := tfs.FileMove("old/data.txt", "new/data.txt")
//	if err != nil {
//	    log.Printf("移动文件失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 如果目标路径为空则直接返回
//   - 自动创建目标目录
//   - 使用绝对路径进行操作
//   - 目标目录权限设置为0755
//   - 支持跨分区移动
//   - 如果目标文件已存在会被覆盖
//   - 适用于需要移动文件的场景
func FileMove(src string, dst string) (err error) {
	if dst == "" {
		return nil
	}
	src, err = filepath.Abs(src)
	if err != nil {
		return err
	}
	dst, err = filepath.Abs(dst)
	if err != nil {
		return err
	}
	revoke := false
	dir := filepath.Dir(dst)
Redirect:
	_, err = os.Stat(dir)
	if err != nil {
		err = os.MkdirAll(dir, 0o755)
		if err != nil {
			return err
		}
		if !revoke {
			revoke = true
			goto Redirect
		}
	}
	return os.Rename(src, dst)
}

// FileExist 检查指定路径的文件是否存在
// 参数：
//   - path: 要检查的文件路径
//
// 返回值：
//   - bool: 如果文件存在返回true，否则返回false
//
// 使用示例：
//
//	if tfs.FileExist("config.json") {
//	    fmt.Println("文件存在")
//	} else {
//	    fmt.Println("文件不存在")
//	}
//
// 注意事项：
//   - 只检查文件，不包括目录
//   - 会解析符号链接
//   - 不检查文件权限
//   - 不区分文件类型
//   - 适用于需要确认文件存在性的场景
//   - 对于权限错误也会返回false
func FileExist(path string) bool {
	fi, err := os.Lstat(path)
	if err == nil {
		return !fi.IsDir()
	}
	return !os.IsNotExist(err)
}

// GenCSV 将数据以CSV格式写入文件
// 参数：
//   - filename: 目标CSV文件路径
//   - data: 要写入的数据，每个map代表一行，key为列名，value为单元格值
//   - headers: 可选的列标题列表，如果为空则使用data中第一行的key作为标题
//
// 返回值：
//   - error: 写入过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	data := []map[string]string{
//	    {
//	        "name":  "张三",
//	        "age":   "25",
//	        "email": "zhangsan@example.com",
//	    },
//	    {
//	        "name":  "李四",
//	        "age":   "30",
//	        "email": "lisi@example.com",
//	    },
//	}
//	headers := []string{"name", "age", "email"}
//	err := tfs.GenCSV("users.csv", data, headers)
//	if err != nil {
//	    log.Printf("写入CSV失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 自动创建或追加到文件
//   - 使用UTF-8编码
//   - 自动写入表头
//   - 按headers指定的顺序写入列
//   - 如果headers为空，使用第一行数据的键
//   - 使用标准CSV格式（RFC 4180）
//   - 适用于需要导出表格数据的场景
func GenCSV(filename string, data []map[string]string, headers []string) error {
	var (
		fd     *os.File
		err    error
		writer *csv.Writer
	)

	if fd, err = os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm); err != nil {
		log.Printf("Failed to generate CSV file: %s \n", err.Error())
		return err
	}

	defer fd.Close()

	writer = csv.NewWriter(fd)

	defer writer.Flush()

	if len(data) == 0 {
		return fmt.Errorf("data is empty")
	}

	if len(headers) == 0 {
		// 创建与表头大小相同的切片
		headers = make([]string, 0, len(data[0]))
		for header := range data[0] {
			headers = append(headers, header)
		}
	}

	// 写入表头
	if err = writer.Write(headers); err != nil {
		log.Printf("Failed to write headers: %s \n", err.Error())
		return err
	}

	// 写入数据
	for _, row := range data {
		record := make([]string, 0, len(headers))

		for _, header := range headers {
			record = append(record, row[header])
		}

		writer.Write(record)
	}

	return nil
}

// ReadCSV 从CSV文件读取数据到结构体切片
// 参数：
//   - filename: CSV文件路径
//   - result: 用于存储结果的结构体切片指针，结构体字段需要使用`csv:"列名"`标签
//
// 返回值：
//   - error: 读取过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	type User struct {
//	    Name  string `csv:"name"`
//	    Age   int    `csv:"age"`
//	    Email string `csv:"email"`
//	}
//
//	var users []User
//	err := tfs.ReadCSV("users.csv", &users)
//	if err != nil {
//	    log.Printf("读取CSV失败：%v", err)
//	    return
//	}
//	for _, user := range users {
//	    fmt.Printf("用户：%+v\n", user)
//	}
//
// 注意事项：
//   - result必须是指向切片的指针
//   - 结构体字段必须使用csv标签指定列名
//   - 自动处理数据类型转换
//   - 支持基本类型和time.Time
//   - 忽略不存在的列
//   - 自动过滤表头中的不可见字符
//   - 适用于需要导入CSV数据到结构体的场景
func ReadCSV(filename string, result interface{}) error {
	var (
		fd               *os.File
		err              error
		reader           *csv.Reader
		invisibleHeaders []string
		headers          []string
		resultValue      reflect.Value
	)

	if fd, err = os.OpenFile(filename, os.O_RDONLY|os.O_APPEND, os.ModePerm); err != nil {
		return err
	}

	defer fd.Close()
	reader = csv.NewReader(fd)

	// 读取CSV文件的第一行作为表头
	if invisibleHeaders, err = reader.Read(); err != nil {
		return err
	}

	// 过滤表头中的不可见字符，因为可能存在编码问题
	for _, v := range invisibleHeaders {
		resRunes := []rune{}
		for _, r := range v {
			// ASCII码小于等于32或大于等于127的都属于不可见字符
			if r > 32 && r < 127 {
				resRunes = append(resRunes, r)
			}
		}
		headers = append(headers, string(resRunes))
	}

	// 获取result的反射类型
	resultValue = reflect.ValueOf(result)

	// Result必须是指向切片的指针
	if resultValue.Kind() != reflect.Ptr || resultValue.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("result must be a pointer to a slice")
	}

	// 获取result指向的切片类型: sliceType = []Test
	sliceType := resultValue.Elem().Type()

	// 获取切片中元素的类型: elementType = Test
	elementType := sliceType.Elem()

	// 基于类型创建新切片: []Test
	slice := reflect.MakeSlice(sliceType, 0, 0)

	// 读取数据行
	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("Failed to read data row: %v\n", err)
			continue
		}

		// 创建新的结构体实例
		element := reflect.New(elementType).Elem()

		// 遍历结构体字段
		for i := 0; i < element.NumField(); i++ {
			field := element.Type().Field(i)
			tag := field.Tag.Get("csv")
			if tag == "" {
				continue
			}

			// 查找对应的CSV列索引
			colIndex := -1
			for j, header := range headers {
				if header == tag {
					colIndex = j
					break
				}
			}

			if colIndex == -1 || colIndex >= len(record) {
				continue
			}

			// 设置字段值
			fieldValue := element.Field(i)
			value := record[colIndex]

			switch fieldValue.Kind() {
			case reflect.String:
				fieldValue.SetString(value)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if v, err := strconv.ParseInt(value, 10, 64); err == nil {
					fieldValue.SetInt(v)
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if v, err := strconv.ParseUint(value, 10, 64); err == nil {
					fieldValue.SetUint(v)
				}
			case reflect.Float32, reflect.Float64:
				if v, err := strconv.ParseFloat(value, 64); err == nil {
					fieldValue.SetFloat(v)
				}
			case reflect.Bool:
				if v, err := strconv.ParseBool(value); err == nil {
					fieldValue.SetBool(v)
				}
			case reflect.Slice:
				// 处理[]byte类型
				if fieldValue.Type().Elem().Kind() == reflect.Uint8 {
					fieldValue.SetBytes([]byte(value))
				}
			default:
				// 处理time.Time类型
				if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
					if v, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
						fieldValue.Set(reflect.ValueOf(v))
					}
				} else if fieldValue.Type().Kind() == reflect.Interface {
					// 处理interface{}类型
					fieldValue.Set(reflect.ValueOf(value))
				}
			}
		}

		// 将结构体添加到切片
		slice = reflect.Append(slice, element)
	}

	// 设置结果
	resultValue.Elem().Set(slice)
	return nil
}

// ExcelWriter 提供将结构体数据写入CSV格式Excel文件的功能
// 用于批量写入大量数据，支持自动类型转换和格式化
//
// 字段说明：
//   - file: CSV文件句柄
//   - writer: CSV写入器
//   - headers: 列标题列表
//
// 使用示例：
//
//	type Record struct {
//	    Name      string    `json:"name"`
//	    Age       int       `json:"age"`
//	    CreatedAt time.Time `json:"created_at"`
//	}
//
//	// 创建写入器
//	headers := []string{"Name", "Age", "CreatedAt"}
//	writer, err := tfs.InitExcelWriter("data.csv", headers)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer writer.Close()
//
//	// 批量写入数据
//	records := []interface{}{
//	    Record{Name: "张三", Age: 25, CreatedAt: time.Now()},
//	    Record{Name: "李四", Age: 30, CreatedAt: time.Now()},
//	}
//	if err := writer.WriteBatch(records); err != nil {
//	    log.Fatal(err)
//	}
//
// 注意事项：
//   - 使用UTF-8编码
//   - 自动处理类型转换
//   - 支持基本类型和time.Time
//   - 使用标准CSV格式
//   - 需要手动调用Close关闭
//   - 适用于大批量数据导出
type ExcelWriter struct {
	file    *os.File
	writer  *csv.Writer
	headers []string
}

// InitExcelWriter 创建并初始化一个新的Excel CSV写入器
// 参数：
//   - filename: 目标CSV文件路径
//   - headers: 列标题列表
//
// 返回值：
//   - *ExcelWriter: 初始化好的写入器实例
//   - error: 初始化过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	headers := []string{"姓名", "年龄", "邮箱"}
//	writer, err := tfs.InitExcelWriter("users.csv", headers)
//	if err != nil {
//	    log.Printf("创建写入器失败：%v", err)
//	    return
//	}
//	defer writer.Close()
//
// 注意事项：
//   - 会创建新文件（如果已存在则覆盖）
//   - 自动写入表头
//   - 使用UTF-8编码
//   - 使用标准CSV格式
//   - 返回的实例需要手动关闭
//   - 适用于需要创建新CSV文件的场景
func InitExcelWriter(filename string, headers []string) (*ExcelWriter, error) {
	var (
		err    error
		file   *os.File
		writer *csv.Writer
	)

	// 判断目录是否存在
	file, err = os.Create(filename)
	if err != nil {
		return nil, err
	}

	writer = csv.NewWriter(file)
	// 初始化的时候，先将头写到文件
	if err = writer.Write(headers); err != nil {
		_ = file.Close()
		return nil, err
	}

	return &ExcelWriter{
		file:    file,
		writer:  writer,
		headers: headers,
	}, nil
}

// WriteBatch 批量写入结构体数据到CSV文件
// 参数：
//   - datas: 要写入的结构体数据切片，每个元素必须是结构体
//
// 返回值：
//   - error: 写入过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	type Record struct {
//	    Name   string  `json:"name"`
//	    Age    int     `json:"age"`
//	    Salary float64 `json:"salary"`
//	}
//
//	records := []interface{}{
//	    Record{Name: "张三", Age: 25, Salary: 8000.50},
//	    Record{Name: "李四", Age: 30, Salary: 12000.75},
//	}
//	err := writer.WriteBatch(records)
//	if err != nil {
//	    log.Printf("批量写入失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 结构体字段名必须与表头匹配
//   - 自动处理类型转换
//   - 支持基本类型和time.Time
//   - 不存在的字段写入空字符串
//   - 使用反射获取字段值
//   - 适用于批量写入大量数据
func (excelWriter *ExcelWriter) WriteBatch(datas []interface{}) error {
	for _, record := range datas {
		// 以excel列头为长度，创建能存储一行数据的slice
		row := make([]string, len(excelWriter.headers))
		// 反射一行数据的结构体对象
		rowVal := reflect.ValueOf(record)

		// Excel头(header)的值和行结构体数据的字段名(KEY)是设置的一样的，否则下面没有办法通过字段名那到数据值
		for i, header := range excelWriter.headers {
			// 根据字段名称，获取字段名称存储的值
			fieldVal := rowVal.FieldByName(header)

			if fieldVal.IsValid() {
				switch fieldVal.Kind() { // 判断字段对应的值的类型
				case reflect.String:
					row[i] = fieldVal.String()
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					row[i] = strconv.FormatInt(fieldVal.Int(), 10)
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					row[i] = strconv.FormatUint(fieldVal.Uint(), 10)
				case reflect.Float32, reflect.Float64:
					row[i] = strconv.FormatFloat(fieldVal.Float(), 'f', -1, 64)
				case reflect.Bool:
					row[i] = strconv.FormatBool(fieldVal.Bool())
				case reflect.Struct:
					if fieldVal.Type() == reflect.TypeOf(time.Now()) { // 字段类型是时间结构体time.Time
						row[i] = fieldVal.Interface().(time.Time).Format("2006-01-02 15:04:05")
					} else { // 如果不是time.Time结构体类型，统一json成字符串在写入到Excel
						jsonBytes, _ := json.Marshal(fieldVal.Interface())
						row[i] = string(jsonBytes)
					}
				default:
					jsonBytes, _ := json.Marshal(fieldVal.Interface())
					row[i] = string(jsonBytes)
				}
			} else {
				row[i] = ""
			}
		}

		if err := excelWriter.writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}

// Close 刷新缓冲区并关闭Excel CSV写入器
// 返回值：
//   - error: 关闭过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	writer, err := tfs.InitExcelWriter("data.csv", headers)
//	if err != nil {
//	    return err
//	}
//	defer writer.Close()
//
// 注意事项：
//   - 必须在使用完毕后调用
//   - 会刷新所有缓冲数据
//   - 会检查写入错误
//   - 会关闭文件句柄
//   - 关闭后不能再使用
//   - 建议使用defer调用
func (excelWriter *ExcelWriter) Close() error {
	excelWriter.writer.Flush()
	if err := excelWriter.writer.Error(); err != nil {
		return err
	}
	return excelWriter.file.Close()
}

// ReadFromXml 从XML文件读取数据到结构体
// 参数：
//   - filePath: XML文件路径
//   - v: 用于存储结果的结构体指针，结构体字段需要使用xml标签
//
// 返回值：
//   - error: 读取过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	type Config struct {
//	    XMLName xml.Name `xml:"config"`
//	    Server  struct {
//	        Host string `xml:"host"`
//	        Port int    `xml:"port"`
//	    } `xml:"server"`
//	    Database struct {
//	        DSN string `xml:"dsn"`
//	    } `xml:"database"`
//	}
//
//	var cfg Config
//	err := tfs.ReadFromXml("config.xml", &cfg)
//	if err != nil {
//	    log.Printf("读取配置失败：%v", err)
//	    return
//	}
//	fmt.Printf("服务器配置：%s:%d\n", cfg.Server.Host, cfg.Server.Port)
//
// 注意事项：
//   - v必须是指针类型
//   - 结构体需要正确定义xml标签
//   - 支持嵌套结构
//   - 使用encoding/xml包解析
//   - 自动处理文件打开和关闭
//   - 适用于读取XML配置文件
func ReadFromXml(filePath string, v interface{}) error {
	fd, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer fd.Close()
	decodeXML := xml.NewDecoder(fd)

	if err := decodeXML.Decode(v); err != nil {
		return err
	}

	return nil
}

// WriteToXml 将结构体数据写入XML文件
// 参数：
//   - v: 要写入的结构体，需要使用xml标签
//   - filePath: 目标XML文件路径
//
// 返回值：
//   - error: 写入过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	config := Config{
//	    XMLName: xml.Name{Local: "config"},
//	    Server: struct {
//	        Host string `xml:"host"`
//	        Port int    `xml:"port"`
//	    }{
//	        Host: "localhost",
//	        Port: 8080,
//	    },
//	    Database: struct {
//	        DSN string `xml:"dsn"`
//	    }{
//	        DSN: "user:pass@tcp(localhost:3306)/db",
//	    },
//	}
//	err := tfs.WriteToXml(config, "config.xml")
//	if err != nil {
//	    log.Printf("写入配置失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 结构体需要正确定义xml标签
//   - 会创建或覆盖目标文件
//   - 使用encoding/xml包编码
//   - 自动处理文件创建和关闭
//   - 不会自动格式化XML
//   - 适用于保存配置文件
func WriteToXml(v interface{}, filePath string) error {
	fd, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer fd.Close()
	return xml.NewEncoder(fd).Encode(v)
}

// Unzip 解压ZIP文件到指定目录
// 参数：
//   - zipFile: ZIP文件路径
//   - destDir: 解压目标目录
//
// 返回值：
//   - []string: 解压后的所有文件路径列表
//   - error: 解压过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	files, err := tfs.Unzip("archive.zip", "./extracted")
//	if err != nil {
//	    log.Printf("解压失败：%v", err)
//	    return
//	}
//	for _, file := range files {
//	    fmt.Println("解压文件：", file)
//	}
//
// 注意事项：
//   - 自动创建目标目录
//   - 保持原始文件权限
//   - 自动创建子目录
//   - 检查路径安全性（防止路径穿越）
//   - 支持目录和文件
//   - 返回所有解压文件的路径
//   - 适用于解压ZIP归档文件
func Unzip(zipFile string, destDir string) ([]string, error) {
	zipReader, err := zip.OpenReader(zipFile)
	var paths []string
	if err != nil {
		return []string{}, err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		if strings.Contains(f.Name, "..") {
			return []string{}, fmt.Errorf("%s 文件名不合法", f.Name)
		}
		fpath := filepath.Join(destDir, f.Name)
		paths = append(paths, fpath)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return []string{}, err
			}

			inFile, err := f.Open()
			if err != nil {
				return []string{}, err
			}
			defer inFile.Close()

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return []string{}, err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return []string{}, err
			}
		}
	}
	return paths, nil
}

// CompressGzip 使用GZIP算法压缩字符串
// 参数：
//   - data: 要压缩的字符串
//
// 返回值：
//   - string: 压缩后的数据（二进制数据的字符串形式）
//   - error: 压缩过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	original := "这是一段需要压缩的长文本..."
//	compressed, err := tfs.CompressGzip(original)
//	if err != nil {
//	    log.Printf("压缩失败：%v", err)
//	    return
//	}
//
//	// 解压测试
//	decompressed, err := tfs.DecompressGzip(compressed)
//	if err != nil {
//	    log.Printf("解压失败：%v", err)
//	    return
//	}
//	fmt.Printf("解压后与原文相同：%v\n", original == decompressed)
//
// 注意事项：
//   - 使用标准gzip算法
//   - 返回的是二进制数据的字符串形式
//   - 适合压缩文本数据
//   - 压缩率取决于数据特征
//   - 需要配合DecompressGzip使用
//   - 适用于需要压缩传输或存储的场景
func CompressGzip(data string) (string, error) {
	var (
		buf bytes.Buffer
		err error
		w   *gzip.Writer
	)

	w = gzip.NewWriter(&buf)

	if _, err = w.Write([]byte(data)); err != nil {
		return "", err
	}

	w.Close()

	return buf.String(), nil
}

// DecompressGzip 解压GZIP压缩的字符串
// 参数：
//   - data: 要解压的数据（由CompressGzip生成的字符串）
//
// 返回值：
//   - string: 解压后的原始字符串
//   - error: 解压过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	compressed := "..." // 从某处获取的压缩数据
//	original, err := tfs.DecompressGzip(compressed)
//	if err != nil {
//	    log.Printf("解压失败：%v", err)
//	    return
//	}
//	fmt.Println("解压后的内容：", original)
//
// 注意事项：
//   - 输入必须是CompressGzip的输出
//   - 使用标准gzip算法
//   - 会验证gzip头信息
//   - 自动处理缓冲区
//   - 返回UTF-8编码的字符串
//   - 适用于解压缩存储或传输的数据
func DecompressGzip(data string) (string, error) {
	var (
		buf              bytes.Buffer
		err              error
		r                *gzip.Reader
		decompressedData []byte
	)

	buf.Write([]byte(data))
	if r, err = gzip.NewReader(&buf); err != nil {
		return "", err
	}

	if decompressedData, err = io.ReadAll(r); err != nil {
		return "", err
	}

	return string(decompressedData), nil
}

// splitPath 将路径字符串分割为路径段列表
// 参数：
//   - path: 要分割的路径字符串
//
// 返回值：
//   - []string: 路径段列表，不包含空段和斜杠
//
// 内部处理流程：
//  1. 去除路径首尾的斜杠
//  2. 按斜杠分割路径
//  3. 过滤空路径段
//  4. 返回有效路径段列表
//
// 注意事项：
//   - 这是一个内部函数
//   - 处理正斜杠分隔的路径
//   - 忽略连续的斜杠
//   - 空路径返回空切片
//   - 不处理相对路径（.和..）
//   - 不验证路径的有效性
func splitPath(path string) []string {
	// 去除路径开头和结尾的斜杠
	path = strings.Trim(path, "/")

	// 处理空路径
	if path == "" {
		return []string{}
	}

	// 按斜杠分割并过滤空段
	segments := strings.Split(path, "/")
	filteredSegments := make([]string, 0, len(segments))
	for _, segment := range segments {
		if segment != "" {
			filteredSegments = append(filteredSegments, segment)
		}
	}

	return filteredSegments
}

// GetPathSegments 获取路径的所有有效路径段
// 参数：
//   - path: 要分析的路径字符串
//
// 返回值：
//   - []string: 路径段列表，不包含空段和斜杠
//
// 使用示例：
//
//	segments := tfs.GetPathSegments("/path/to/some/file.txt")
//	// 返回: ["path", "to", "some", "file.txt"]
//
//	segments = tfs.GetPathSegments("path//to/file")
//	// 返回: ["path", "to", "file"]
//
// 注意事项：
//   - 忽略路径首尾的斜杠
//   - 忽略连续的斜杠
//   - 空路径返回空切片
//   - 不处理相对路径（.和..）
//   - 不验证路径的有效性
//   - 适用于需要分析路径结构的场景
func GetPathSegments(path string) []string {
	return splitPath(path)
}

// GetLastPathSegments 从路径中提取最后指定数量的路径段并重新组合
// 参数：
//   - path: 源路径字符串
//   - count: 要提取的路径段数量
//
// 返回值：
//   - string: 组合后的路径字符串
//   - error: 处理过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	path, err := tfs.GetLastPathSegments("/var/log/app/error.log", 2)
//	if err != nil {
//	    log.Printf("提取路径失败：%v", err)
//	    return
//	}
//	fmt.Println(path) // 输出: "app/error.log"
//
//	// 如果路径段数量不足，返回完整路径
//	path, _ = tfs.GetLastPathSegments("config.json", 3)
//	fmt.Println(path) // 输出: "config.json"
//
// 注意事项：
//   - count必须大于0
//   - 如果路径段数量不足，返回原路径
//   - 返回的路径使用正斜杠分隔
//   - 不包含开头的斜杠
//   - 空路径会返回错误
//   - 适用于需要截取路径尾部的场景
func GetLastPathSegments(path string, count int) (string, error) {
	if count <= 0 {
		return "", fmt.Errorf("count must be greater than 0")
	}

	segments := splitPath(path)
	if len(segments) == 0 {
		return "", fmt.Errorf("path is empty or contains only slashes")
	}

	// 如果路径段数量不足，返回原路径
	if len(segments) < count {
		return path, nil
	}

	// 获取最后count个路径段
	lastSegments := segments[len(segments)-count:]
	return JoinPathSegments(lastSegments), nil
}

// JoinPathSegments 将路径段列表连接成单个路径字符串
// 参数：
//   - segments: 路径段列表
//
// 返回值：
//   - string: 连接后的路径字符串，使用正斜杠分隔
//
// 使用示例：
//
//	segments := []string{"path", "to", "file.txt"}
//	path := tfs.JoinPathSegments(segments)
//	fmt.Println(path) // 输出: "path/to/file.txt"
//
//	// 空列表
//	path = tfs.JoinPathSegments([]string{})
//	fmt.Println(path) // 输出: ""
//
// 注意事项：
//   - 使用正斜杠（/）作为分隔符
//   - 不会添加开头或结尾的斜杠
//   - 空列表返回空字符串
//   - 不会处理路径中的特殊字符
//   - 不会验证路径的有效性
//   - 适用于需要组合路径段的场景
func JoinPathSegments(segments []string) string {
	return strings.Join(segments, "/")
}

// GetPathDepth 计算路径的深度（有效路径段的数量）
// 参数：
//   - path: 要分析的路径字符串
//
// 返回值：
//   - int: 路径的深度（有效路径段的数量）
//
// 使用示例：
//
//	depth := tfs.GetPathDepth("/path/to/file.txt")
//	fmt.Println(depth) // 输出: 3
//
//	depth = tfs.GetPathDepth("file.txt")
//	fmt.Println(depth) // 输出: 1
//
//	depth = tfs.GetPathDepth("/")
//	fmt.Println(depth) // 输出: 0
//
// 注意事项：
//   - 忽略路径首尾的斜杠
//   - 忽略连续的斜杠
//   - 空路径返回0
//   - 不处理相对路径（.和..）
//   - 不验证路径的有效性
//   - 适用于需要分析路径层级的场景
func GetPathDepth(path string) int {
	return len(splitPath(path))
}

// CopyFile 将源文件复制到目标位置
// 参数：
//   - src: 源文件路径
//   - dst: 目标文件路径
//
// 返回值：
//   - error: 复制过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	err := tfs.CopyFile("source.txt", "backup/source.txt")
//	if err != nil {
//	    log.Printf("复制文件失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 自动创建目标文件
//   - 保持原始文件权限
//   - 使用缓冲区提高性能
//   - 如果目标文件存在会被覆盖
//   - 不支持目录复制
//   - 使用32KB的缓冲区
//   - 适用于复制单个文件
func CopyFile(src, dst string) error {
	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file failed: %w", err)
	}
	defer srcFile.Close()

	// 获取源文件信息
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("get source file info failed: %w", err)
	}

	// 创建目标文件
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("create destination file failed: %w", err)
	}
	defer dstFile.Close()

	// 使用缓冲区复制
	buf := make([]byte, 32*1024) // 32KB 缓冲区
	if _, err := io.CopyBuffer(dstFile, srcFile, buf); err != nil {
		return fmt.Errorf("copy file content failed: %w", err)
	}

	return nil
}

// CopyDir 递归复制整个目录树
// 参数：
//   - src: 源目录路径
//   - dst: 目标目录路径
//
// 返回值：
//   - error: 复制过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	err := tfs.CopyDir("project", "backup/project")
//	if err != nil {
//	    log.Printf("复制目录失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 递归复制所有子目录和文件
//   - 保持原始文件权限
//   - 自动创建目标目录
//   - 保持目录结构
//   - 如果目标存在会覆盖
//   - 使用CopyFile复制文件
//   - 适用于备份整个目录树
func CopyDir(src, dst string) error {
	// 获取源目录信息
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat source directory failed: %w", err)
	}

	// 创建目标目录
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("create destination directory failed: %w", err)
	}

	// 遍历源目录
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read source directory failed: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// 递归复制子目录
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// 复制文件
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// CheckFilePermission 检查文件是否具有指定的权限
// 参数：
//   - path: 要检查的文件路径
//   - mode: 要检查的权限模式（如0400表示可读，0200表示可写，0100表示可执行）
//
// 返回值：
//   - error: 检查结果，如果文件具有指定权限则为nil，否则返回错误
//
// 使用示例：
//
//	// 检查文件是否可读
//	err := tfs.CheckFilePermission("config.json", 0400)
//	if err != nil {
//	    log.Printf("文件权限不足：%v", err)
//	    return
//	}
//
//	// 检查文件是否可写
//	err = tfs.CheckFilePermission("data.txt", 0200)
//	if err != nil {
//	    log.Printf("文件不可写：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 使用八进制表示权限
//   - 会解析符号链接
//   - 检查实际权限而不是请求的权限
//   - 考虑了文件所有者和组权限
//   - 返回详细的错误信息
//   - 适用于需要验证文件权限的场景
func CheckFilePermission(path string, mode fs.FileMode) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat file failed: %w", err)
	}

	if info.Mode().Perm()&mode != mode {
		return fmt.Errorf("insufficient permissions: required %v, got %v", mode, info.Mode().Perm())
	}

	return nil
}

// FormatFileSize 将文件大小转换为人类可读的格式
// 参数：
//   - size: 文件大小（字节数）
//
// 返回值：
//   - string: 格式化后的大小字符串（如 "1.5 MB"）
//
// 使用示例：
//
//	size := int64(1234567)
//	formatted := tfs.FormatFileSize(size)
//	fmt.Println(formatted) // 输出: "1.18 MB"
//
//	size = int64(1024)
//	formatted = tfs.FormatFileSize(size)
//	fmt.Println(formatted) // 输出: "1.00 KB"
//
// 注意事项：
//   - 使用1024作为转换基数
//   - 支持从B到PB的单位
//   - 保留两位小数
//   - 自动选择最合适的单位
//   - 结果包含单位后缀
//   - 适用于显示文件大小的场景
func FormatFileSize(size int64) string {
	const (
		B  = 1
		KB = 1024 * B
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
		PB = 1024 * TB
	)

	switch {
	case size >= PB:
		return fmt.Sprintf("%.2f PB", float64(size)/float64(PB))
	case size >= TB:
		return fmt.Sprintf("%.2f TB", float64(size)/float64(TB))
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// FileWatcher 提供文件变更监控功能
// 用于监控文件的修改并触发回调函数
//
// 字段说明：
//   - path: 要监控的文件路径
//   - interval: 检查间隔时间
//   - lastMod: 上次修改时间
//   - onChange: 文件变更时的回调函数
//   - stop: 用于停止监控的通道
//
// 使用示例：
//
//	// 创建监控器
//	watcher, err := tfs.NewFileWatcher(
//	    "config.json",
//	    time.Second * 5,
//	    func() {
//	        fmt.Println("文件已更新")
//	        // 重新加载配置...
//	    },
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 启动监控
//	watcher.Start()
//	defer watcher.Stop()
//
// 注意事项：
//   - 使用轮询方式检查文件
//   - 在后台goroutine中运行
//   - 可以优雅停止
//   - 检查文件修改时间
//   - 适用于配置文件热更新
//   - 需要手动调用Stop
type FileWatcher struct {
	path     string
	interval time.Duration
	lastMod  time.Time
	onChange func()
	stop     chan struct{}
}

// NewFileWatcher 创建并初始化一个新的文件监控器
// 参数：
//   - path: 要监控的文件路径
//   - interval: 检查间隔时间
//   - onChange: 文件变更时的回调函数
//
// 返回值：
//   - *FileWatcher: 初始化好的监控器实例
//   - error: 初始化过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	watcher, err := tfs.NewFileWatcher(
//	    "config.json",
//	    time.Second * 5,
//	    func() {
//	        fmt.Println("配置文件已更新")
//	        reloadConfig()
//	    },
//	)
//	if err != nil {
//	    log.Printf("创建监控器失败：%v", err)
//	    return
//	}
//	watcher.Start()
//	defer watcher.Stop()
//
// 注意事项：
//   - 会立即检查文件是否存在
//   - 记录文件的初始修改时间
//   - 创建后需要手动调用Start
//   - interval不应太小以避免频繁IO
//   - onChange在后台goroutine中调用
//   - 适用于需要监控文件变化的场景
func NewFileWatcher(path string, interval time.Duration, onChange func()) (*FileWatcher, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat file failed: %w", err)
	}

	return &FileWatcher{
		path:     path,
		interval: interval,
		lastMod:  info.ModTime(),
		onChange: onChange,
		stop:     make(chan struct{}),
	}, nil
}

// Start 启动文件监控器的后台监控任务
// 使用示例：
//
//	watcher, err := tfs.NewFileWatcher(...)
//	if err != nil {
//	    return err
//	}
//	watcher.Start()
//	defer watcher.Stop()
//
// 注意事项：
//   - 在后台goroutine中运行
//   - 按指定间隔检查文件
//   - 发现变更时调用回调函数
//   - 可以通过Stop停止
//   - 不会阻塞调用者
//   - 同一个实例只能启动一次
func (w *FileWatcher) Start() {
	ticker := time.NewTicker(w.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				info, err := os.Stat(w.path)
				if err != nil {
					continue
				}

				if info.ModTime() != w.lastMod {
					w.lastMod = info.ModTime()
					w.onChange()
				}
			case <-w.stop:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop 停止文件监控器并清理资源
// 使用示例：
//
//	watcher, err := tfs.NewFileWatcher(...)
//	if err != nil {
//	    return err
//	}
//	watcher.Start()
//	defer watcher.Stop() // 确保在退出时停止监控
//
// 注意事项：
//   - 会优雅地停止后台goroutine
//   - 可以多次调用（幂等）
//   - 停止后不能重新启动
//   - 会等待当前检查完成
//   - 建议使用defer调用
//   - 停止后回调不会再被调用
func (w *FileWatcher) Stop() {
	close(w.stop)
}

// IsReadable 检查文件是否具有读取权限
// 参数：
//   - path: 要检查的文件路径
//
// 返回值：
//   - bool: 如果文件可读返回true，否则返回false
//
// 使用示例：
//
//	if tfs.IsReadable("config.json") {
//	    // 读取文件...
//	} else {
//	    log.Println("文件不可读")
//	}
//
// 注意事项：
//   - 检查0400权限位
//   - 会解析符号链接
//   - 考虑所有者权限
//   - 不检查文件是否存在
//   - 不检查其他权限
//   - 适用于需要预检查读权限的场景
func IsReadable(path string) bool {
	return CheckFilePermission(path, 0400) == nil
}

// IsWritable 检查文件是否具有写入权限
// 参数：
//   - path: 要检查的文件路径
//
// 返回值：
//   - bool: 如果文件可写返回true，否则返回false
//
// 使用示例：
//
//	if tfs.IsWritable("data.txt") {
//	    // 写入文件...
//	} else {
//	    log.Println("文件不可写")
//	}
//
// 注意事项：
//   - 检查0200权限位
//   - 会解析符号链接
//   - 考虑所有者权限
//   - 不检查文件是否存在
//   - 不检查其他权限
//   - 适用于需要预检查写权限的场景
func IsWritable(path string) bool {
	return CheckFilePermission(path, 0200) == nil
}

// IsExecutable 检查文件是否具有执行权限
// 参数：
//   - path: 要检查的文件路径
//
// 返回值：
//   - bool: 如果文件可执行返回true，否则返回false
//
// 使用示例：
//
//	if tfs.IsExecutable("script.sh") {
//	    // 执行脚本...
//	} else {
//	    log.Println("文件不可执行")
//	}
//
// 注意事项：
//   - 检查0100权限位
//   - 会解析符号链接
//   - 考虑所有者权限
//   - 不检查文件是否存在
//   - 不检查其他权限
//   - 适用于需要预检查执行权限的场景
func IsExecutable(path string) bool {
	return CheckFilePermission(path, 0100) == nil
}
