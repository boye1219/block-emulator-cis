package core

import (
	"container/heap"
	"time"
)

type txItem struct {
	tx    *Transaction
	index int
}

type txPriorityQueue []*txItem

func (pq txPriorityQueue) Len() int { return len(pq) }

func (pq txPriorityQueue) Less(i, j int) bool {
	ti, tj := pq[i].tx, pq[j].tx
	if ti.GasFee != tj.GasFee {
		return ti.GasFee > tj.GasFee
	}
	var tiT, tjT time.Time = ti.Time, tj.Time
	if tiT.IsZero() && !tjT.IsZero() {
		return false
	}
	if !tiT.IsZero() && tjT.IsZero() {
		return true
	}
	return tiT.Before(tjT)
}

func (pq txPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *txPriorityQueue) Push(x any) {
	it := x.(*txItem)
	it.index = len(*pq)
	*pq = append(*pq, it)
}

func (pq *txPriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	it := old[n-1]
	old[n-1] = nil
	it.index = -1
	*pq = old[:n-1]
	return it
}

func (pq *txPriorityQueue) PushTx(tx *Transaction) {
	heap.Push(pq, &txItem{tx: tx})
}
func (pq *txPriorityQueue) PopTx() *Transaction {
	if pq.Len() == 0 {
		return nil
	}
	it := heap.Pop(pq).(*txItem)
	return it.tx
}
func (pq *txPriorityQueue) PeekTx() *Transaction {
	if pq.Len() == 0 {
		return nil
	}
	return (*pq)[0].tx
}
