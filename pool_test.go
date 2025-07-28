package tin

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPool(t *testing.T) {
	t.Run("基本功能测试", func(t *testing.T) {
		// 创建一个计数器，用于跟踪创建的对象数量
		counter := 0
		pool := NewPool(func() interface{} {
			counter++
			val := counter
			return &val // 返回独立变量的指针
		})

		// 初始状态验证
		assert.Equal(t, 0, pool.Len(), "初始池长度应为0")

		// 第一次获取对象
		obj1 := pool.Get()
		assert.Equal(t, 1, counter, "应创建第一个对象")
		assert.Equal(t, 1, *(obj1.(*int)), "对象值应为1")

		// 第二次获取对象
		obj2 := pool.Get()
		assert.Equal(t, 2, counter, "应创建第二个对象")
		assert.Equal(t, 2, *(obj2.(*int)), "对象值应为2")

		// 将对象放回池中
		pool.Put(obj1)
		assert.Equal(t, 1, pool.Len(), "放回一个对象后池长度应为1")

		pool.Put(obj2)
		assert.Equal(t, 2, pool.Len(), "放回两个对象后池长度应为2")

		// 再次获取对象 - 应重用之前放回的对象
		obj3 := pool.Get()
		assert.Equal(t, 2, counter, "不应创建新对象")
		assert.Equal(t, 2, *(obj3.(*int)), "应获取最后放回的对象")
		assert.Equal(t, 1, pool.Len(), "获取一个对象后池长度应为1")

		obj4 := pool.Get()
		assert.Equal(t, 2, counter, "不应创建新对象")
		assert.Equal(t, 1, *(obj4.(*int)), "应获取第一个放回的对象")
		assert.Equal(t, 0, pool.Len(), "池应为空")
	})

	t.Run("池为空时创建新对象", func(t *testing.T) {
		counter := 0
		pool := NewPool(func() interface{} {
			counter++
			return counter
		})

		// 池为空时获取对象
		obj := pool.Get()
		assert.Equal(t, 1, obj.(int), "应创建新对象")
		assert.Equal(t, 0, pool.Len(), "池长度应为0")
	})

	t.Run("清空池", func(t *testing.T) {
		pool := NewPool(func() interface{} {
			return "new"
		})

		// 添加多个对象
		pool.Put("item1")
		pool.Put("item2")
		pool.Put("item3")
		assert.Equal(t, 3, pool.Len(), "池长度应为3")

		// 清空特定索引
		pool.clear(1)
		assert.Equal(t, 3, pool.Len(), "池长度不变")
		assert.Nil(t, pool.Values[1], "索引1应为nil")

		// 获取对象时应跳过nil
		item1 := pool.Get()
		assert.Equal(t, "item3", item1, "应获取最后一个对象")

		item2 := pool.Get()
		assert.Equal(t, "item1", item2, "应获取第一个对象") // 恢复原始断言

		item3 := pool.Get()
		assert.Equal(t, "new", item3, "应创建新对象")
	})

	t.Run("不同类型对象", func(t *testing.T) {
		pool := NewPool(func() interface{} {
			return struct{ value int }{value: 42}
		})

		// 获取新对象
		obj := pool.Get()
		assert.Equal(t, 42, obj.(struct{ value int }).value, "对象值应为42")

		// 放回后重用
		pool.Put(obj)
		reused := pool.Get()
		assert.Equal(t, 42, reused.(struct{ value int }).value, "应重用对象")
	})

	t.Run("NAN_DATA 测试", func(t *testing.T) {
		nanData := NAN_DATA
		assert.True(t, math.IsNaN(nanData.data[0]), "NAN_DATA 的第一个元素应为 NaN")
		assert.True(t, math.IsNaN(nanData.data[1]), "NAN_DATA 的第二个元素应为 NaN")
		assert.Nil(t, nanData.lface, "NAN_DATA 的 lface 应为 nil")
	})

	t.Run("New 函数测试", func(t *testing.T) {
		pool := NewPool(func() interface{} {
			return "default"
		})

		// 使用 New 函数创建 QuadEdge
		qe := New(pool)
		assert.NotNil(t, qe, "New 函数应返回非 nil 对象")
		assert.Equal(t, 0, qe.index, "新对象的索引应为0")
		assert.Equal(t, 1, pool.Len(), "池长度应为1")
	})
}
