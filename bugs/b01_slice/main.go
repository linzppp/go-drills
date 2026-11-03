package main

import "fmt"

// 取出前 n 个学校作为"重点校"，再追加一个"试点校"
func keySchools(all []string, n int) []string {
	key := all[:n]
	return append(key, "试点校")
}

func main() {
	all := []string{"一中", "二中", "三中", "四中", "五中"}
	key := keySchools(all, 2)
	fmt.Println("重点校:", key)
	fmt.Println("全部学校:", all) // 期望：保持 一中 二中 三中 四中 五中
}
