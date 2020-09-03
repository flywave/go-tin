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
}

func NewPool() *Pool {
	return &Pool{}
}

func (p *Pool) clear(index int) {
	p.Values[index] = nil
}

func New(p *Pool) *QuadEdge {
	qe := &QuadEdge{pool: p, index: len(p.Values)}
	p.Values = append(p.Values, qe)
	return qe
}
