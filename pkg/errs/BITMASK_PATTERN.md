# 位掩码快速上手：四个运算符 + 示例表

| 作用       | 符号  | 记忆法                 | 示例（结果）                                                |
|------------|-------|------------------------|------------------------------------------------------------|
| 设位/组合  | `a|b` | 两边对应位只要有一个 1 就是 1   | `0b0010 \| 0b0100 = 0b0110` （把 2 和 4 都打开）           |
| 检查交集   | `a&b` | 两边对应位两个都是 1 才是 1 | `0b0110 & 0b0010 = 0b0010` （结果非 0，说明包含 0b0010）   |
| 切换       | `a^b` | 两边对应位不一样就是 1      | `0b0110 ^ 0b0010 = 0b0100` （第 2 位从 1 变 0，其它不变）  |
| 清零       | `a&^b`| 把 a 中 b 为 1 的位清 0 | `0b0110 &^ 0b0010 = 0b0100` （把 2 清掉，剩下 4）          |

## 常用权限操作函数示例

```go
// 定义权限标志
const (
    PermissionRead   uint64 = 1 << 0  // 0001: 读权限
    PermissionWrite  uint64 = 1 << 1  // 0010: 写权限
    PermissionDelete uint64 = 1 << 2  // 0100: 删除权限
    PermissionAdmin  uint64 = 1 << 3  // 1000: 管理员权限
)

// 1. 添加权限（单个或组合）
// 原理：使用 | 操作，只要有一个 1 就是 1
func AddPermission(flags uint64, perm uint64) uint64 {
    return flags | perm  // 添加单个或组合权限
}

// 示例：
// flags = 0b0001 (只有读权限)
// AddPermission(flags, PermissionWrite) = 0b0011 (读+写)
// AddPermission(flags, PermissionWrite|PermissionDelete) = 0b0111 (读+写+删除)

// 2. 删除权限（单个或组合）
// 原理：使用 &^ 操作，把 a 中 b 为 1 的位清 0
func RemovePermission(flags uint64, perm uint64) uint64 {
    return flags &^ perm  // 删除单个或组合权限
}

// 示例：
// flags = 0b0111 (读+写+删除)
// RemovePermission(flags, PermissionWrite) = 0b0101 (读+删除)
// RemovePermission(flags, PermissionWrite|PermissionDelete) = 0b0001 (只剩读)

// 3. 判断是否包含权限（单个或组合）
// 原理：使用 & 操作，两个都是 1 才是 1
func HasPermission(flags uint64, perm uint64) bool {
    return flags & perm != 0  // 包含单个或组合中的任意一个
}

// 示例：
// flags = 0b0111 (读+写+删除)
// HasPermission(flags, PermissionRead) = true (包含读)
// HasPermission(flags, PermissionWrite|PermissionDelete) = true (包含写或删除)
// HasPermission(flags, PermissionAdmin) = false (不包含管理员)

// 4. 判断是否包含所有指定权限（必须全部包含）
// 原理：使用 & 操作，只保留目标位后，结果是否等于目标本身
func HasAllPermissions(flags uint64, perm uint64) bool {
    return flags & perm == perm  // 必须全部包含
}

// 示例：
// flags = 0b0111 (读+写+删除)
// HasAllPermissions(flags, PermissionRead|PermissionWrite) = true (同时有读和写)
// HasAllPermissions(flags, PermissionRead|PermissionAdmin) = false (没有管理员权限)

// 5. 切换权限（单个或组合）
// 原理：使用 ^ 操作，不一样就是 1
func TogglePermission(flags uint64, perm uint64) uint64 {
    return flags ^ perm  // 切换单个或组合权限
}

// 示例：
// flags = 0b0011 (读+写)
// TogglePermission(flags, PermissionWrite) = 0b0001 (写被关闭)
// TogglePermission(flags, PermissionDelete) = 0b0111 (删除被打开)
// TogglePermission(flags, PermissionWrite|PermissionDelete) = 0b0101 (写关闭，删除打开)

// 6. 清空所有权限
// 原理：直接赋值为 0
func ClearAllPermissions(flags uint64) uint64 {
    return 0  // 清空所有
}

// 示例：
// flags = 0b0111 (读+写+删除)
// ClearAllPermissions(flags) = 0b0000 (全部清空)

// 7. 获取所有设置的权限列表（辅助函数）
// 原理：遍历每一位，检查是否为 1
func GetSetPermissions(flags uint64) []uint64 {
    var result []uint64
    for i := 0; i < 64; i++ {
        bit := uint64(1 << i)
        if flags & bit != 0 {  // 检查这一位是否为 1
            result = append(result, bit)
        }
    }
    return result
}

// 示例：
// flags = 0b0111 (读+写+删除)
// GetSetPermissions(flags) = [0b0001, 0b0010, 0b0100] (读、写、删除)

// 8. 检查是否只包含指定的权限（没有其他权限）
// 原理：检查 flags 是否完全等于目标权限
func OnlyHasPermissions(flags uint64, perm uint64) bool {
    return flags == perm  // 完全相等
}

// 示例：
// flags = 0b0011 (读+写)
// OnlyHasPermissions(flags, PermissionRead|PermissionWrite) = true (只有读和写)
// OnlyHasPermissions(flags, PermissionRead) = false (还有写权限)

// 9. 批量添加权限（从另一个权限集合添加）
// 原理：使用 | 操作合并
func MergePermissions(flags1 uint64, flags2 uint64) uint64 {
    return flags1 | flags2  // 合并两个权限集合
}

// 示例：
// flags1 = 0b0001 (读)
// flags2 = 0b0110 (写+删除)
// MergePermissions(flags1, flags2) = 0b0111 (读+写+删除)

// 10. 获取权限的交集（共同拥有的权限）
// 原理：使用 & 操作，两个都是 1 才是 1
func IntersectPermissions(flags1 uint64, flags2 uint64) uint64 {
    return flags1 & flags2  // 取交集
}

// 示例：
// flags1 = 0b0111 (读+写+删除)
// flags2 = 0b0011 (读+写)
// IntersectPermissions(flags1, flags2) = 0b0011 (共同有读+写)
```

## 完整使用示例

```go
package main

import "fmt"

func main() {
    // 定义权限
    const (
        PermissionRead   uint64 = 1 << 0
        PermissionWrite  uint64 = 1 << 1
        PermissionDelete uint64 = 1 << 2
        PermissionAdmin  uint64 = 1 << 3
    )
    
    var userPerms uint64
    
    // 1. 添加权限
    userPerms = AddPermission(userPerms, PermissionRead|PermissionWrite)
    fmt.Printf("添加后: %b\n", userPerms)  // 0011
    
    // 2. 检查权限
    fmt.Println("有读权限:", HasPermission(userPerms, PermissionRead))  // true
    fmt.Println("有管理员:", HasPermission(userPerms, PermissionAdmin))  // false
    
    // 3. 检查全部包含
    fmt.Println("同时有读写:", HasAllPermissions(userPerms, PermissionRead|PermissionWrite))  // true
    
    // 4. 删除权限
    userPerms = RemovePermission(userPerms, PermissionWrite)
    fmt.Printf("删除写后: %b\n", userPerms)  // 0001
    
    // 5. 切换权限
    userPerms = TogglePermission(userPerms, PermissionDelete)
    fmt.Printf("切换删除后: %b\n", userPerms)  // 0101
    
    // 6. 获取所有权限
    perms := GetSetPermissions(userPerms)
    fmt.Println("所有权限:", perms)  // [1, 4]
    
    // 7. 清空所有
    userPerms = ClearAllPermissions(userPerms)
    fmt.Printf("清空后: %b\n", userPerms)  // 0000
}
```

