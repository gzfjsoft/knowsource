package utils

// AdvancedTreeUtils 功能增强的树工具

type TreeNode struct {
	Id       int64      `json:"Id"`
	ParentId int64      `json:"parentId"`
	Level    int64      `json:"level"`
	Name     string     `json:"name"`
	Code     string     `json:"code"`
	Sort     int64      `json:"sort"`
	Children []TreeNode `json:"children,omitempty"`
}

// FindNodeAndChildrenIds 查找节点及其子节点的Id集合
func FindNodeAndChildrenIds(root TreeNode, targetId int64, includeSelf bool) []int64 {
	targetNode := FindNodeById(root, targetId)
	if targetNode.Id == 0 && len(targetNode.Children) == 0 {
		return []int64{}
	}

	if !includeSelf {
		return collectChildrenIdsOnly(targetNode)
	}

	return collectAllChildrenIds(targetNode)
}

// FindNodeById 查找指定Id的节点
func FindNodeById(root TreeNode, targetId int64) TreeNode {
	stack := []TreeNode{root}

	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if node.Id == targetId {
			return node
		}

		for i := len(node.Children) - 1; i >= 0; i-- {
			stack = append(stack, node.Children[i])
		}
	}

	return TreeNode{}
}

// collectAllChildrenIds 收集节点及其所有子节点的Id
func collectAllChildrenIds(root TreeNode) []int64 {
	result := []int64{}
	stack := []TreeNode{root}

	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		result = append(result, node.Id)

		for i := len(node.Children) - 1; i >= 0; i-- {
			stack = append(stack, node.Children[i])
		}
	}

	return result
}

// collectChildrenIdsOnly 只收集子节点的Id（不包含自身）
func collectChildrenIdsOnly(root TreeNode) []int64 {
	result := []int64{}
	stack := make([]TreeNode, len(root.Children))
	copy(stack, root.Children)

	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		result = append(result, node.Id)

		for i := len(node.Children) - 1; i >= 0; i-- {
			stack = append(stack, node.Children[i])
		}
	}

	return result
}
