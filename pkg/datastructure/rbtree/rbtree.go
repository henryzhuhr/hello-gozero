// Package rbtree 红黑树
package rbtree

// ---------------------- Color 树的节点颜色 ----------------------

// Color 树的节点颜色
type Color bool

const (
	Black Color = false
	Red   Color = true
)

// String 方法便于调试
func (c Color) String() string {
	if c == Red {
		return "Red"
	}
	return "Black"
}

// ---------------------- Node 是红黑树的节点 ----------------------

// node 是红黑树的内部节点（小写表示不导出）
type node[K comparable, V any] struct {
	key    K
	value  V
	color  Color
	left   *node[K, V]
	right  *node[K, V]
	parent *node[K, V]
}

// newNode 创建新节点（默认红色）
func newNode[K comparable, V any](key K, value V) *node[K, V] {
	return &node[K, V]{
		key:   key,
		value: value,
		color: Red, // 新插入节点默认为红色
	}
}

// ---------------------- RBTree 红黑树结构 ----------------------

// RBTree 红黑树（Red-Black Tree）
// 是一种自平衡的二叉查找树（Binary Search Tree, BST），
// 它在插入和删除操作后通过特定规则自动保持树的近似平衡，
// 从而保证基本操作（如查找、插入、删除）的时间复杂度为 O(log n)。
//
// 一棵红黑树必须满足以下五条性质：
//  1. 每个节点要么是红色，要么是黑色。
//  2. 根节点是黑色。
//  3. 所有叶子节点（NIL 节点，即空指针）。
//  4. 如果一个节点是红色的，则它的两个子节点都必须是黑色的（即不能有两个连续的红色节点）。
//  5. 从任一节点到其每个叶子的所有路径都包含相同数量的黑色节点（称为“黑高”相同）。
//
// 注：实际实现中，通常用一个共享的哨兵节点（如 nil）代表所有叶子（NIL），以节省空间。
//
// 为什么需要红黑树？
//
// 普通的二叉搜索树在最坏情况下（如按顺序插入）会退化成链表，导致操作时间复杂度变为 O(n)。
// 红黑树通过上述约束，确保树的高度始终不超过 2 log₂(n+1)，从而维持高效性能。
type RBTree[K comparable, V any] interface {
	// Insert 插入键值对
	Insert(key K, value V)

	// Delete(key K) bool
	// Search(key K) (value V, found bool)
	// InOrder(fn func(key K, value V))
	// Min() (key K, value V, ok bool)
	// Max() (key K, value V, ok bool)
	// Len() int // 可选：返回元素个数

	// Print 打印树结构（用于调试）
	// Print()
}

// rbTree 是 RBTree 接口的具体实现（小写结构体名，通过 New 返回接口）
type rbTree[K comparable, V any] struct {
	root  *node[K, V]
	nil   *node[K, V]       // 哨兵 NIL 节点（黑色）
	less  func(a, b K) bool // 比较函数：less(a, b) == true 表示 a < b
	count int               // 可选：记录节点数量
}

// NewRBTree 创建一个新的红黑树
//
// less 是比较函数：less(a, b) == true 表示 a < b
func NewRBTree[K comparable, V any](less func(a, b K) bool) RBTree[K, V] {
	nilNode := &node[K, V]{
		color:  Black,
		left:   nil, // 实际会在使用时指向自己，但初始为 nil 也可
		right:  nil,
		parent: nil,
	}
	nilNode.left = nilNode // 可选：让哨兵自指（更严谨）
	nilNode.right = nilNode
	nilNode.parent = nilNode

	return &rbTree[K, V]{
		root:  nilNode,
		nil:   nilNode,
		less:  less,
		count: 0,
	}
}

// Insert implements [RBTree.Insert].
func (r *rbTree[K, V]) Insert(key K, value V) {
	// 创建新节点
	z := &node[K, V]{
		key:    key,
		value:  value,
		color:  Red,
		left:   r.nil,
		right:  r.nil,
		parent: r.nil,
	}

	// 标准 BST 插入
	var y = r.nil
	x := r.root
	for x != r.nil {
		y = x
		if r.less(key, x.key) {
			x = x.left
		} else if r.less(x.key, key) {
			x = x.right
		} else {
			// Key already exists: update value
			x.value = value
			return
		}
	}

	z.parent = y
	if y == r.nil {
		r.root = z
	} else if r.less(z.key, y.key) {
		y.left = z
	} else {
		y.right = z
	}

	// 修复红黑树性质
	r.insertFixup(z)
	r.count++
}

// ---------------------- RBTree 辅助函数 ----------------------
// leftRotate 对节点 x 进行左旋
func (r *rbTree[K, V]) leftRotate(x *node[K, V]) {
	y := x.right
	x.right = y.left
	if y.left != r.nil {
		y.left.parent = x
	}
	y.parent = x.parent
	if x.parent == r.nil {
		r.root = y
	} else if x == x.parent.left {
		x.parent.left = y
	} else {
		x.parent.right = y
	}
	y.left = x
	x.parent = y
}

// rightRotate 对节点 y 进行右旋
func (r *rbTree[K, V]) rightRotate(y *node[K, V]) {
	x := y.left
	y.left = x.right
	if x.right != r.nil {
		x.right.parent = y
	}
	x.parent = y.parent
	if y.parent == r.nil {
		r.root = x
	} else if y == y.parent.right {
		y.parent.right = x
	} else {
		y.parent.left = x
	}
	x.right = y
	y.parent = x
}

// insertFixup 修复插入后可能违反的红黑树性质
func (r *rbTree[K, V]) insertFixup(z *node[K, V]) {
	for z.parent.color == Red {
		if z.parent == z.parent.parent.left {
			y := z.parent.parent.right // uncle
			if y.color == Red {
				// Case 1: uncle is red
				z.parent.color = Black
				y.color = Black
				z.parent.parent.color = Red
				z = z.parent.parent
			} else {
				if z == z.parent.right {
					// Case 2: uncle is black and z is right child
					z = z.parent
					r.leftRotate(z)
				}
				// Case 3: uncle is black and z is left child
				z.parent.color = Black
				z.parent.parent.color = Red
				r.rightRotate(z.parent.parent)
			}
		} else {
			// Mirror of above: parent is right child
			y := z.parent.parent.left // uncle
			if y.color == Red {
				// Case 1
				z.parent.color = Black
				y.color = Black
				z.parent.parent.color = Red
				z = z.parent.parent
			} else {
				if z == z.parent.left {
					// Case 2
					z = z.parent
					r.rightRotate(z)
				}
				// Case 3
				z.parent.color = Black
				z.parent.parent.color = Red
				r.leftRotate(z.parent.parent)
			}
		}
	}
	r.root.color = Black // 确保根始终为黑色
}
