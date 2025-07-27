package tin

import "math"

type Data struct {
	data  [2]float64
	lface interface{}
}

var (
	NAN_DATA = Data{data: [2]float64{math.NaN(), math.NaN()}, lface: nil}
)

type Pool struct {
	Values []interface{}
	New    func() interface{}
}

func NewPool(newFn func() interface{}) *Pool {
	return &Pool{New: newFn}
}

func (p *Pool) clear(index int) {
	p.Values[index] = nil
}

func New(p *Pool) *QuadEdge {
	qe := &QuadEdge{pool: p, index: len(p.Values)}
	p.Values = append(p.Values, qe)
	return qe
}

func (p *Pool) Len() int {
	return len(p.Values)
}

func (p *Pool) Get() interface{} {
	if len(p.Values) > 0 {
		item := p.Values[len(p.Values)-1]
		p.Values = p.Values[:len(p.Values)-1]
		return item
	}
	return p.New()
}

func (p *Pool) Put(item interface{}) {
	p.Values = append(p.Values, item)
}
