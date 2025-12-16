package bitmask

import "testing"

// 定义测试用的掩码常量
const (
	Bit0  uint64 = 1 << 0  // 0001
	Bit1  uint64 = 1 << 1  // 0010
	Bit2  uint64 = 1 << 2  // 0100
	Bit3  uint64 = 1 << 3  // 1000
	Bit63 uint64 = 1 << 63 // 最高位
)

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		mask     uint64
		value    uint64
		expected uint64
	}{
		{"添加单个位到空掩码", 0, Bit0, Bit0},
		{"添加单个位到已有掩码", Bit0, Bit1, Bit0 | Bit1},
		{"添加组合位", Bit0, Bit1 | Bit2, Bit0 | Bit1 | Bit2},
		{"添加已存在的位", Bit0, Bit0, Bit0},
		{"添加多个位", Bit0 | Bit1, Bit2 | Bit3, Bit0 | Bit1 | Bit2 | Bit3},
		{"添加最高位", 0, Bit63, Bit63},
		{"添加全0", Bit0, 0, Bit0},
		{"从全0添加", 0, Bit0 | Bit1, Bit0 | Bit1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Add(tt.mask, tt.value)
			if result != tt.expected {
				t.Errorf("Add(%d, %d) = %d, expected %d", tt.mask, tt.value, result, tt.expected)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name     string
		mask     uint64
		value    uint64
		expected uint64
	}{
		{"移除单个位", Bit0 | Bit1, Bit0, Bit1},
		{"移除不存在的位", Bit0, Bit1, Bit0},
		{"移除组合位", Bit0 | Bit1 | Bit2, Bit1 | Bit2, Bit0},
		{"移除所有位", Bit0 | Bit1, Bit0 | Bit1, 0},
		{"从空掩码移除", 0, Bit0, 0},
		{"移除最高位", Bit63 | Bit0, Bit63, Bit0},
		{"移除全0", Bit0 | Bit1, 0, Bit0 | Bit1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Remove(tt.mask, tt.value)
			if result != tt.expected {
				t.Errorf("Remove(%d, %d) = %d, expected %d", tt.mask, tt.value, result, tt.expected)
			}
		})
	}
}

func TestHas(t *testing.T) {
	tests := []struct {
		name     string
		mask     uint64
		value    uint64
		expected bool
	}{
		{"包含单个位", Bit0 | Bit1, Bit0, true},
		{"不包含单个位", Bit0, Bit1, false},
		{"包含组合位中的任意一个", Bit0 | Bit1, Bit1 | Bit2, true},
		{"不包含组合位", Bit0, Bit1 | Bit2, false},
		{"空掩码不包含任何位", 0, Bit0, false},
		{"包含所有位", Bit0 | Bit1 | Bit2, Bit0 | Bit1 | Bit2, true},
		{"检查全0", Bit0, 0, false},
		{"从全0检查", 0, Bit0, false},
		{"包含最高位", Bit63, Bit63, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Has(tt.mask, tt.value)
			if result != tt.expected {
				t.Errorf("Has(%d, %d) = %v, expected %v", tt.mask, tt.value, result, tt.expected)
			}
		})
	}
}

func TestHasAll(t *testing.T) {
	tests := []struct {
		name     string
		mask     uint64
		value    uint64
		expected bool
	}{
		{"包含所有单个位", Bit0 | Bit1, Bit0, true},
		{"包含所有组合位", Bit0 | Bit1 | Bit2, Bit0 | Bit1, true},
		{"不包含所有组合位", Bit0 | Bit1, Bit0 | Bit1 | Bit2, false},
		{"完全相等", Bit0 | Bit1, Bit0 | Bit1, true},
		{"空掩码不包含任何位", 0, Bit0, false},
		{"空值总是包含", Bit0, 0, true},
		{"包含最高位", Bit63 | Bit0, Bit63, true},
		{"不包含最高位", Bit0, Bit63, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasAll(tt.mask, tt.value)
			if result != tt.expected {
				t.Errorf("HasAll(%d, %d) = %v, expected %v", tt.mask, tt.value, result, tt.expected)
			}
		})
	}
}

func TestEquals(t *testing.T) {
	tests := []struct {
		name     string
		mask     uint64
		value    uint64
		expected bool
	}{
		{"完全相等", Bit0 | Bit1, Bit0 | Bit1, true},
		{"不相等", Bit0 | Bit1, Bit0 | Bit2, false},
		{"都为空", 0, 0, true},
		{"一个为空", 0, Bit0, false},
		{"单个位相等", Bit0, Bit0, true},
		{"单个位不等", Bit0, Bit1, false},
		{"最高位相等", Bit63, Bit63, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Equals(tt.mask, tt.value)
			if result != tt.expected {
				t.Errorf("Equals(%d, %d) = %v, expected %v", tt.mask, tt.value, result, tt.expected)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		mask     uint64
		expected bool
	}{
		{"空掩码", 0, true},
		{"非空掩码", Bit0, false},
		{"组合位非空", Bit0 | Bit1, false},
		{"最高位非空", Bit63, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmpty(tt.mask)
			if result != tt.expected {
				t.Errorf("IsEmpty(%d) = %v, expected %v", tt.mask, result, tt.expected)
			}
		})
	}
}

func TestToggle(t *testing.T) {
	tests := []struct {
		name     string
		mask     uint64
		value    uint64
		expected uint64
	}{
		{"切换单个位（添加）", Bit0, Bit1, Bit0 | Bit1},
		{"切换单个位（移除）", Bit0 | Bit1, Bit1, Bit0},
		{"切换组合位", Bit0 | Bit1, Bit1 | Bit2, Bit0 | Bit2},
		{"切换已存在的位", Bit0, Bit0, 0},
		{"从空掩码切换", 0, Bit0, Bit0},
		{"切换最高位", Bit0, Bit63, Bit0 | Bit63},
		{"切换全0", Bit0, 0, Bit0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Toggle(tt.mask, tt.value)
			if result != tt.expected {
				t.Errorf("Toggle(%d, %d) = %d, expected %d", tt.mask, tt.value, result, tt.expected)
			}
		})
	}

	// 测试切换两次应该恢复原值
	mask := Bit0 | Bit1
	value := Bit1
	result1 := Toggle(mask, value)
	result2 := Toggle(result1, value)
	if result2 != mask {
		t.Errorf("Toggle twice should restore: Toggle(Toggle(%d, %d), %d) = %d, expected %d",
			mask, value, value, result2, mask)
	}
}

func TestClear(t *testing.T) {
	tests := []struct {
		name     string
		mask     uint64
		expected uint64
	}{
		{"清空空掩码", 0, 0},
		{"清空单个位", Bit0, 0},
		{"清空组合位", Bit0 | Bit1 | Bit2, 0},
		{"清空最高位", Bit63, 0},
		{"清空所有位", ^uint64(0), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Clear(tt.mask)
			if result != tt.expected {
				t.Errorf("Clear(%d) = %d, expected %d", tt.mask, result, tt.expected)
			}
		})
	}
}

func TestUnion(t *testing.T) {
	tests := []struct {
		name     string
		mask1    uint64
		mask2    uint64
		expected uint64
	}{
		{"两个空掩码", 0, 0, 0},
		{"一个空掩码", 0, Bit0, Bit0},
		{"两个单个位", Bit0, Bit1, Bit0 | Bit1},
		{"有重叠的位", Bit0 | Bit1, Bit1 | Bit2, Bit0 | Bit1 | Bit2},
		{"完全不同的位", Bit0, Bit2, Bit0 | Bit2},
		{"完全相同的位", Bit0 | Bit1, Bit0 | Bit1, Bit0 | Bit1},
		{"包含最高位", Bit0, Bit63, Bit0 | Bit63},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Union(tt.mask1, tt.mask2)
			if result != tt.expected {
				t.Errorf("Union(%d, %d) = %d, expected %d", tt.mask1, tt.mask2, result, tt.expected)
			}
			// 验证 Union 和 Add 结果一致
			if result != Add(tt.mask1, tt.mask2) {
				t.Errorf("Union(%d, %d) should equal Add(%d, %d)", tt.mask1, tt.mask2, tt.mask1, tt.mask2)
			}
		})
	}
}

func TestIntersection(t *testing.T) {
	tests := []struct {
		name     string
		mask1    uint64
		mask2    uint64
		expected uint64
	}{
		{"两个空掩码", 0, 0, 0},
		{"一个空掩码", 0, Bit0, 0},
		{"有共同位", Bit0 | Bit1, Bit1 | Bit2, Bit1},
		{"无共同位", Bit0, Bit1, 0},
		{"完全相同的位", Bit0 | Bit1, Bit0 | Bit1, Bit0 | Bit1},
		{"包含关系", Bit0 | Bit1 | Bit2, Bit0 | Bit1, Bit0 | Bit1},
		{"最高位交集", Bit63 | Bit0, Bit63 | Bit1, Bit63},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Intersection(tt.mask1, tt.mask2)
			if result != tt.expected {
				t.Errorf("Intersection(%d, %d) = %d, expected %d", tt.mask1, tt.mask2, result, tt.expected)
			}
		})
	}
}

func TestDifference(t *testing.T) {
	tests := []struct {
		name     string
		mask1    uint64
		mask2    uint64
		expected uint64
	}{
		{"两个空掩码", 0, 0, 0},
		{"从空掩码减去", 0, Bit0, 0},
		{"减去单个位", Bit0 | Bit1, Bit1, Bit0},
		{"减去不存在的位", Bit0, Bit1, Bit0},
		{"减去组合位", Bit0 | Bit1 | Bit2, Bit1 | Bit2, Bit0},
		{"完全减去", Bit0 | Bit1, Bit0 | Bit1, 0},
		{"包含最高位", Bit63 | Bit0, Bit63, Bit0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Difference(tt.mask1, tt.mask2)
			if result != tt.expected {
				t.Errorf("Difference(%d, %d) = %d, expected %d", tt.mask1, tt.mask2, result, tt.expected)
			}
			// 验证 Difference 和 Remove 结果一致
			if result != Remove(tt.mask1, tt.mask2) {
				t.Errorf("Difference(%d, %d) should equal Remove(%d, %d)", tt.mask1, tt.mask2, tt.mask1, tt.mask2)
			}
		})
	}
}

func TestXOR(t *testing.T) {
	tests := []struct {
		name     string
		mask1    uint64
		mask2    uint64
		expected uint64
	}{
		{"两个空掩码", 0, 0, 0},
		{"一个空掩码", 0, Bit0, Bit0},
		{"有共同位", Bit0 | Bit1, Bit1 | Bit2, Bit0 | Bit2},
		{"无共同位", Bit0, Bit1, Bit0 | Bit1},
		{"完全相同的位", Bit0 | Bit1, Bit0 | Bit1, 0},
		{"包含最高位", Bit0, Bit63, Bit0 | Bit63},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := XOR(tt.mask1, tt.mask2)
			if result != tt.expected {
				t.Errorf("XOR(%d, %d) = %d, expected %d", tt.mask1, tt.mask2, result, tt.expected)
			}
			// 验证 XOR 和 Toggle 结果一致
			if result != Toggle(tt.mask1, tt.mask2) {
				t.Errorf("XOR(%d, %d) should equal Toggle(%d, %d)", tt.mask1, tt.mask2, tt.mask1, tt.mask2)
			}
		})
	}

	// 测试 XOR 的对称性：XOR(a, b) == XOR(b, a)
	mask1 := Bit0 | Bit1
	mask2 := Bit1 | Bit2
	if XOR(mask1, mask2) != XOR(mask2, mask1) {
		t.Errorf("XOR should be symmetric: XOR(%d, %d) != XOR(%d, %d)", mask1, mask2, mask2, mask1)
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		name     string
		mask     uint64
		expected int
	}{
		{"空掩码", 0, 0},
		{"单个位", Bit0, 1},
		{"两个位", Bit0 | Bit1, 2},
		{"三个位", Bit0 | Bit1 | Bit2, 3},
		{"最高位", Bit63, 1},
		{"所有位", ^uint64(0), 64},
		{"交替位", 0xAAAAAAAAAAAAAAAA, 32},
		{"连续位", 0x000000000000000F, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Count(tt.mask)
			if result != tt.expected {
				t.Errorf("Count(%d) = %d, expected %d", tt.mask, result, tt.expected)
			}
		})
	}

	// 验证 Count 和 List 结果一致
	mask := Bit0 | Bit1 | Bit2
	if Count(mask) != len(List(mask)) {
		t.Errorf("Count(%d) = %d, but List(%d) has %d elements", mask, Count(mask), mask, len(List(mask)))
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		name     string
		mask     uint64
		expected []uint64
	}{
		{"空掩码", 0, []uint64{}},
		{"单个位", Bit0, []uint64{Bit0}},
		{"两个位", Bit0 | Bit1, []uint64{Bit0, Bit1}},
		{"三个位", Bit0 | Bit1 | Bit2, []uint64{Bit0, Bit1, Bit2}},
		{"不连续位", Bit0 | Bit2, []uint64{Bit0, Bit2}},
		{"最高位", Bit63, []uint64{Bit63}},
		{"多个位", Bit0 | Bit1 | Bit2 | Bit3, []uint64{Bit0, Bit1, Bit2, Bit3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := List(tt.mask)
			if len(result) != len(tt.expected) {
				t.Errorf("List(%d) length = %d, expected %d", tt.mask, len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("List(%d)[%d] = %d, expected %d", tt.mask, i, v, tt.expected[i])
				}
			}
		})
	}

	// 验证 List 结果可以重建原掩码
	mask := Bit0 | Bit1 | Bit2 | Bit3
	list := List(mask)
	reconstructed := uint64(0)
	for _, bit := range list {
		reconstructed = Add(reconstructed, bit)
	}
	if reconstructed != mask {
		t.Errorf("List(%d) reconstruction failed: got %d, expected %d", mask, reconstructed, mask)
	}
}

func TestListIndices(t *testing.T) {
	tests := []struct {
		name     string
		mask     uint64
		expected []int
	}{
		{"空掩码", 0, []int{}},
		{"单个位", Bit0, []int{0}},
		{"两个位", Bit0 | Bit1, []int{0, 1}},
		{"三个位", Bit0 | Bit1 | Bit2, []int{0, 1, 2}},
		{"不连续位", Bit0 | Bit2, []int{0, 2}},
		{"最高位", Bit63, []int{63}},
		{"多个位", Bit0 | Bit1 | Bit2 | Bit3, []int{0, 1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ListIndices(tt.mask)
			if len(result) != len(tt.expected) {
				t.Errorf("ListIndices(%d) length = %d, expected %d", tt.mask, len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("ListIndices(%d)[%d] = %d, expected %d", tt.mask, i, v, tt.expected[i])
				}
			}
		})
	}

	// 验证 ListIndices 和 List 结果一致
	mask := Bit0 | Bit1 | Bit2
	indices := ListIndices(mask)
	list := List(mask)
	if len(indices) != len(list) {
		t.Errorf("ListIndices(%d) and List(%d) have different lengths", mask, mask)
	}
	for i, idx := range indices {
		if list[i] != (uint64(1) << idx) {
			t.Errorf("ListIndices(%d)[%d] = %d, but List(%d)[%d] = %d (expected %d)",
				mask, i, idx, mask, i, list[i], uint64(1)<<idx)
		}
	}
}

// 综合测试：模拟权限管理场景
func TestPermissionScenario(t *testing.T) {
	const (
		PermissionRead   = Bit0
		PermissionWrite  = Bit1
		PermissionDelete = Bit2
		PermissionAdmin  = Bit3
	)

	// 初始用户权限为空
	userPerms := uint64(0)

	// 添加读和写权限
	userPerms = Add(userPerms, PermissionRead|PermissionWrite)
	if !HasAll(userPerms, PermissionRead|PermissionWrite) {
		t.Error("Failed to add read and write permissions")
	}
	if Has(userPerms, PermissionDelete) {
		t.Error("Should not have delete permission")
	}

	// 检查是否有读权限
	if !Has(userPerms, PermissionRead) {
		t.Error("Should have read permission")
	}

	// 添加管理员权限
	userPerms = Add(userPerms, PermissionAdmin)
	if !HasAll(userPerms, PermissionRead|PermissionWrite|PermissionAdmin) {
		t.Error("Failed to add admin permission")
	}

	// 移除写权限
	userPerms = Remove(userPerms, PermissionWrite)
	if Has(userPerms, PermissionWrite) {
		t.Error("Should not have write permission after removal")
	}
	if !Has(userPerms, PermissionRead) {
		t.Error("Should still have read permission")
	}

	// 切换删除权限（添加）
	userPerms = Toggle(userPerms, PermissionDelete)
	if !Has(userPerms, PermissionDelete) {
		t.Error("Should have delete permission after toggle")
	}

	// 切换删除权限（移除）
	userPerms = Toggle(userPerms, PermissionDelete)
	if Has(userPerms, PermissionDelete) {
		t.Error("Should not have delete permission after second toggle")
	}

	// 统计权限数量
	expectedCount := 2 // Read + Admin
	if Count(userPerms) != expectedCount {
		t.Errorf("Permission count = %d, expected %d", Count(userPerms), expectedCount)
	}

	// 列出所有权限
	perms := List(userPerms)
	if len(perms) != expectedCount {
		t.Errorf("Permission list length = %d, expected %d", len(perms), expectedCount)
	}
}

// 综合测试：模拟错误码组合场景
func TestErrorCodeScenario(t *testing.T) {
	const (
		ErrInvalidParam = Bit0
		ErrNotFound     = Bit1
		ErrTimeout      = Bit2
		ErrPermission   = Bit3
	)

	// 创建组合错误
	combinedErr := ErrInvalidParam | ErrNotFound | ErrTimeout

	// 检查是否包含特定错误
	if !Has(combinedErr, ErrInvalidParam) {
		t.Error("Should contain ErrInvalidParam")
	}
	if !Has(combinedErr, ErrNotFound) {
		t.Error("Should contain ErrNotFound")
	}
	if Has(combinedErr, ErrPermission) {
		t.Error("Should not contain ErrPermission")
	}

	// 检查是否包含所有指定的错误
	if !HasAll(combinedErr, ErrInvalidParam|ErrNotFound) {
		t.Error("Should contain both ErrInvalidParam and ErrNotFound")
	}
	if HasAll(combinedErr, ErrInvalidParam|ErrPermission) {
		t.Error("Should not contain both ErrInvalidParam and ErrPermission")
	}

	// 移除超时错误
	combinedErr = Remove(combinedErr, ErrTimeout)
	if Has(combinedErr, ErrTimeout) {
		t.Error("Should not contain ErrTimeout after removal")
	}

	// 统计错误数量
	if Count(combinedErr) != 2 {
		t.Errorf("Error count = %d, expected 2", Count(combinedErr))
	}
}

// 基准测试
func BenchmarkAdd(b *testing.B) {
	mask := Bit0 | Bit1
	value := Bit2 | Bit3
	for i := 0; i < b.N; i++ {
		_ = Add(mask, value)
	}
}

func BenchmarkHas(b *testing.B) {
	mask := Bit0 | Bit1 | Bit2
	value := Bit1 | Bit2
	for i := 0; i < b.N; i++ {
		_ = Has(mask, value)
	}
}

func BenchmarkCount(b *testing.B) {
	mask := Bit0 | Bit1 | Bit2 | Bit3 | Bit63
	for i := 0; i < b.N; i++ {
		_ = Count(mask)
	}
}

func BenchmarkList(b *testing.B) {
	mask := Bit0 | Bit1 | Bit2 | Bit3 | Bit63
	for i := 0; i < b.N; i++ {
		_ = List(mask)
	}
}
