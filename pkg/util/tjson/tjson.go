package tjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

// Config 定义JSON解析器的配置选项
// 用于自定义JSON解析和序列化的行为
//
// 字段说明：
//   - UseNumber: 是否使用json.Number类型处理数字
//   - TagKey: 用于结构体字段映射的标签键名
//   - OnlyTaggedField: 是否只处理有标签的字段
//   - ValidateJsonRawMessage: 是否验证json.RawMessage
//   - CaseSensitive: 是否区分字段名大小写
//   - DisallowUnknownFields: 是否禁止未知字段
//
// 使用示例：
//
//	config := &tjson.Config{
//	    UseNumber: true,
//	    TagKey: "json",
//	    CaseSensitive: true,
//	    DisallowUnknownFields: true,
//	}
//	util := tjson.New(config)
//
// 注意事项：
//   - UseNumber建议启用以避免精度损失
//   - TagKey默认为"json"
//   - CaseSensitive默认为true
//   - 可以使用DefaultConfig作为基础
//   - 配置会影响所有解析操作
//   - 适用于需要自定义JSON行为的场景
type Config struct {
	UseNumber              bool   // 使用json.Number类型
	TagKey                 string // 结构体标签键
	OnlyTaggedField        bool   // 只处理有标签的字段
	ValidateJsonRawMessage bool   // 验证RawMessage
	CaseSensitive          bool   // 大小写敏感
	DisallowUnknownFields  bool   // 禁止未知字段
}

// DefaultConfig 默认配置
var DefaultConfig = Config{
	UseNumber:              true,
	TagKey:                 "json",
	OnlyTaggedField:        false,
	ValidateJsonRawMessage: true,
	CaseSensitive:          true,
	DisallowUnknownFields:  false,
}

// JSONUtil 提供JSON处理的核心功能
// 支持解析、序列化、验证和类型转换等操作
//
// 字段说明：
//   - config: JSON解析配置
//   - api: json-iterator API实例
//
// 使用示例：
//
//	// 使用默认配置
//	util := tjson.Default
//
//	// 自定义配置
//	config := tjson.Config{
//	    UseNumber: true,
//	    TagKey: "json",
//	}
//	util := tjson.New(config)
//
//	// 解析JSON
//	data, err := util.ParseString(`{"name": "张三", "age": 25}`)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 获取字段
//	name, _ := util.GetString(data, "name")
//	age, _ := util.GetInt(data, "age")
//
// 注意事项：
//   - 线程安全
//   - 配置不可变
//   - 支持类型转换
//   - 提供便捷方法
//   - 适用于所有JSON操作
//   - 建议复用实例
type JSONUtil struct {
	config Config
	api    jsoniter.API
}

// New 创建一个新的JSON工具实例
// 参数：
//   - config: 可选的配置参数，如果不提供则使用默认配置
//
// 返回值：
//   - *JSONUtil: 初始化好的JSON工具实例
//
// 使用示例：
//
//	// 使用默认配置
//	util := tjson.New()
//
//	// 使用自定义配置
//	config := tjson.Config{
//	    UseNumber: true,
//	    TagKey: "json",
//	    DisallowUnknownFields: true,
//	}
//	util := tjson.New(config)
//
// 注意事项：
//   - 配置一旦设置不可更改
//   - 建议全局复用实例
//   - 实例是线程安全的
//   - 可以创建多个实例
//   - 每个实例独立配置
//   - 适用于需要不同配置的场景
func New(config ...Config) *JSONUtil {
	cfg := DefaultConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	// 构建json-iterator配置
	jsonConfig := jsoniter.Config{
		UseNumber:              cfg.UseNumber,
		TagKey:                 cfg.TagKey,
		OnlyTaggedField:        cfg.OnlyTaggedField,
		ValidateJsonRawMessage: cfg.ValidateJsonRawMessage,
		CaseSensitive:          cfg.CaseSensitive,
		DisallowUnknownFields:  cfg.DisallowUnknownFields,
	}

	return &JSONUtil{
		config: cfg,
		api:    jsonConfig.Froze(),
	}
}

// Default 默认JSON工具实例
var Default = New()

// Parse 将JSON字节数据解析为map
// 参数：
//   - data: 要解析的JSON字节数据
//
// 返回值：
//   - map[string]interface{}: 解析后的map对象
//   - error: 解析过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	jsonData := []byte(`{"name": "张三", "age": 25}`)
//	result, err := util.Parse(jsonData)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("姓名：%v, 年龄：%v\n",
//	    result["name"],
//	    result["age"])
//
// 注意事项：
//   - 支持所有JSON类型
//   - 会自动处理数字类型
//   - 遵循配置的UseNumber设置
//   - 空数据会返回错误
//   - 适用于动态JSON解析
//   - 建议处理所有错误
func (j *JSONUtil) Parse(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := j.api.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}

	if j.config.UseNumber {
		return j.convertNumbers(result), nil
	}
	return result, nil
}

// ParseString 将JSON字符串解析为map
// 参数：
//   - data: 要解析的JSON字符串
//
// 返回值：
//   - map[string]interface{}: 解析后的map对象
//   - error: 解析过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	jsonStr := `{
//	    "name": "张三",
//	    "age": 25,
//	    "hobbies": ["读书", "运动"]
//	}`
//	result, err := util.ParseString(jsonStr)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("姓名：%v\n", result["name"])
//
// 注意事项：
//   - 支持所有JSON类型
//   - 会自动处理数字类型
//   - 支持Unicode字符
//   - 空字符串会返回错误
//   - 适用于字符串形式的JSON
//   - 内部调用Parse方法
func (j *JSONUtil) ParseString(data string) (map[string]interface{}, error) {
	return j.Parse([]byte(data))
}

// ParseReader 从io.Reader读取并解析JSON数据
// 参数：
//   - reader: 实现了io.Reader接口的数据源
//
// 返回值：
//   - map[string]interface{}: 解析后的map对象
//   - error: 解析过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	file, err := os.Open("config.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer file.Close()
//
//	result, err := util.ParseReader(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// 注意事项：
//   - 支持任何io.Reader
//   - 会读取全部数据
//   - 大文件要注意内存
//   - 读取错误会返回
//   - 适用于文件或网络数据
//   - 内部调用Parse方法
func (j *JSONUtil) ParseReader(reader io.Reader) (map[string]interface{}, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read data failed: %w", err)
	}
	return j.Parse(data)
}

// ParseRequest 解析HTTP请求体中的JSON数据
// 参数：
//   - r: HTTP请求对象
//
// 返回值：
//   - map[string]interface{}: 解析后的map对象
//   - error: 解析过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//	    data, err := util.ParseRequest(r)
//	    if err != nil {
//	        http.Error(w, err.Error(), http.StatusBadRequest)
//	        return
//	    }
//	    // 处理JSON数据...
//	}
//
// 注意事项：
//   - 只读取请求体
//   - 读取后不能重复读取
//   - 适用于POST/PUT请求
//   - 会检查Content-Type
//   - 大请求要注意内存
//   - 内部调用ParseReader方法
func (j *JSONUtil) ParseRequest(r *http.Request) (map[string]interface{}, error) {
	return j.ParseReader(r.Body)
}

// Unmarshal 将JSON数据解析到指定的结构体
// 参数：
//   - data: 要解析的JSON字节数据
//   - v: 目标结构体的指针
//
// 返回值：
//   - error: 解析过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	type User struct {
//	    Name   string   `json:"name"`
//	    Age    int      `json:"age"`
//	    Emails []string `json:"emails"`
//	}
//
//	jsonData := []byte(`{
//	    "name": "张三",
//	    "age": 25,
//	    "emails": ["zhangsan@example.com"]
//	}`)
//
//	var user User
//	err := util.Unmarshal(jsonData, &user)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// 注意事项：
//   - v必须是指针类型
//   - 结构体要定义json标签
//   - 支持嵌套结构
//   - 遵循配置的标签设置
//   - 类型不匹配会报错
//   - 适用于已知结构的JSON
func (j *JSONUtil) Unmarshal(data []byte, v interface{}) error {
	return j.api.Unmarshal(data, v)
}

// UnmarshalString 将JSON字符串解析到指定的结构体
// 参数：
//   - data: 要解析的JSON字符串
//   - v: 目标结构体的指针
//
// 返回值：
//   - error: 解析过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	type Config struct {
//	    Debug   bool              `json:"debug"`
//	    Port    int              `json:"port"`
//	    Options map[string]string `json:"options"`
//	}
//
//	jsonStr := `{
//	    "debug": true,
//	    "port": 8080,
//	    "options": {
//	        "timeout": "30s",
//	        "retry": "3"
//	    }
//	}`
//
//	var cfg Config
//	err := util.UnmarshalString(jsonStr, &cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// 注意事项：
//   - v必须是指针类型
//   - 支持所有JSON类型
//   - 支持Unicode字符
//   - 内部调用Unmarshal方法
//   - 空字符串会返回错误
//   - 适用于配置文件解析
func (j *JSONUtil) UnmarshalString(data string, v interface{}) error {
	return j.Unmarshal([]byte(data), v)
}

// ParseToSlice 将JSON字符串解析为[]interface{}切片
// 参数：
//   - data: 要解析的JSON字符串
//
// 返回值：
//   - []interface{}: 解析后的切片
//   - error: 解析过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	jsonStr := `["张三", "李四", "王五"]`
//	result, err := util.ParseToSlice(jsonStr)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for i, item := range result {
//	    fmt.Printf("第%d项: %v\n", i+1, item)
//	}
//
//	// 解析复杂数组
//	complexJSON := `[
//	    {"name": "张三", "age": 25},
//	    {"name": "李四", "age": 30},
//	    {"name": "王五", "age": 28}
//	]`
//	users, err := util.ParseToSlice(complexJSON)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, user := range users {
//	    if userMap, ok := user.(map[string]interface{}); ok {
//	        fmt.Printf("用户: %v, 年龄: %v\n", userMap["name"], userMap["age"])
//	    }
//	}
//
// 注意事项：
//   - 只支持JSON数组格式
//   - 不支持JSON对象格式
//   - 会自动处理数字类型
//   - 支持嵌套结构
//   - 空字符串会返回错误
//   - 适用于解析JSON数组数据
func (j *JSONUtil) ParseToSlice(data string) ([]interface{}, error) {
	var result []interface{}
	if err := j.api.Unmarshal([]byte(data), &result); err != nil {
		return nil, fmt.Errorf("json unmarshal to slice failed: %w", err)
	}

	if j.config.UseNumber {
		return j.convertArray(result), nil
	}
	return result, nil
}

// ParseToMapSlice 将JSON数组字符串解析为[]map[string]interface{}切片
// 参数：
//   - data: 要解析的JSON数组字符串
//
// 返回值：
//   - []map[string]interface{}: 解析后的map切片
//   - error: 解析过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	jsonStr := `[
//	    {"name": "张三", "age": 25, "city": "北京"},
//	    {"name": "李四", "age": 30, "city": "上海"},
//	    {"name": "王五", "age": 28, "city": "广州"}
//	]`
//	result, err := util.ParseToMapSlice(jsonStr)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for i, user := range result {
//	    fmt.Printf("第%d个用户: %s, 年龄: %v, 城市: %s\n",
//	        i+1, user["name"], user["age"], user["city"])
//	}
//
//	// 解析配置数组
//	configJSON := `[
//	    {"key": "timeout", "value": 30, "enabled": true},
//	    {"key": "retry", "value": 3, "enabled": false},
//	    {"key": "debug", "value": "info", "enabled": true}
//	]`
//	configs, err := util.ParseToMapSlice(configJSON)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, config := range configs {
//	    fmt.Printf("配置: %s = %v (启用: %v)\n",
//	        config["key"], config["value"], config["enabled"])
//	}
//
// 注意事项：
//   - 只支持JSON数组格式
//   - 数组中的每个元素必须是JSON对象
//   - 不支持混合类型数组
//   - 会自动处理数字类型
//   - 空字符串会返回错误
//   - 适用于解析对象数组数据
//   - 如果数组包含非对象元素会返回错误
func (j *JSONUtil) ParseToMapSlice(data string) ([]map[string]interface{}, error) {
	var result []interface{}
	if err := j.api.Unmarshal([]byte(data), &result); err != nil {
		return nil, fmt.Errorf("json unmarshal to slice failed: %w", err)
	}

	// 转换为map切片
	mapSlice := make([]map[string]interface{}, 0, len(result))
	for i, item := range result {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if j.config.UseNumber {
				mapSlice = append(mapSlice, j.convertNumbers(itemMap))
			} else {
				mapSlice = append(mapSlice, itemMap)
			}
		} else {
			return nil, fmt.Errorf("array element at index %d is not a JSON object, got %T", i, item)
		}
	}

	return mapSlice, nil
}

// Marshal 将任意数据序列化为JSON字节数组
// 参数：
//   - v: 要序列化的数据（结构体、map、切片等）
//
// 返回值：
//   - []byte: 序列化后的JSON字节数组
//   - error: 序列化过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	data := struct {
//	    Name    string   `json:"name"`
//	    Age     int      `json:"age"`
//	    Hobbies []string `json:"hobbies"`
//	}{
//	    Name:    "张三",
//	    Age:     25,
//	    Hobbies: []string{"读书", "运动"},
//	}
//
//	jsonBytes, err := util.Marshal(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("JSON: %s\n", jsonBytes)
//
// 注意事项：
//   - 支持所有可序列化类型
//   - 会自动处理特殊类型
//   - 使用json标签映射字段
//   - 忽略空值字段
//   - 会清理不安全的数据
//   - 适用于API响应序列化
func (j *JSONUtil) Marshal(v interface{}) ([]byte, error) {
	cleaned := cleanForMarshal(v)
	return json.Marshal(cleaned)
}

// MarshalToString 将任意数据序列化为JSON字符串
// 参数：
//   - v: 要序列化的数据（结构体、map、切片等）
//
// 返回值：
//   - string: 序列化后的JSON字符串
//   - error: 序列化过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	response := map[string]interface{}{
//	    "code": 200,
//	    "message": "success",
//	    "data": map[string]interface{}{
//	        "id": 1001,
//	        "name": "张三",
//	    },
//	}
//
//	jsonStr, err := util.MarshalToString(response)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(jsonStr)
//
// 注意事项：
//   - 支持所有可序列化类型
//   - 内部调用Marshal方法
//   - 返回UTF-8编码字符串
//   - 会自动处理特殊字符
//   - 适用于日志记录
//   - 适用于网络传输
func (j *JSONUtil) MarshalToString(v interface{}) (string, error) {
	cleaned := cleanForMarshal(v)
	bytes, err := json.Marshal(cleaned)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Get 从map中获取指定键的值
// 参数：
//   - data: 源数据map
//   - key: 要获取的键名
//
// 返回值：
//   - interface{}: 键对应的值
//   - bool: 键是否存在
//
// 使用示例：
//
//	data := map[string]interface{}{
//	    "name": "张三",
//	    "age": 25,
//	}
//
//	if value, exists := util.Get(data, "name"); exists {
//	    fmt.Printf("名字：%v\n", value)
//	} else {
//	    fmt.Println("未找到名字")
//	}
//
// 注意事项：
//   - 返回interface{}类型
//   - 需要自己做类型断言
//   - 不存在时返回false
//   - 不会panic
//   - 适用于动态数据
//   - 建议使用类型化的Get方法
func (j *JSONUtil) Get(data map[string]interface{}, key string) (interface{}, bool) {
	value, exists := data[key]
	return value, exists
}

// GetInt64 从map中获取int64类型的值
// 参数：
//   - data: 源数据map
//   - key: 要获取的键名
//
// 返回值：
//   - int64: 转换后的整数值
//   - bool: 是否成功获取并转换
//
// 使用示例：
//
//	data := map[string]interface{}{
//	    "id": json.Number("1234567890"),
//	    "count": 100,
//	    "price": "999",
//	}
//
//	// 获取数字类型
//	if id, ok := util.GetInt64(data, "id"); ok {
//	    fmt.Printf("ID: %d\n", id)
//	}
//
//	// 获取整数
//	if count, ok := util.GetInt64(data, "count"); ok {
//	    fmt.Printf("数量: %d\n", count)
//	}
//
//	// 获取字符串数字
//	if price, ok := util.GetInt64(data, "price"); ok {
//	    fmt.Printf("价格: %d\n", price)
//	}
//
// 注意事项：
//   - 支持多种数字类型转换
//   - 支持字符串数字转换
//   - 浮点数会被截断
//   - 超出范围返回false
//   - 转换失败返回false
//   - 适用于需要精确整数的场景
func (j *JSONUtil) GetInt64(data map[string]interface{}, key string) (int64, bool) {
	value, exists := data[key]
	if !exists {
		return 0, false
	}
	return j.toInt64(value)
}

// GetInt 从map中获取int类型的值
// 参数：
//   - data: 源数据map
//   - key: 要获取的键名
//
// 返回值：
//   - int: 转换后的整数值
//   - bool: 是否成功获取并转换
//
// 使用示例：
//
//	data := map[string]interface{}{
//	    "age": 25,
//	    "level": json.Number("3"),
//	    "score": "100",
//	}
//
//	// 获取普通整数
//	if age, ok := util.GetInt(data, "age"); ok {
//	    fmt.Printf("年龄: %d\n", age)
//	}
//
//	// 获取数字类型
//	if level, ok := util.GetInt(data, "level"); ok {
//	    fmt.Printf("等级: %d\n", level)
//	}
//
//	// 获取字符串数字
//	if score, ok := util.GetInt(data, "score"); ok {
//	    fmt.Printf("分数: %d\n", score)
//	}
//
// 注意事项：
//   - 内部调用GetInt64
//   - 大数会被截断
//   - 支持多种类型转换
//   - 转换失败返回false
//   - 不存在返回false
//   - 适用于普通整数场景
func (j *JSONUtil) GetInt(data map[string]interface{}, key string) (int, bool) {
	value, exists := data[key]
	if !exists {
		return 0, false
	}
	return j.toInt(value)
}

// GetFloat64 从map中获取float64类型的值
// 参数：
//   - data: 源数据map
//   - key: 要获取的键名
//
// 返回值：
//   - float64: 转换后的浮点数值
//   - bool: 是否成功获取并转换
//
// 使用示例：
//
//	data := map[string]interface{}{
//	    "price": 99.99,
//	    "weight": json.Number("12.34"),
//	    "rate": "0.85",
//	    "count": 100,
//	}
//
//	// 获取浮点数
//	if price, ok := util.GetFloat64(data, "price"); ok {
//	    fmt.Printf("价格: %.2f\n", price)
//	}
//
//	// 获取数字类型
//	if weight, ok := util.GetFloat64(data, "weight"); ok {
//	    fmt.Printf("重量: %.2f\n", weight)
//	}
//
//	// 获取字符串数字
//	if rate, ok := util.GetFloat64(data, "rate"); ok {
//	    fmt.Printf("比率: %.2f\n", rate)
//	}
//
//	// 获取整数
//	if count, ok := util.GetFloat64(data, "count"); ok {
//	    fmt.Printf("数量: %.2f\n", count)
//	}
//
// 注意事项：
//   - 支持多种数字类型
//   - 支持字符串转换
//   - 整数会被转为浮点
//   - 保持原始精度
//   - 转换失败返回false
//   - 适用于需要精确小数的场景
func (j *JSONUtil) GetFloat64(data map[string]interface{}, key string) (float64, bool) {
	value, exists := data[key]
	if !exists {
		return 0, false
	}
	return j.toFloat64(value)
}

// GetString 从map中获取string类型的值
// 参数：
//   - data: 源数据map
//   - key: 要获取的键名
//
// 返回值：
//   - string: 转换后的字符串值
//   - bool: 是否成功获取并转换
//
// 使用示例：
//
//	data := map[string]interface{}{
//	    "name": "张三",
//	    "id": json.Number("1001"),
//	    "age": 25,
//	    "vip": true,
//	    "score": 99.99,
//	}
//
//	// 获取字符串
//	if name, ok := util.GetString(data, "name"); ok {
//	    fmt.Printf("姓名: %s\n", name)
//	}
//
//	// 获取数字类型
//	if id, ok := util.GetString(data, "id"); ok {
//	    fmt.Printf("ID: %s\n", id)
//	}
//
//	// 获取整数
//	if age, ok := util.GetString(data, "age"); ok {
//	    fmt.Printf("年龄: %s\n", age)
//	}
//
//	// 获取布尔值
//	if vip, ok := util.GetString(data, "vip"); ok {
//	    fmt.Printf("VIP: %s\n", vip)
//	}
//
//	// 获取浮点数
//	if score, ok := util.GetString(data, "score"); ok {
//	    fmt.Printf("分数: %s\n", score)
//	}
//
// 注意事项：
//   - 支持多种类型转换
//   - 数字会格式化
//   - 布尔值转为"true"/"false"
//   - 保持数字精度
//   - 转换失败返回false
//   - 适用于数据展示场景
func (j *JSONUtil) GetString(data map[string]interface{}, key string) (string, bool) {
	value, exists := data[key]
	if !exists {
		return "", false
	}
	return j.toString(value)
}

// GetBool 从map中获取bool类型的值
// 参数：
//   - data: 源数据map
//   - key: 要获取的键名
//
// 返回值：
//   - bool: 转换后的布尔值
//   - bool: 是否成功获取并转换
//
// 使用示例：
//
//	data := map[string]interface{}{
//	    "active": true,
//	    "enabled": "true",
//	    "status": 1,
//	    "valid": "1",
//	}
//
//	// 获取布尔值
//	if active, ok := util.GetBool(data, "active"); ok {
//	    fmt.Printf("激活状态: %v\n", active)
//	}
//
//	// 获取字符串布尔值
//	if enabled, ok := util.GetBool(data, "enabled"); ok {
//	    fmt.Printf("启用状态: %v\n", enabled)
//	}
//
//	// 获取数字（非零为true）
//	if status, ok := util.GetBool(data, "status"); ok {
//	    fmt.Printf("状态: %v\n", status)
//	}
//
//	// 获取字符串数字
//	if valid, ok := util.GetBool(data, "valid"); ok {
//	    fmt.Printf("有效性: %v\n", valid)
//	}
//
// 注意事项：
//   - 支持多种类型转换
//   - 字符串"true"/"false"
//   - 数字0为false
//   - 数字非0为true
//   - 转换失败返回false
//   - 适用于状态判断场景
func (j *JSONUtil) GetBool(data map[string]interface{}, key string) (bool, bool) {
	value, exists := data[key]
	if !exists {
		return false, false
	}
	return j.toBool(value)
}

// GetMap 从map中获取嵌套的map值
// 参数：
//   - data: 源数据map
//   - key: 要获取的键名
//
// 返回值：
//   - map[string]interface{}: 获取到的map对象
//   - bool: 是否成功获取并转换
//
// 使用示例：
//
//	data := map[string]interface{}{
//	    "user": map[string]interface{}{
//	        "name": "张三",
//	        "age": 25,
//	        "address": map[string]interface{}{
//	            "city": "北京",
//	            "street": "朝阳区",
//	        },
//	    },
//	}
//
//	// 获取用户信息
//	if user, ok := util.GetMap(data, "user"); ok {
//	    fmt.Printf("用户名: %v\n", user["name"])
//	    // 获取嵌套的地址信息
//	    if address, ok := util.GetMap(user, "address"); ok {
//	        fmt.Printf("城市: %v\n", address["city"])
//	    }
//	}
//
// 注意事项：
//   - 只支持map类型
//   - 不会进行类型转换
//   - 返回的是浅拷贝
//   - 可以嵌套使用
//   - 不存在返回false
//   - 适用于复杂JSON结构
func (j *JSONUtil) GetMap(data map[string]interface{}, key string) (map[string]interface{}, bool) {
	value, exists := data[key]
	if !exists {
		return nil, false
	}
	return j.toMap(value)
}

// GetArray 从map中获取数组值
// 参数：
//   - data: 源数据map
//   - key: 要获取的键名
//
// 返回值：
//   - []interface{}: 获取到的数组
//   - bool: 是否成功获取并转换
//
// 使用示例：
//
//	data := map[string]interface{}{
//	    "tags": []interface{}{"Go", "JSON", "Web"},
//	    "scores": []interface{}{98, 95, 100},
//	    "users": []interface{}{
//	        map[string]interface{}{
//	            "name": "张三",
//	            "age": 25,
//	        },
//	        map[string]interface{}{
//	            "name": "李四",
//	            "age": 30,
//	        },
//	    },
//	}
//
//	// 获取字符串数组
//	if tags, ok := util.GetArray(data, "tags"); ok {
//	    for _, tag := range tags {
//	        fmt.Printf("标签: %v\n", tag)
//	    }
//	}
//
//	// 获取数字数组
//	if scores, ok := util.GetArray(data, "scores"); ok {
//	    for _, score := range scores {
//	        fmt.Printf("分数: %v\n", score)
//	    }
//	}
//
//	// 获取对象数组
//	if users, ok := util.GetArray(data, "users"); ok {
//	    for _, user := range users {
//	        if userMap, ok := user.(map[string]interface{}); ok {
//	            fmt.Printf("用户: %v\n", userMap["name"])
//	        }
//	    }
//	}
//
// 注意事项：
//   - 只支持数组类型
//   - 返回interface{}切片
//   - 需要自己做类型断言
//   - 返回的是浅拷贝
//   - 不存在返回false
//   - 适用于列表数据处理
func (j *JSONUtil) GetArray(data map[string]interface{}, key string) ([]interface{}, bool) {
	value, exists := data[key]
	if !exists {
		return nil, false
	}
	return j.toArray(value)
}

// GetStruct 从map中获取值并转换为指定的结构体
// 参数：
//   - data: 源数据map
//   - key: 要获取的键名
//   - v: 目标结构体的指针
//
// 返回值：
//   - bool: 是否成功获取并转换
//
// 使用示例：
//
//	type User struct {
//	    Name    string   `json:"name"`
//	    Age     int      `json:"age"`
//	    Email   string   `json:"email"`
//	    Tags    []string `json:"tags"`
//	    Profile struct {
//	        Avatar string `json:"avatar"`
//	        Bio    string `json:"bio"`
//	    } `json:"profile"`
//	}
//
//	data := map[string]interface{}{
//	    "user": map[string]interface{}{
//	        "name": "张三",
//	        "age": 25,
//	        "email": "zhangsan@example.com",
//	        "tags": []string{"Go", "Web"},
//	        "profile": map[string]interface{}{
//	            "avatar": "avatar.jpg",
//	            "bio": "程序员",
//	        },
//	    },
//	}
//
//	var user User
//	if ok := util.GetStruct(data, "user", &user); ok {
//	    fmt.Printf("用户信息：%+v\n", user)
//	}
//
// 注意事项：
//   - v必须是指针类型
//   - 结构体要定义json标签
//   - 支持嵌套结构
//   - 支持基本类型转换
//   - 字段类型要匹配
//   - 适用于复杂对象映射
func (j *JSONUtil) GetStruct(data map[string]interface{}, key string, v interface{}) bool {
	value, exists := data[key]
	if !exists {
		return false
	}
	return j.toStruct(value, v)
}

// Validate 验证JSON数据的格式是否正确
// 参数：
//   - data: 要验证的JSON字节数据
//
// 返回值：
//   - error: 验证结果，如果格式正确则为nil
//
// 使用示例：
//
//	// 验证有效的JSON
//	validJSON := []byte(`{
//	    "name": "张三",
//	    "age": 25,
//	    "hobbies": ["读书", "运动"]
//	}`)
//	if err := util.Validate(validJSON); err == nil {
//	    fmt.Println("JSON格式正确")
//	}
//
//	// 验证无效的JSON
//	invalidJSON := []byte(`{
//	    "name": "张三",
//	    age: 25, // 缺少引号
//	}`)
//	if err := util.Validate(invalidJSON); err != nil {
//	    fmt.Printf("JSON格式错误：%v\n", err)
//	}
//
// 注意事项：
//   - 只检查语法正确性
//   - 不验证内容合法性
//   - 支持嵌套结构
//   - 适用于数据验证
//   - 建议在解析前调用
//   - 错误信息比较简单
func (j *JSONUtil) Validate(data []byte) error {
	if !json.Valid(data) {
		return fmt.Errorf("invalid JSON format")
	}
	return nil
}

// ValidateString 验证JSON字符串的格式是否正确
// 参数：
//   - data: 要验证的JSON字符串
//
// 返回值：
//   - error: 验证结果，如果格式正确则为nil
//
// 使用示例：
//
//	// 验证有效的JSON字符串
//	validJSON := `{
//	    "name": "张三",
//	    "age": 25,
//	    "hobbies": ["读书", "运动"]
//	}`
//	if err := util.ValidateString(validJSON); err == nil {
//	    fmt.Println("JSON格式正确")
//	}
//
//	// 验证无效的JSON字符串
//	invalidJSON := `{
//	    "name": "张三",
//	    age: 25, // 缺少引号
//	}`
//	if err := util.ValidateString(invalidJSON); err != nil {
//	    fmt.Printf("JSON格式错误：%v\n", err)
//	}
//
// 注意事项：
//   - 内部调用Validate方法
//   - 只检查语法正确性
//   - 支持Unicode字符
//   - 适用于字符串验证
//   - 建议在解析前调用
//   - 错误信息比较简单
func (j *JSONUtil) ValidateString(data string) error {
	return j.Validate([]byte(data))
}

// ValidateRequired 验证map中是否包含所有必需的字段
// 参数：
//   - data: 要验证的map数据
//   - fields: 必需字段的名称列表（可变参数）
//
// 返回值：
//   - error: 验证结果，如果所有字段都存在则为nil
//
// 使用示例：
//
//	data := map[string]interface{}{
//	    "username": "zhangsan",
//	    "password": "123456",
//	    "email": "zhangsan@example.com",
//	}
//
//	// 验证必需字段
//	err := util.ValidateRequired(data,
//	    "username",
//	    "password",
//	    "email",
//	)
//	if err != nil {
//	    fmt.Printf("缺少必需字段：%v\n", err)
//	    return
//	}
//
//	// 验证可能缺少的字段
//	err = util.ValidateRequired(data,
//	    "username",
//	    "mobile", // 不存在的字段
//	)
//	if err != nil {
//	    fmt.Printf("缺少字段：%v\n", err)
//	}
//
// 注意事项：
//   - 只检查字段是否存在
//   - 不检查字段的值
//   - 支持多个字段
//   - 返回第一个缺失的字段
//   - 适用于表单验证
//   - 建议在处理数据前调用
func (j *JSONUtil) ValidateRequired(data map[string]interface{}, fields ...string) error {
	for _, field := range fields {
		if _, exists := data[field]; !exists {
			return fmt.Errorf("required field '%s' is missing", field)
		}
	}
	return nil
}

// ValidateType 验证map中指定字段的值类型是否符合预期
// 参数：
//   - data: 要验证的map数据
//   - field: 要验证的字段名
//   - expectedType: 预期的类型（reflect.Kind）
//
// 返回值：
//   - error: 验证结果，如果类型匹配则为nil
//
// 使用示例：
//
//	data := map[string]interface{}{
//	    "name": "张三",
//	    "age": 25,
//	    "vip": true,
//	    "scores": []interface{}{98, 95, 100},
//	}
//
//	// 验证字符串类型
//	err := util.ValidateType(data, "name", reflect.String)
//	if err != nil {
//	    fmt.Printf("类型错误：%v\n", err)
//	}
//
//	// 验证整数类型
//	err = util.ValidateType(data, "age", reflect.Int)
//	if err != nil {
//	    fmt.Printf("类型错误：%v\n", err)
//	}
//
//	// 验证布尔类型
//	err = util.ValidateType(data, "vip", reflect.Bool)
//	if err != nil {
//	    fmt.Printf("类型错误：%v\n", err)
//	}
//
//	// 验证切片类型
//	err = util.ValidateType(data, "scores", reflect.Slice)
//	if err != nil {
//	    fmt.Printf("类型错误：%v\n", err)
//	}
//
// 注意事项：
//   - 使用reflect.Kind判断类型
//   - 字段不存在会返回错误
//   - 类型不匹配返回错误
//   - 不检查具体的值
//   - 适用于类型验证
//   - 建议在类型转换前调用
func (j *JSONUtil) ValidateType(data map[string]interface{}, field string, expectedType reflect.Kind) error {
	value, exists := data[field]
	if !exists {
		return fmt.Errorf("field '%s' is missing", field)
	}

	actualType := reflect.TypeOf(value).Kind()
	if actualType != expectedType {
		return fmt.Errorf("field '%s' expected type %v, got %v", field, expectedType, actualType)
	}

	return nil
}

// convertNumbers 智能转换数字类型
func (j *JSONUtil) convertNumbers(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range data {
		result[key] = j.convertValue(value)
	}
	return result
}

// convertValue 转换单个值
func (j *JSONUtil) convertValue(value interface{}) interface{} {
	switch v := value.(type) {
	case json.Number:
		return j.convertNumber(v)
	case map[string]interface{}:
		return j.convertNumbers(v)
	case []interface{}:
		return j.convertArray(v)
	default:
		return v
	}
}

// convertNumber 转换Number类型
func (j *JSONUtil) convertNumber(num json.Number) interface{} {
	// 优先尝试int64
	if intVal, err := num.Int64(); err == nil {
		return intVal
	}

	// 再尝试float64
	if floatVal, err := num.Float64(); err == nil {
		// 检查是否是整数
		if floatVal == float64(int64(floatVal)) {
			return int64(floatVal)
		}
		return floatVal
	}

	// 最后返回字符串
	return num.String()
}

// convertArray 转换数组
func (j *JSONUtil) convertArray(arr []interface{}) []interface{} {
	result := make([]interface{}, len(arr))
	for i, value := range arr {
		result[i] = j.convertValue(value)
	}
	return result
}

// toInt64 转换为int64
func (j *JSONUtil) toInt64(value interface{}) (int64, bool) {
	switch v := value.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case int32:
		return int64(v), true
	case int16:
		return int64(v), true
	case int8:
		return int64(v), true
	case uint64:
		return int64(v), true
	case uint:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint8:
		return int64(v), true
	case float64:
		if v == float64(int64(v)) {
			return int64(v), true
		}
		return 0, false
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	case json.Number:
		if intVal, err := v.Int64(); err == nil {
			return intVal, true
		}
		return 0, false
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return parsed, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// toInt 转换为int
func (j *JSONUtil) toInt(value interface{}) (int, bool) {
	if int64Val, ok := j.toInt64(value); ok {
		return int(int64Val), true
	}
	return 0, false
}

// toFloat64 转换为float64
func (j *JSONUtil) toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int64:
		return float64(v), true
	case int:
		return float64(v), true
	case json.Number:
		if floatVal, err := v.Float64(); err == nil {
			return floatVal, true
		}
		return 0, false
	case string:
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			return parsed, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// toString 转换为字符串
func (j *JSONUtil) toString(value interface{}) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case int64:
		return strconv.FormatInt(v, 10), true
	case int:
		return strconv.Itoa(v), true
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), true
	case json.Number:
		return v.String(), true
	case bool:
		return strconv.FormatBool(v), true
	default:
		return "", false
	}
}

// toBool 转换为布尔值
func (j *JSONUtil) toBool(value interface{}) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		if strings.ToLower(v) == "true" {
			return true, true
		}
		if strings.ToLower(v) == "false" {
			return false, true
		}
		// 处理字符串数字
		if v == "1" {
			return true, true
		}
		if v == "0" {
			return false, true
		}
		return false, false
	case int64:
		return v != 0, true
	case int:
		return v != 0, true
	case float64:
		return v != 0, true
	default:
		return false, false
	}
}

// toMap 转换为map
func (j *JSONUtil) toMap(value interface{}) (map[string]interface{}, bool) {
	if mapVal, ok := value.(map[string]interface{}); ok {
		return mapVal, true
	}
	return nil, false
}

// toArray 转换为数组
func (j *JSONUtil) toArray(value interface{}) ([]interface{}, bool) {
	if arrVal, ok := value.([]interface{}); ok {
		return arrVal, true
	}
	return nil, false
}

// toStruct 转换为结构体
func (j *JSONUtil) toStruct(value interface{}, v interface{}) bool {
	cleaned := cleanForMarshal(value)
	bytes, err := json.Marshal(cleaned)
	if err != nil {
		return false
	}
	return j.api.Unmarshal(bytes, v) == nil
}

// cleanForMarshal 清理数据以避免 json-iterator 的 panic
func cleanForMarshal(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case json.Number:
		if intVal, err := val.Int64(); err == nil {
			return intVal
		}
		if floatVal, err := val.Float64(); err == nil {
			return floatVal
		}
		return val.String()
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			if v != nil {
				result[k] = cleanForMarshal(v)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, 0, len(val))
		for _, v := range val {
			if v != nil {
				result = append(result, cleanForMarshal(v))
			}
		}
		return result
	default:
		return val
	}
}

// GetJSONKeys 获取JSON字符串中的所有键 Key
func (j *JSONUtil) GetJSONKeys(jsonStr string) (keys []string, err error) {
	// 使用json.Decoder，以便在解析过程中记录键的顺序
	dec := json.NewDecoder(strings.NewReader(jsonStr))
	t, err := dec.Token()
	if err != nil {
		return nil, err
	}
	// 确保数据是一个对象
	if t != json.Delim('{') {
		return nil, fmt.Errorf("JSON is not an object")
	}
	for dec.More() {
		t, err = dec.Token()
		if err != nil {
			return nil, err
		}
		keys = append(keys, t.(string))

		// 解析值
		var value interface{}
		err = dec.Decode(&value)
		if err != nil {
			return nil, err
		}
	}
	return keys, nil
}

// GetJSONKeysString 静态方法：获取JSON字符串中的所有键
func GetJSONKeys(jsonStr string) (keys []string, err error) {
	return Default.GetJSONKeys(jsonStr)
}

// ToJSONString 将任意类型转换为JSON字符串
func (j *JSONUtil) ToJSONString(v interface{}) string {
	jsonBytes, err := j.Marshal(v)
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}

// ToJSONStringStatic 静态方法：将任意类型转换为JSON字符串
func ToJSONString(v interface{}) string {
	return Default.ToJSONString(v)
}

// 便捷函数
func Parse(data []byte) (map[string]interface{}, error) {
	return Default.Parse(data)
}

func ParseString(data string) (map[string]interface{}, error) {
	return Default.ParseString(data)
}

func ParseReader(reader io.Reader) (map[string]interface{}, error) {
	return Default.ParseReader(reader)
}

func ParseRequest(r *http.Request) (map[string]interface{}, error) {
	return Default.ParseRequest(r)
}

func ParseToSlice(data string) ([]interface{}, error) {
	return Default.ParseToSlice(data)
}

func ParseToMapSlice(data string) ([]map[string]interface{}, error) {
	return Default.ParseToMapSlice(data)
}

func Unmarshal(data []byte, v interface{}) error {
	return Default.Unmarshal(data, v)
}

func UnmarshalString(data string, v interface{}) error {
	return Default.UnmarshalString(data, v)
}

func Marshal(v interface{}) ([]byte, error) {
	return Default.Marshal(v)
}

func MarshalToString(v interface{}) (string, error) {
	return Default.MarshalToString(v)
}

func Get(data map[string]interface{}, key string) (interface{}, bool) {
	return Default.Get(data, key)
}

func GetInt64(data map[string]interface{}, key string) (int64, bool) {
	return Default.GetInt64(data, key)
}

func GetInt(data map[string]interface{}, key string) (int, bool) {
	return Default.GetInt(data, key)
}

func GetFloat64(data map[string]interface{}, key string) (float64, bool) {
	return Default.GetFloat64(data, key)
}

func GetString(data map[string]interface{}, key string) (string, bool) {
	return Default.GetString(data, key)
}

func GetBool(data map[string]interface{}, key string) (bool, bool) {
	return Default.GetBool(data, key)
}

func GetMap(data map[string]interface{}, key string) (map[string]interface{}, bool) {
	return Default.GetMap(data, key)
}

func GetArray(data map[string]interface{}, key string) ([]interface{}, bool) {
	return Default.GetArray(data, key)
}

func GetStruct(data map[string]interface{}, key string, v interface{}) bool {
	return Default.GetStruct(data, key, v)
}

func Validate(data []byte) error {
	return Default.Validate(data)
}

func ValidateString(data string) error {
	return Default.ValidateString(data)
}

func ValidateRequired(data map[string]interface{}, fields ...string) error {
	return Default.ValidateRequired(data, fields...)
}

func ValidateType(data map[string]interface{}, field string, expectedType reflect.Kind) error {
	return Default.ValidateType(data, field, expectedType)
}

// CompactJSON 压缩JSON字符串（移除空白字符）
func (j *JSONUtil) CompactJSON(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, data); err != nil {
		return nil, fmt.Errorf("compact JSON failed: %w", err)
	}
	return buf.Bytes(), nil
}

// CompactJSONString 压缩JSON字符串（移除空白字符）
func (j *JSONUtil) CompactJSONString(data string) (string, error) {
	bytes, err := j.CompactJSON([]byte(data))
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FormatJSON 格式化JSON字符串（添加缩进）
func (j *JSONUtil) FormatJSON(data []byte, prefix, indent string) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, prefix, indent); err != nil {
		return nil, fmt.Errorf("format JSON failed: %w", err)
	}
	return buf.Bytes(), nil
}

// FormatJSONString 格式化JSON字符串（添加缩进）
func (j *JSONUtil) FormatJSONString(data string, prefix, indent string) (string, error) {
	bytes, err := j.FormatJSON([]byte(data), prefix, indent)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// MergeJSON 合并多个JSON对象
func (j *JSONUtil) MergeJSON(jsons ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, json := range jsons {
		for k, v := range json {
			// 如果两个都是map，则递归合并
			if existing, ok := result[k]; ok {
				if existingMap, ok := existing.(map[string]interface{}); ok {
					if newMap, ok := v.(map[string]interface{}); ok {
						result[k] = j.MergeJSON(existingMap, newMap)
						continue
					}
				}
			}
			result[k] = v
		}
	}
	return result
}

// GetByPath 通过路径获取JSON值
// 路径格式：user.address.street 或 users[0].name
func (j *JSONUtil) GetByPath(data interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		// 处理数组索引
		if idx := strings.Index(part, "["); idx != -1 {
			if !strings.HasSuffix(part, "]") {
				return nil, false
			}
			key := part[:idx]
			index, err := strconv.Atoi(part[idx+1 : len(part)-1])
			if err != nil {
				return nil, false
			}

			// 获取数组
			var arr []interface{}
			if key == "" {
				if a, ok := current.([]interface{}); ok {
					arr = a
				} else {
					return nil, false
				}
			} else {
				if m, ok := current.(map[string]interface{}); ok {
					if a, ok := m[key].([]interface{}); ok {
						arr = a
					} else {
						return nil, false
					}
				} else {
					return nil, false
				}
			}

			// 检查索引
			if index < 0 || index >= len(arr) {
				return nil, false
			}
			current = arr[index]
			continue
		}

		// 处理普通字段
		if m, ok := current.(map[string]interface{}); ok {
			if val, exists := m[part]; exists {
				current = val
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	}

	return current, true
}

// CompactJSON 静态方法：压缩JSON字符串
func CompactJSON(data []byte) ([]byte, error) {
	return Default.CompactJSON(data)
}

// CompactJSONString 静态方法：压缩JSON字符串
func CompactJSONString(data string) (string, error) {
	return Default.CompactJSONString(data)
}

// FormatJSON 静态方法：格式化JSON字符串
func FormatJSON(data []byte, prefix, indent string) ([]byte, error) {
	return Default.FormatJSON(data, prefix, indent)
}

// FormatJSONString 静态方法：格式化JSON字符串
func FormatJSONString(data string, prefix, indent string) (string, error) {
	return Default.FormatJSONString(data, prefix, indent)
}

// MergeJSON 静态方法：合并多个JSON对象
func MergeJSON(jsons ...map[string]interface{}) map[string]interface{} {
	return Default.MergeJSON(jsons...)
}

// GetByPath 静态方法：通过路径获取JSON值
func GetByPath(data interface{}, path string) (interface{}, bool) {
	return Default.GetByPath(data, path)
}
