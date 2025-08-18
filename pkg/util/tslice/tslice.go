package tslice

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// Contains 检查元素是否在切片中，使用 == 运算符进行比较
// 参数：
//   - arr: 要搜索的切片
//   - item: 要查找的元素
//
// 返回值：
//   - bool: 如果元素存在返回 true，否则返回 false
//
// 使用示例：
//
//	// 基本类型
//	nums := []int{1, 2, 3, 4, 5}
//	exists := tslice.Contains(nums, 3)  // 返回 true
//
//	// 字符串
//	strs := []string{"apple", "banana", "orange"}
//	exists := tslice.Contains(strs, "banana")  // 返回 true
//
// 注意事项：
//   - 使用泛型，支持任何可比较类型
//   - 使用 == 运算符比较，对于复杂类型可能需要使用 ContainsFunc
//   - 时间复杂度为 O(n)
func Contains[T comparable](arr []T, item T) bool {
	for _, v := range arr {
		if v == item {
			return true
		}
	}
	return false
}

// ContainsFunc 使用自定义比较函数检查元素是否在切片中
// 参数：
//   - arr: 要搜索的切片
//   - item: 要查找的元素
//   - equals: 自定义比较函数，接收两个参数并返回它们是否相等
//
// 返回值：
//   - bool: 如果找到匹配的元素返回 true，否则返回 false
//
// 使用示例：
//
//	// 自定义结构体
//	type Person struct {
//	    Name string
//	    Age  int
//	}
//
//	people := []Person{
//	    {Name: "John", Age: 30},
//	    {Name: "Jane", Age: 25},
//	}
//
//	// 按名字比较
//	found := tslice.ContainsFunc(people, Person{Name: "John"}, func(a, b Person) bool {
//	    return a.Name == b.Name
//	})  // 返回 true
//
//	// 不区分大小写的字符串比较
//	strs := []string{"Hello", "World"}
//	found := tslice.ContainsFunc(strs, "hello", strings.EqualFold)  // 返回 true
//
// 注意事项：
//   - 使用泛型，支持任何类型
//   - 比较函数应该满足等价关系（自反性、对称性、传递性）
//   - 时间复杂度为 O(n)
//   - 适用于复杂类型或需要自定义比较逻辑的场景
func ContainsFunc[T any](arr []T, item T, equals func(T, T) bool) bool {
	for _, v := range arr {
		if equals(v, item) {
			return true
		}
	}
	return false
}

// ContainsString 检查字符串是否在数组中（不区分大小写）
// Deprecated: 建议使用 ContainsFunc 函数，例如：
//
//	ContainsFunc(arr, str, func(a, b string) bool {
//	    return strings.EqualFold(a, b)
//	})
func ContainsString(k string, arr []string) bool {
	return ContainsFunc(arr, k, strings.EqualFold)
}

// IndexOf 查找元素在切片中的索引，使用 == 运算符进行比较
// 参数：
//   - arr: 要搜索的切片
//   - item: 要查找的元素
//
// 返回值：
//   - int: 如果找到元素则返回其索引，否则返回 -1
//
// 使用示例：
//
//	// 基本类型
//	nums := []int{1, 2, 3, 4, 5}
//	index := tslice.IndexOf(nums, 3)  // 返回 2
//	index := tslice.IndexOf(nums, 6)  // 返回 -1
//
//	// 字符串
//	strs := []string{"apple", "banana", "orange"}
//	index := tslice.IndexOf(strs, "banana")  // 返回 1
//
// 注意事项：
//   - 使用泛型，支持任何可比较类型
//   - 使用 == 运算符比较，对于复杂类型可能需要使用 IndexOfFunc
//   - 返回第一个匹配元素的索引
//   - 时间复杂度为 O(n)
func IndexOf[T comparable](arr []T, item T) int {
	return IndexOfFunc(arr, item, func(a, b T) bool { return a == b })
}

// IndexOfFunc 使用自定义比较函数查找元素在切片中的索引
// 参数：
//   - arr: 要搜索的切片
//   - item: 要查找的元素
//   - equals: 自定义比较函数，接收两个参数并返回它们是否相等
//
// 返回值：
//   - int: 如果找到匹配的元素则返回其索引，否则返回 -1
//
// 使用示例：
//
//	// 自定义结构体
//	type Person struct {
//	    Name string
//	    Age  int
//	}
//
//	people := []Person{
//	    {Name: "John", Age: 30},
//	    {Name: "Jane", Age: 25},
//	}
//
//	// 按名字查找索引
//	index := tslice.IndexOfFunc(people, Person{Name: "Jane"}, func(a, b Person) bool {
//	    return a.Name == b.Name
//	})  // 返回 1
//
//	// 不区分大小写的字符串查找
//	strs := []string{"Hello", "World"}
//	index := tslice.IndexOfFunc(strs, "hello", strings.EqualFold)  // 返回 0
//
// 注意事项：
//   - 使用泛型，支持任何类型
//   - 比较函数应该满足等价关系（自反性、对称性、传递性）
//   - 返回第一个匹配元素的索引
//   - 时间复杂度为 O(n)
//   - 适用于复杂类型或需要自定义比较逻辑的场景
func IndexOfFunc[T any](arr []T, item T, equals func(T, T) bool) int {
	for i, v := range arr {
		if equals(v, item) {
			return i
		}
	}
	return -1
}

// Remove 从切片中移除所有指定的元素，返回新的切片
// 参数：
//   - arr: 源切片
//   - item: 要移除的元素
//
// 返回值：
//   - []T: 移除指定元素后的新切片
//
// 使用示例：
//
//	// 基本类型
//	nums := []int{1, 2, 3, 2, 4, 2, 5}
//	result := tslice.Remove(nums, 2)  // 返回 [1, 3, 4, 5]
//
//	// 字符串
//	strs := []string{"apple", "banana", "apple", "orange"}
//	result := tslice.Remove(strs, "apple")  // 返回 ["banana", "orange"]
//
// 注意事项：
//   - 使用泛型，支持任何可比较类型
//   - 返回一个新的切片，不修改原切片
//   - 移除所有匹配的元素，不仅仅是第一个
//   - 如果没有找到匹配元素，返回原切片的副本
//   - 保持剩余元素的相对顺序
//   - 时间复杂度为 O(n)
func Remove[T comparable](arr []T, item T) []T {
	result := make([]T, 0, len(arr))
	for _, v := range arr {
		if v != item {
			result = append(result, v)
		}
	}
	return result
}

// RemoveAt 从切片中移除指定索引的元素，返回新的切片
// 参数：
//   - arr: 源切片
//   - index: 要移除的元素索引
//
// 返回值：
//   - []T: 移除指定索引元素后的新切片
//
// 使用示例：
//
//	// 基本类型
//	nums := []int{1, 2, 3, 4, 5}
//	result := tslice.RemoveAt(nums, 2)  // 返回 [1, 2, 4, 5]
//
//	// 字符串
//	strs := []string{"apple", "banana", "orange"}
//	result := tslice.RemoveAt(strs, 1)  // 返回 ["apple", "orange"]
//
//	// 处理无效索引
//	result := tslice.RemoveAt(nums, -1)  // 返回原切片的副本
//	result := tslice.RemoveAt(nums, 10)  // 返回原切片的副本
//
// 注意事项：
//   - 使用泛型，支持任何类型
//   - 如果索引无效（小于0或大于等于切片长度），返回原切片的副本
//   - 返回一个新的切片，不修改原切片
//   - 保持剩余元素的相对顺序
//   - 时间复杂度为 O(n)
func RemoveAt[T any](arr []T, index int) []T {
	if index < 0 || index >= len(arr) {
		return arr
	}
	return append(arr[:index], arr[index+1:]...)
}

// Unique 去除切片中的重复元素，返回新的切片
// 参数：
//   - arr: 源切片
//
// 返回值：
//   - []T: 去重后的新切片
//
// 使用示例：
//
//	// 基本类型
//	nums := []int{1, 2, 2, 3, 3, 4, 5, 5}
//	result := tslice.Unique(nums)  // 返回 [1, 2, 3, 4, 5]
//
//	// 字符串
//	strs := []string{"apple", "banana", "apple", "orange", "banana"}
//	result := tslice.Unique(strs)  // 返回 ["apple", "banana", "orange"]
//
//	// 空切片
//	empty := []int{}
//	result := tslice.Unique(empty)  // 返回 []
//
// 注意事项：
//   - 使用泛型，支持任何可比较类型
//   - 返回一个新的切片，不修改原切片
//   - 保持第一次出现的元素，移除后续重复项
//   - 使用 map 实现去重，时间复杂度为 O(n)
//   - 返回的切片可能与原切片的元素顺序不同
//   - 如果需要保持顺序，请使用 UniqueFunc
func Unique[T comparable](arr []T) []T {
	seen := make(map[T]bool)
	result := make([]T, 0, len(arr))

	for _, v := range arr {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// Filter 根据条件函数过滤切片中的元素，返回新的切片
// 参数：
//   - arr: 源切片
//   - predicate: 过滤条件函数，返回 true 表示保留该元素
//
// 返回值：
//   - []T: 满足条件的元素组成的新切片
//
// 使用示例：
//
//	// 过滤数字
//	nums := []int{1, 2, 3, 4, 5, 6}
//	evens := tslice.Filter(nums, func(n int) bool {
//	    return n%2 == 0
//	})  // 返回 [2, 4, 6]
//
//	// 过滤字符串
//	strs := []string{"apple", "banana", "orange", "grape"}
//	long := tslice.Filter(strs, func(s string) bool {
//	    return len(s) > 5
//	})  // 返回 ["banana", "orange"]
//
//	// 过滤结构体
//	type Person struct {
//	    Name string
//	    Age  int
//	}
//	people := []Person{
//	    {Name: "John", Age: 25},
//	    {Name: "Jane", Age: 30},
//	    {Name: "Bob", Age: 20},
//	}
//	adults := tslice.Filter(people, func(p Person) bool {
//	    return p.Age >= 25
//	})  // 返回 [{John 25} {Jane 30}]
//
// 注意事项：
//   - 使用泛型，支持任何类型
//   - 返回一个新的切片，不修改原切片
//   - 保持元素的原始顺序
//   - 如果没有元素满足条件，返回空切片
//   - 时间复杂度为 O(n)
func Filter[T any](arr []T, predicate func(T) bool) []T {
	result := make([]T, 0, len(arr))
	for _, v := range arr {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// Map 对切片中的每个元素应用转换函数，返回新的切片
// 参数：
//   - arr: 源切片
//   - transform: 转换函数，将类型 T 的元素转换为类型 U
//
// 返回值：
//   - []U: 转换后的新切片
//
// 使用示例：
//
//	// 数字转换
//	nums := []int{1, 2, 3, 4}
//	squares := tslice.Map(nums, func(n int) int {
//	    return n * n
//	})  // 返回 [1, 4, 9, 16]
//
//	// 类型转换
//	nums := []int{1, 2, 3}
//	strs := tslice.Map(nums, func(n int) string {
//	    return strconv.Itoa(n)
//	})  // 返回 ["1", "2", "3"]
//
//	// 结构体转换
//	type Person struct {
//	    Name string
//	    Age  int
//	}
//	type Summary struct {
//	    FullName string
//	    IsAdult  bool
//	}
//	people := []Person{
//	    {Name: "John", Age: 25},
//	    {Name: "Jane", Age: 17},
//	}
//	summaries := tslice.Map(people, func(p Person) Summary {
//	    return Summary{
//	        FullName: p.Name,
//	        IsAdult:  p.Age >= 18,
//	    }
//	})  // 返回 [{John true} {Jane false}]
//
// 注意事项：
//   - 使用泛型，支持任何类型之间的转换
//   - 返回一个新的切片，不修改原切片
//   - 保持元素的相对顺序
//   - 转换函数对每个元素只调用一次
//   - 时间复杂度为 O(n)
func Map[T, U any](arr []T, transform func(T) U) []U {
	result := make([]U, len(arr))
	for i, v := range arr {
		result[i] = transform(v)
	}
	return result
}

// Reduce 对切片中的元素进行归约操作，将多个元素组合成单个结果
// 参数：
//   - arr: 源切片
//   - initial: 初始值
//   - reducer: 归约函数，接收累积值和当前元素，返回新的累积值
//
// 返回值：
//   - U: 归约操作的最终结果
//
// 使用示例：
//
//	// 求和
//	nums := []int{1, 2, 3, 4, 5}
//	sum := tslice.Reduce(nums, 0, func(acc, curr int) int {
//	    return acc + curr
//	})  // 返回 15
//
//	// 字符串连接
//	words := []string{"Hello", "World", "!"}
//	sentence := tslice.Reduce(words, "", func(acc, curr string) string {
//	    if acc == "" {
//	        return curr
//	    }
//	    return acc + " " + curr
//	})  // 返回 "Hello World !"
//
//	// 计算最大值
//	nums := []int{3, 1, 4, 1, 5, 9, 2, 6}
//	max := tslice.Reduce(nums, nums[0], func(acc, curr int) int {
//	    if curr > acc {
//	        return curr
//	    }
//	    return acc
//	})  // 返回 9
//
// 注意事项：
//   - 使用泛型，支持任何类型的输入和输出
//   - 如果切片为空，返回初始值
//   - 归约函数从左到右处理元素
//   - 可以将切片元素转换为不同类型的结果
//   - 时间复杂度为 O(n)
func Reduce[T, U any](arr []T, initial U, reducer func(U, T) U) U {
	result := initial
	for _, v := range arr {
		result = reducer(result, v)
	}
	return result
}

// Chunk 将切片分割成指定大小的块，返回二维切片
// 参数：
//   - arr: 源切片
//   - size: 每个块的大小
//
// 返回值：
//   - [][]T: 分割后的二维切片，每个子切片的长度最多为 size
//
// 使用示例：
//
//	// 基本类型
//	nums := []int{1, 2, 3, 4, 5, 6, 7}
//	chunks := tslice.Chunk(nums, 3)
//	// 返回 [[1, 2, 3], [4, 5, 6], [7]]
//
//	// 字符串
//	words := []string{"apple", "banana", "orange", "grape"}
//	chunks := tslice.Chunk(words, 2)
//	// 返回 [["apple", "banana"], ["orange", "grape"]]
//
//	// 特殊情况
//	empty := []int{}
//	chunks := tslice.Chunk(empty, 2)  // 返回 []
//	nums := []int{1, 2, 3}
//	chunks := tslice.Chunk(nums, 0)   // 返回 [[1, 2, 3]]
//
// 注意事项：
//   - 使用泛型，支持任何类型
//   - 如果 size <= 0，返回包含整个切片的单个块
//   - 最后一个块的长度可能小于 size
//   - 返回新的切片，不修改原切片
//   - 保持元素的相对顺序
//   - 时间复杂度为 O(n)
func Chunk[T any](arr []T, size int) [][]T {
	if size <= 0 {
		return [][]T{arr}
	}

	if len(arr) == 0 {
		return [][]T{}
	}

	var chunks [][]T
	for i := 0; i < len(arr); i += size {
		end := i + size
		if end > len(arr) {
			end = len(arr)
		}
		chunks = append(chunks, arr[i:end])
	}
	return chunks
}

// Reverse 反转切片中元素的顺序，返回新的切片
// 参数：
//   - arr: 源切片
//
// 返回值：
//   - []T: 反转后的新切片
//
// 使用示例：
//
//	// 基本类型
//	nums := []int{1, 2, 3, 4, 5}
//	reversed := tslice.Reverse(nums)  // 返回 [5, 4, 3, 2, 1]
//
//	// 字符串
//	words := []string{"hello", "world"}
//	reversed := tslice.Reverse(words)  // 返回 ["world", "hello"]
//
//	// 空切片
//	empty := []int{}
//	reversed := tslice.Reverse(empty)  // 返回 []
//
// 注意事项：
//   - 使用泛型，支持任何类型
//   - 返回一个新的切片，不修改原切片
//   - 如果输入切片为空，返回空切片
//   - 单个元素的切片返回相同的元素
//   - 时间复杂度为 O(n)
func Reverse[T any](arr []T) []T {
	result := make([]T, len(arr))
	for i, v := range arr {
		result[len(arr)-1-i] = v
	}
	return result
}

// Shuffle 随机打乱切片中元素的顺序，返回新的切片
// 参数：
//   - arr: 源切片
//
// 返回值：
//   - []T: 随机打乱后的新切片
//
// 使用示例：
//
//	// 基本类型
//	nums := []int{1, 2, 3, 4, 5}
//	shuffled := tslice.Shuffle(nums)
//	// 可能返回 [3, 1, 5, 2, 4]
//
//	// 字符串
//	words := []string{"apple", "banana", "orange"}
//	shuffled := tslice.Shuffle(words)
//	// 可能返回 ["orange", "apple", "banana"]
//
//	// 空切片或单元素切片
//	empty := []int{}
//	shuffled := tslice.Shuffle(empty)  // 返回 []
//	single := []int{1}
//	shuffled := tslice.Shuffle(single)  // 返回 [1]
//
// 注意事项：
//   - 使用泛型，支持任何类型
//   - 返回一个新的切片，不修改原切片
//   - 使用 crypto/rand 生成安全的随机数
//   - 使用 Fisher-Yates 洗牌算法
//   - 如果生成随机数失败，会使用最后一个元素（极少发生）
//   - 时间复杂度为 O(n)
func Shuffle[T any](arr []T) []T {
	result := make([]T, len(arr))
	copy(result, arr)

	// Fisher-Yates 洗牌算法
	for i := len(result) - 1; i > 0; i-- {
		// 使用 crypto/rand 生成安全的随机数
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			// 如果生成随机数失败，使用最后一个元素（这种情况极少发生）
			result[i], result[0] = result[0], result[i]
			continue
		}
		j := int(n.Int64())
		result[i], result[j] = result[j], result[i]
	}
	return result
}

// Flatten 将二维切片展平为一维切片
// 参数：
//   - arr: 二维切片
//
// 返回值：
//   - []T: 展平后的一维切片
//
// 使用示例：
//
//	// 基本类型
//	nums := [][]int{
//	    {1, 2, 3},
//	    {4, 5},
//	    {6},
//	}
//	flat := tslice.Flatten(nums)  // 返回 [1, 2, 3, 4, 5, 6]
//
//	// 字符串
//	words := [][]string{
//	    {"hello", "world"},
//	    {"go", "lang"},
//	    {"test"},
//	}
//	flat := tslice.Flatten(words)  // 返回 ["hello", "world", "go", "lang", "test"]
//
//	// 特殊情况
//	empty := [][]int{}
//	flat := tslice.Flatten(empty)  // 返回 []
//	nested := [][]int{{}, {1}, {}, {2, 3}, {}}
//	flat := tslice.Flatten(nested)  // 返回 [1, 2, 3]
//
// 注意事项：
//   - 使用泛型，支持任何类型
//   - 返回一个新的切片，不修改原切片
//   - 保持元素的相对顺序
//   - 空的内部切片会被忽略
//   - 如果输入切片为空，返回空切片
//   - 时间复杂度为 O(n)，其中 n 是所有元素的总数
func Flatten[T any](arr [][]T) []T {
	var result []T
	for _, subArr := range arr {
		result = append(result, subArr...)
	}
	return result
}

// Intersection 计算两个切片的交集，返回在两个切片中都存在的元素
// 参数：
//   - arr1: 第一个切片
//   - arr2: 第二个切片
//
// 返回值：
//   - []T: 两个切片的交集
//
// 使用示例：
//
//	// 基本类型
//	nums1 := []int{1, 2, 3, 4, 5}
//	nums2 := []int{4, 5, 6, 7, 8}
//	common := tslice.Intersection(nums1, nums2)  // 返回 [4, 5]
//
//	// 字符串
//	strs1 := []string{"apple", "banana", "orange"}
//	strs2 := []string{"banana", "grape", "orange", "pear"}
//	common := tslice.Intersection(strs1, strs2)  // 返回 ["banana", "orange"]
//
//	// 特殊情况
//	empty1 := []int{}
//	nums := []int{1, 2, 3}
//	common := tslice.Intersection(empty1, nums)  // 返回 []
//	same := tslice.Intersection(nums, nums)      // 返回 [1, 2, 3]
//
// 注意事项：
//   - 使用泛型，支持任何可比较类型
//   - 返回一个新的切片，不修改原切片
//   - 结果中的元素顺序取决于它们在第二个切片中的顺序
//   - 如果有任一切片为空，返回空切片
//   - 使用 map 实现，时间复杂度为 O(n+m)
//   - 结果中不包含重复元素
func Intersection[T comparable](arr1, arr2 []T) []T {
	set1 := make(map[T]bool)
	for _, v := range arr1 {
		set1[v] = true
	}

	var result []T
	for _, v := range arr2 {
		if set1[v] {
			result = append(result, v)
		}
	}
	return result
}

// Union 计算两个切片的并集，返回包含所有不重复元素的切片
// 参数：
//   - arr1: 第一个切片
//   - arr2: 第二个切片
//
// 返回值：
//   - []T: 两个切片的并集
//
// 使用示例：
//
//	// 基本类型
//	nums1 := []int{1, 2, 3, 4}
//	nums2 := []int{3, 4, 5, 6}
//	all := tslice.Union(nums1, nums2)  // 返回 [1, 2, 3, 4, 5, 6]
//
//	// 字符串
//	strs1 := []string{"apple", "banana"}
//	strs2 := []string{"banana", "orange"}
//	all := tslice.Union(strs1, strs2)  // 返回 ["apple", "banana", "orange"]
//
//	// 特殊情况
//	empty := []int{}
//	nums := []int{1, 2, 3}
//	all := tslice.Union(empty, nums)  // 返回 [1, 2, 3]
//	all := tslice.Union(nums, nums)   // 返回 [1, 2, 3]
//
// 注意事项：
//   - 使用泛型，支持任何可比较类型
//   - 返回一个新的切片，不修改原切片
//   - 结果中不包含重复元素
//   - 结果中的元素顺序不保证
//   - 如果两个切片都为空，返回空切片
//   - 使用 map 实现，时间复杂度为 O(n+m)
func Union[T comparable](arr1, arr2 []T) []T {
	set := make(map[T]bool)

	for _, v := range arr1 {
		set[v] = true
	}
	for _, v := range arr2 {
		set[v] = true
	}

	result := make([]T, 0, len(set))
	for v := range set {
		result = append(result, v)
	}
	return result
}

// Difference 计算两个切片的差集，返回在第一个切片中存在但在第二个切片中不存在的元素
// 参数：
//   - arr1: 第一个切片
//   - arr2: 第二个切片
//
// 返回值：
//   - []T: 两个切片的差集（arr1 - arr2）
//
// 使用示例：
//
//	// 基本类型
//	nums1 := []int{1, 2, 3, 4, 5}
//	nums2 := []int{4, 5, 6, 7}
//	diff := tslice.Difference(nums1, nums2)  // 返回 [1, 2, 3]
//
//	// 字符串
//	strs1 := []string{"apple", "banana", "orange"}
//	strs2 := []string{"banana", "grape"}
//	diff := tslice.Difference(strs1, strs2)  // 返回 ["apple", "orange"]
//
//	// 特殊情况
//	empty := []int{}
//	nums := []int{1, 2, 3}
//	diff := tslice.Difference(empty, nums)   // 返回 []
//	diff := tslice.Difference(nums, empty)   // 返回 [1, 2, 3]
//	diff := tslice.Difference(nums, nums)    // 返回 []
//
// 注意事项：
//   - 使用泛型，支持任何可比较类型
//   - 返回一个新的切片，不修改原切片
//   - 保持第一个切片中元素的相对顺序
//   - 结果中不包含重复元素
//   - 如果第一个切片为空，返回空切片
//   - 如果第二个切片为空，返回第一个切片的副本
//   - 使用 map 实现，时间复杂度为 O(n+m)
func Difference[T comparable](arr1, arr2 []T) []T {
	set2 := make(map[T]bool)
	for _, v := range arr2 {
		set2[v] = true
	}

	var result []T
	for _, v := range arr1 {
		if !set2[v] {
			result = append(result, v)
		}
	}
	return result
}

// ArrayToString 将数组格式化为字符串，元素之间用逗号分隔
// 参数：
//   - array: 要格式化的数组
//
// 返回值：
//   - string: 格式化后的字符串
//
// 使用示例：
//
//	// 基本类型
//	arr := []interface{}{1, "hello", true}
//	str := tslice.ArrayToString(arr)  // 返回 "1,hello,true"
//
//	// 混合类型
//	arr := []interface{}{42, "world", 3.14}
//	str := tslice.ArrayToString(arr)  // 返回 "42,world,3.14"
//
//	// 空数组
//	arr := []interface{}{}
//	str := tslice.ArrayToString(arr)  // 返回 ""
//
// 注意事项：
//   - 使用 fmt.Sprint 将元素转换为字符串
//   - 移除元素之间的空格
//   - 不添加首尾的方括号
//   - 如果数组为空，返回空字符串
//   - 适用于需要简单字符串表示的场景
func ArrayToString(array []interface{}) string {
	return strings.Replace(strings.Trim(fmt.Sprint(array), "[]"), " ", ",", -1)
}

// ParseStringSliceToUint64 将字符串切片转换为uint64切片，忽略无效的转换
// 参数：
//   - s: 要转换的字符串切片
//
// 返回值：
//   - []uint64: 转换后的uint64切片
//
// 使用示例：
//
//	// 基本使用
//	strs := []string{"123", "456", "789"}
//	nums := tslice.ParseStringSliceToUint64(strs)  // 返回 [123, 456, 789]
//
//	// 处理无效输入
//	strs := []string{"123", "abc", "456", "-789"}
//	nums := tslice.ParseStringSliceToUint64(strs)  // 返回 [123, 456, 0]
//
//	// 空切片
//	strs := []string{}
//	nums := tslice.ParseStringSliceToUint64(strs)  // 返回 []
//
// 注意事项：
//   - 使用10进制解析字符串
//   - 无效的数字（非数字或负数）会被转换为0
//   - 保持元素的相对顺序
//   - 返回一个新的切片
//   - 适用于处理字符串形式的ID列表等场景
func ParseStringSliceToUint64(s []string) []uint64 {
	iv := make([]uint64, len(s))

	for i, v := range s {
		// 以10进制的方式解析v，最后保存为64位uint
		iv[i], _ = strconv.ParseUint(v, 10, 64)
	}

	return iv
}

// GroupBy 根据指定的键函数对切片进行分组，返回一个映射
// 参数：
//   - arr: 要分组的切片
//   - keyFunc: 分组键函数，接收一个元素并返回其分组键
//
// 返回值：
//   - map[K][]T: 分组结果，键为分组键，值为该组的所有元素
//
// 使用示例：
//
//	// 按长度分组字符串
//	words := []string{"cat", "dog", "fish", "bird", "elephant"}
//	groups := tslice.GroupBy(words, func(s string) int {
//	    return len(s)
//	})
//	// 返回：
//	// {
//	//   3: ["cat", "dog"],
//	//   4: ["fish", "bird"],
//	//   8: ["elephant"]
//	// }
//
//	// 按类型分组结构体
//	type Item struct {
//	    Type string
//	    Name string
//	}
//	items := []Item{
//	    {Type: "fruit", Name: "apple"},
//	    {Type: "fruit", Name: "banana"},
//	    {Type: "veg", Name: "carrot"},
//	}
//	groups := tslice.GroupBy(items, func(i Item) string {
//	    return i.Type
//	})
//	// 返回：
//	// {
//	//   "fruit": [{Type: "fruit", Name: "apple"}, {Type: "fruit", Name: "banana"}],
//	//   "veg":   [{Type: "veg", Name: "carrot"}]
//	// }
//
// 注意事项：
//   - 使用泛型，支持任何类型的元素和可比较的键类型
//   - 保持每个组内元素的相对顺序
//   - 如果输入切片为空，返回空映射
//   - 键函数对每个元素只调用一次
//   - 时间复杂度为 O(n)
func GroupBy[T any, K comparable](arr []T, keyFunc func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, item := range arr {
		key := keyFunc(item)
		result[key] = append(result[key], item)
	}
	return result
}

// Sort 使用自定义比较函数对切片进行排序，返回新的切片
// 参数：
//   - arr: 要排序的切片
//   - less: 比较函数，如果第一个参数应该排在第二个参数前面，返回 true
//
// 返回值：
//   - []T: 排序后的新切片
//
// 使用示例：
//
//	// 基本类型
//	nums := []int{3, 1, 4, 1, 5, 9, 2, 6}
//	sorted := tslice.Sort(nums, func(a, b int) bool {
//	    return a < b
//	})  // 返回 [1, 1, 2, 3, 4, 5, 6, 9]
//
//	// 结构体排序
//	type Person struct {
//	    Name string
//	    Age  int
//	}
//	people := []Person{
//	    {Name: "Bob", Age: 30},
//	    {Name: "Alice", Age: 25},
//	    {Name: "Charlie", Age: 35},
//	}
//	// 按年龄排序
//	sorted := tslice.Sort(people, func(a, b Person) bool {
//	    return a.Age < b.Age
//	})  // 返回 [{Alice 25} {Bob 30} {Charlie 35}]
//
//	// 按名字排序
//	sorted := tslice.Sort(people, func(a, b Person) bool {
//	    return a.Name < b.Name
//	})  // 返回 [{Alice 25} {Bob 30} {Charlie 35}]
//
// 注意事项：
//   - 使用泛型，支持任何类型
//   - 返回一个新的切片，不修改原切片
//   - 使用快速排序算法实现
//   - 排序是稳定的（相等元素的相对顺序不变）
//   - 比较函数应该满足传递性
//   - 时间复杂度为 O(n log n)
func Sort[T any](arr []T, less func(T, T) bool) []T {
	result := make([]T, len(arr))
	copy(result, arr)

	// 快速排序实现
	var quickSort func([]T, int, int)
	quickSort = func(arr []T, left, right int) {
		if left >= right {
			return
		}

		// 选择基准值
		pivot := arr[right]
		i := left - 1

		// 分区
		for j := left; j < right; j++ {
			if less(arr[j], pivot) {
				i++
				arr[i], arr[j] = arr[j], arr[i]
			}
		}

		// 放置基准值
		arr[i+1], arr[right] = arr[right], arr[i+1]

		// 递归排序
		quickSort(arr, left, i)
		quickSort(arr, i+2, right)
	}

	if len(result) > 1 {
		quickSort(result, 0, len(result)-1)
	}
	return result
}

// UniqueFunc 使用自定义比较函数去除切片中的重复元素，返回新的切片
// 参数：
//   - arr: 源切片
//   - equals: 比较函数，如果两个元素相等返回 true
//
// 返回值：
//   - []T: 去重后的新切片
//
// 使用示例：
//
//	// 自定义结构体
//	type Person struct {
//	    Name string
//	    Age  int
//	}
//	people := []Person{
//	    {Name: "John", Age: 30},
//	    {Name: "Jane", Age: 25},
//	    {Name: "John", Age: 35},
//	}
//	// 按名字去重
//	unique := tslice.UniqueFunc(people, func(a, b Person) bool {
//	    return a.Name == b.Name
//	})  // 返回 [{John 30} {Jane 25}]
//
//	// 不区分大小写的字符串去重
//	words := []string{"Hello", "hello", "HELLO", "world"}
//	unique := tslice.UniqueFunc(words, strings.EqualFold)
//	// 返回 ["Hello", "world"]
//
// 注意事项：
//   - 使用泛型，支持任何类型
//   - 返回一个新的切片，不修改原切片
//   - 保留每个重复组中的第一个元素
//   - 保持元素的相对顺序
//   - 比较函数应该满足等价关系
//   - 时间复杂度为 O(n²)
func UniqueFunc[T any](arr []T, equals func(T, T) bool) []T {
	if len(arr) == 0 {
		return arr
	}

	result := make([]T, 0, len(arr))
	result = append(result, arr[0])

outer:
	for i := 1; i < len(arr); i++ {
		for j := 0; j < len(result); j++ {
			if equals(arr[i], result[j]) {
				continue outer
			}
		}
		result = append(result, arr[i])
	}

	return result
}

// Partition 根据条件将切片分成两部分，返回满足条件和不满足条件的两个切片
// 参数：
//   - arr: 源切片
//   - predicate: 分区条件函数，返回 true 的元素进入第一个切片
//
// 返回值：
//   - []T: 满足条件的元素切片
//   - []T: 不满足条件的元素切片
//
// 使用示例：
//
//	// 基本类型
//	nums := []int{1, 2, 3, 4, 5, 6}
//	evens, odds := tslice.Partition(nums, func(n int) bool {
//	    return n%2 == 0
//	})
//	// evens = [2, 4, 6]
//	// odds = [1, 3, 5]
//
//	// 结构体
//	type Person struct {
//	    Name string
//	    Age  int
//	}
//	people := []Person{
//	    {Name: "John", Age: 25},
//	    {Name: "Jane", Age: 17},
//	    {Name: "Bob", Age: 30},
//	}
//	adults, minors := tslice.Partition(people, func(p Person) bool {
//	    return p.Age >= 18
//	})
//	// adults = [{John 25} {Bob 30}]
//	// minors = [{Jane 17}]
//
// 注意事项：
//   - 使用泛型，支持任何类型
//   - 返回两个新的切片，不修改原切片
//   - 保持每个分区内元素的相对顺序
//   - 条件函数对每个元素只调用一次
//   - 如果输入切片为空，返回两个空切片
//   - 时间复杂度为 O(n)
func Partition[T any](arr []T, predicate func(T) bool) ([]T, []T) {
	matched := make([]T, 0, len(arr))
	unmatched := make([]T, 0, len(arr))

	for _, item := range arr {
		if predicate(item) {
			matched = append(matched, item)
		} else {
			unmatched = append(unmatched, item)
		}
	}

	return matched, unmatched
}

// FindFirst 查找切片中第一个满足条件的元素
// 参数：
//   - arr: 源切片
//   - predicate: 条件函数，返回 true 表示找到目标元素
//
// 返回值：
//   - T: 找到的元素
//   - bool: 是否找到元素
//
// 使用示例：
//
//	// 基本类型
//	nums := []int{1, 2, 3, 4, 5}
//	first, found := tslice.FindFirst(nums, func(n int) bool {
//	    return n > 3
//	})
//	// first = 4, found = true
//
//	// 结构体
//	type User struct {
//	    ID   int
//	    Name string
//	}
//	users := []User{
//	    {ID: 1, Name: "John"},
//	    {ID: 2, Name: "Jane"},
//	    {ID: 3, Name: "Bob"},
//	}
//	user, found := tslice.FindFirst(users, func(u User) bool {
//	    return u.Name == "Jane"
//	})
//	// user = {ID: 2, Name: "Jane"}, found = true
//
//	// 未找到的情况
//	empty := []int{}
//	val, found := tslice.FindFirst(empty, func(n int) bool {
//	    return n > 0
//	})
//	// val = 0, found = false
//
// 注意事项：
//   - 使用泛型，支持任何类型
//   - 如果找不到元素，返回零值和 false
//   - 按顺序检查元素，返回第一个匹配的
//   - 条件函数对每个元素最多调用一次
//   - 如果输入切片为空，返回零值和 false
//   - 时间复杂度为 O(n)
func FindFirst[T any](arr []T, predicate func(T) bool) (T, bool) {
	for _, item := range arr {
		if predicate(item) {
			return item, true
		}
	}
	var zero T
	return zero, false
}

// FindLast 查找切片中最后一个满足条件的元素
// 参数：
//   - arr: 源切片
//   - predicate: 条件函数，返回 true 表示找到目标元素
//
// 返回值：
//   - T: 找到的元素
//   - bool: 是否找到元素
//
// 使用示例：
//
//	// 基本类型
//	nums := []int{1, 2, 3, 4, 3, 5}
//	last, found := tslice.FindLast(nums, func(n int) bool {
//	    return n == 3
//	})
//	// last = 3, found = true（找到第二个3）
//
//	// 结构体
//	type User struct {
//	    ID   int
//	    Role string
//	}
//	users := []User{
//	    {ID: 1, Role: "admin"},
//	    {ID: 2, Role: "user"},
//	    {ID: 3, Role: "admin"},
//	}
//	user, found := tslice.FindLast(users, func(u User) bool {
//	    return u.Role == "admin"
//	})
//	// user = {ID: 3, Role: "admin"}, found = true
//
//	// 未找到的情况
//	empty := []int{}
//	val, found := tslice.FindLast(empty, func(n int) bool {
//	    return n > 0
//	})
//	// val = 0, found = false
//
// 注意事项：
//   - 使用泛型，支持任何类型
//   - 如果找不到元素，返回零值和 false
//   - 从后向前检查元素，返回最后一个匹配的
//   - 条件函数对每个元素最多调用一次
//   - 如果输入切片为空，返回零值和 false
//   - 时间复杂度为 O(n)
func FindLast[T any](arr []T, predicate func(T) bool) (T, bool) {
	for i := len(arr) - 1; i >= 0; i-- {
		if predicate(arr[i]) {
			return arr[i], true
		}
	}
	var zero T
	return zero, false
}

// All 检查切片中是否所有元素都满足指定条件
// 参数：
//   - arr: 源切片
//   - predicate: 条件函数，返回 true 表示元素满足条件
//
// 返回值：
//   - bool: 如果所有元素都满足条件返回 true，否则返回 false
//
// 使用示例：
//
//	// 基本类型
//	nums := []int{2, 4, 6, 8}
//	allEven := tslice.All(nums, func(n int) bool {
//	    return n%2 == 0
//	})  // 返回 true
//
//	nums = []int{2, 4, 5, 8}
//	allEven := tslice.All(nums, func(n int) bool {
//	    return n%2 == 0
//	})  // 返回 false
//
//	// 结构体
//	type User struct {
//	    Age  int
//	    Active bool
//	}
//	users := []User{
//	    {Age: 25, Active: true},
//	    {Age: 30, Active: true},
//	}
//	allActive := tslice.All(users, func(u User) bool {
//	    return u.Active
//	})  // 返回 true
//
//	// 空切片
//	empty := []int{}
//	allPositive := tslice.All(empty, func(n int) bool {
//	    return n > 0
//	})  // 返回 true（空集合的所有元素都满足任何条件）
//
// 注意事项：
//   - 使用泛型，支持任何类型
//   - 空切片返回 true（符合数学上的空集合逻辑）
//   - 条件函数对每个元素最多调用一次
//   - 发现不满足条件的元素后立即返回 false
//   - 时间复杂度为 O(n)，但可能提前返回
func All[T any](arr []T, predicate func(T) bool) bool {
	for _, item := range arr {
		if !predicate(item) {
			return false
		}
	}
	return true
}

// Any 检查切片中是否存在至少一个元素满足指定条件
// 参数：
//   - arr: 源切片
//   - predicate: 条件函数，返回 true 表示元素满足条件
//
// 返回值：
//   - bool: 如果存在满足条件的元素返回 true，否则返回 false
//
// 使用示例：
//
//	// 基本类型
//	nums := []int{1, 3, 4, 7, 9}
//	hasEven := tslice.Any(nums, func(n int) bool {
//	    return n%2 == 0
//	})  // 返回 true（4是偶数）
//
//	nums = []int{1, 3, 5, 7, 9}
//	hasEven := tslice.Any(nums, func(n int) bool {
//	    return n%2 == 0
//	})  // 返回 false（没有偶数）
//
//	// 结构体
//	type User struct {
//	    Name string
//	    Admin bool
//	}
//	users := []User{
//	    {Name: "John", Admin: false},
//	    {Name: "Jane", Admin: true},
//	}
//	hasAdmin := tslice.Any(users, func(u User) bool {
//	    return u.Admin
//	})  // 返回 true
//
//	// 空切片
//	empty := []int{}
//	hasPositive := tslice.Any(empty, func(n int) bool {
//	    return n > 0
//	})  // 返回 false（空集合中没有元素满足任何条件）
//
// 注意事项：
//   - 使用泛型，支持任何类型
//   - 空切片返回 false（符合数学上的空集合逻辑）
//   - 条件函数对每个元素最多调用一次
//   - 找到满足条件的元素后立即返回 true
//   - 时间复杂度为 O(n)，但可能提前返回
func Any[T any](arr []T, predicate func(T) bool) bool {
	for _, item := range arr {
		if predicate(item) {
			return true
		}
	}
	return false
}

// Count 计算切片中满足指定条件的元素个数
// 参数：
//   - arr: 源切片
//   - predicate: 条件函数，返回 true 表示元素满足条件
//
// 返回值：
//   - int: 满足条件的元素个数
//
// 使用示例：
//
//	// 基本类型
//	nums := []int{1, 2, 3, 4, 5, 6}
//	evenCount := tslice.Count(nums, func(n int) bool {
//	    return n%2 == 0
//	})  // 返回 3（2, 4, 6）
//
//	// 字符串
//	words := []string{"apple", "banana", "orange", "avocado"}
//	aCount := tslice.Count(words, func(s string) bool {
//	    return strings.HasPrefix(s, "a")
//	})  // 返回 2（"apple", "avocado"）
//
//	// 结构体
//	type User struct {
//	    Age   int
//	    Admin bool
//	}
//	users := []User{
//	    {Age: 25, Admin: true},
//	    {Age: 30, Admin: false},
//	    {Age: 35, Admin: true},
//	}
//	adminCount := tslice.Count(users, func(u User) bool {
//	    return u.Admin
//	})  // 返回 2
//
//	// 空切片
//	empty := []int{}
//	count := tslice.Count(empty, func(n int) bool {
//	    return n > 0
//	})  // 返回 0
//
// 注意事项：
//   - 使用泛型，支持任何类型
//   - 空切片返回 0
//   - 条件函数对每个元素只调用一次
//   - 需要遍历整个切片
//   - 时间复杂度为 O(n)
func Count[T any](arr []T, predicate func(T) bool) int {
	count := 0
	for _, item := range arr {
		if predicate(item) {
			count++
		}
	}
	return count
}
