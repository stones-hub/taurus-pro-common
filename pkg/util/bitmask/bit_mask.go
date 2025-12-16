package bitmask

// 子掩码位运算应用

// Add 添加一个掩码 (比如：添加一个权限码)
// 本质是做「并集」运算：mask ∪ value
// 对应按位或：只要某一位在 mask 或 value 中为 1，则结果中该位为 1
func Add(mask uint64, value uint64) uint64 {
	return mask | value
}

// Remove 移除一个掩码 (比如：移除一个权限码)
// 可以理解为从集合 mask 中减去 value，属于「差集」运算
func Remove(mask uint64, value uint64) uint64 {
	return mask &^ value
}

// Has 判断一个掩码是否存在(比如：判断一个权限码是否存在)
// 本质是判断「交集是否非空」：mask ∩ value ≠ ∅
// 只要有任意一位同时在 mask 和 value 中为 1，则返回 true
func Has(mask uint64, value uint64) bool {
	return mask&value != 0
}

// HasAll 判断一个掩码是否全部存在(比如：判断一个权限码是否全部存在)
// 本质是判断「value 是否是 mask 的子集」：
// 即 mask ∩ value == value，说明 mask 包含 value 的所有位
func HasAll(mask uint64, value uint64) bool {
	return mask&value == value
}

// Equals 判断两个掩码是否完全相等
// 相当于判断两个集合是否完全一致
func Equals(mask uint64, value uint64) bool {
	return mask == value
}

// IsEmpty 判断一个掩码是否为空
// 相当于判断集合是否为空集
func IsEmpty(mask uint64) bool {
	return mask == 0
}

// Toggle 切换一个掩码 (比如：切换一个权限码的状态)
// 如果掩码中存在 value，则移除；如果不存在，则添加
// 本质是「对称差集」运算：mask △ value
// 对应按位异或：对应位不同则为 1，相同则为 0
func Toggle(mask uint64, value uint64) uint64 {
	return mask ^ value
}

// Clear 清空所有掩码
// 返回 0，相当于重置为空集
func Clear(mask uint64) uint64 {
	return 0
}

// Union 计算两个掩码的并集
// 本质是「并集」运算：mask1 ∪ mask2
// 功能与 Add 相同，但更符合集合运算的语义
func Union(mask1 uint64, mask2 uint64) uint64 {
	return Add(mask1, mask2)
}

// Intersection 计算两个掩码的交集
// 本质是「交集」运算：mask1 ∩ mask2
// 返回两个掩码共同拥有的位
func Intersection(mask1 uint64, mask2 uint64) uint64 {
	return mask1 & mask2
}

// Difference 计算两个掩码的差集
// 本质是「差集」运算：mask1 - mask2
// 从 mask1 中移除 mask2 中存在的位
// 功能与 Remove 相同，但更符合集合运算的语义
func Difference(mask1 uint64, mask2 uint64) uint64 {
	return Remove(mask1, mask2)
}

// XOR 计算两个掩码的异或（对称差集）
// 本质是「对称差集」运算：mask1 △ mask2
// 返回只在其中一个掩码中存在的位
// 功能与 Toggle 相同，但更符合集合运算的语义
func XOR(mask1 uint64, mask2 uint64) uint64 {
	return Toggle(mask1, mask2)
}

// Count 计算掩码中设置的位数（popcount）
// 返回掩码中值为 1 的位的数量
func Count(mask uint64) int {
	// 使用 Brian Kernighan 算法，高效计算 1 的个数
	count := 0
	for mask != 0 {
		mask &= mask - 1 // 清除最低位的 1
		count++
	}
	return count
}

// List 列出掩码中所有设置的位
// 返回所有值为 1 的位的掩码值列表（如 [1, 2, 4, 8...]）
// 用于调试或需要遍历所有设置的位时使用
func List(mask uint64) []uint64 {
	var result []uint64
	for i := 0; i < 64; i++ {
		bit := uint64(1 << i)
		if mask&bit != 0 {
			result = append(result, bit)
		}
	}
	return result
}

// ListIndices 列出掩码中所有设置的位的索引
// 返回所有值为 1 的位的索引列表（如 [0, 1, 2, 3...]）
// 用于需要知道具体是第几位被设置时使用
func ListIndices(mask uint64) []int {
	var result []int
	for i := 0; i < 64; i++ {
		if mask&(1<<i) != 0 {
			result = append(result, i)
		}
	}
	return result
}
