package misaki

import (
	"time"
)

type Mediator struct {
	current int
	initial int
	maximum int
}

func NewMediator(initial, maximum int) *Mediator {
	return &Mediator{
		current: initial,
		initial: initial,
		maximum: maximum,
	}
}

func (m *Mediator) GetCurrent() int {
	return m.current
}

func (m *Mediator) Reset() {
	m.current = m.initial
}

func (m *Mediator) Wait() {
	time.Sleep(time.Duration(m.current) * time.Second)
}

func (m *Mediator) Increment() {
	m.current = m.current * 2
	if m.current > m.maximum {
		m.current = m.maximum
	}
}
