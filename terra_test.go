package tin

import (
	"container/heap"
	"testing"
)

func TestPriorityQueue(t *testing.T) {
	pq := make(PQ, 0)
	heap.Init(&pq)

	// 添加候选点
	candidates := []*Candidate{
		{Importance: 5},
		{Importance: 10},
		{Importance: 3},
	}
	for _, c := range candidates {
		heap.Push(&pq, c)
	}

	// 验证最大堆特性
	if pq[0].Importance != 10 {
		t.Errorf("Expected highest importance 10, got %f", pq[0].Importance)
	}

	// 测试更新优先级
	pq.Update(candidates[2], 15)
	if pq[0] != candidates[2] {
		t.Error("Update failed, max importance not updated")
	}

	// 测试弹出顺序
	expected := []float64{15, 10, 5}
	for _, imp := range expected {
		c := heap.Pop(&pq).(*Candidate)
		if c.Importance != imp {
			t.Errorf("Expected %f, got %f", imp, c.Importance)
		}
	}
}

func TestCandidateList(t *testing.T) {
	cl := &CandidateList{}
	cl.Push(&Candidate{Importance: 8})
	cl.Push(&Candidate{Importance: 12})

	if cl.Size() != 2 {
		t.Errorf("Expected size 2, got %d", cl.Size())
	}

	// Test grabbing all elements
	c1 := cl.GrabGreatest()
	if c1.Importance != 12 {
		t.Error("Did not get highest importance candidate")
	}

	c2 := cl.GrabGreatest()
	if c2.Importance != 8 {
		t.Error("Did not get second candidate")
	}

	if !cl.Empty() {
		t.Error("List should be empty after grabbing all")
	}
}
