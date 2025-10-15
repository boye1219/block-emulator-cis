// txpool.go
// the define and some operation of txpool
package core

import (
	// "blockEmulator/utils"
	"container/heap"
	"sync"
	"math/big"
	"log"
	// "unsafe"
)

type TxPool struct {
	TxQueue   txPriorityQueue           // Transaction Priority Queue
	RelayPool map[uint64][]*Transaction // Designed for sharded blockchain, from Monoxide
	lock      sync.Mutex
	// The pending list is ignored
}

func NewTxPool() *TxPool {
	p := &TxPool{
		// TxQueue:   make([]*Transaction, 0),
		TxQueue:   make(txPriorityQueue, 0),
		RelayPool: make(map[uint64][]*Transaction),
	}
	heap.Init(&p.TxQueue)
	return p
}

// Add a transaction to the pool (consider the queue only)
// core/txpool.go
func (txpool *TxPool) AddTx2Pool(tx *Transaction) {
    txpool.lock.Lock()
    defer txpool.lock.Unlock()
    if tx.GasFee == nil {
        if tx.GasPrice != nil {
            tx.GasFee = new(big.Int).Mul(tx.GasPrice, new(big.Int).SetUint64(tx.GasUsed))
        } else {
            tx.GasFee = new(big.Int) // 視為 0
        }
    }
    txpool.TxQueue.PushTx(tx)
}

// Add a list of transactions to the pool
func (txpool *TxPool) AddTxs2Pool(txs []*Transaction) {
    txpool.lock.Lock()
    defer txpool.lock.Unlock()
    for _, tx := range txs {
        if tx.GasFee == nil {
            if tx.GasPrice != nil {
                tx.GasFee = new(big.Int).Mul(tx.GasPrice, new(big.Int).SetUint64(tx.GasUsed))
            } else {
                tx.GasFee = new(big.Int)
            }
        }
        txpool.TxQueue.PushTx(tx)
    }
}


// add transactions into the pool head
// func (txpool *TxPool) AddTxs2Pool_Head(tx []*Transaction) {
// 	txpool.lock.Lock()
// 	defer txpool.lock.Unlock()
// 	txpool.TxQueue = append(tx, txpool.TxQueue...)
// }

// Pack transactions for a proposal
func (txpool *TxPool) PackTxs(max_txs uint64) []*Transaction {
	top := txpool.TxQueue.PeekTx()
	if top != nil { 
		log.Printf("Top GasFee=%s GasPrice=%s GasUsed=%d", top.GasFee, top.GasPrice, top.GasUsed) 
	}

	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	txNum := max_txs
	if uint64(txpool.TxQueue.Len()) < txNum {
		txNum = uint64(txpool.TxQueue.Len())
	}

	out := make([]*Transaction, 0, txNum)
	for i := 0; i < int(txNum); i++ {
		tx := txpool.TxQueue.PopTx()
		if tx == nil {
			break
		}
		out = append(out, tx)
	}
	return out

}

// Pack transaction for a proposal (use 'BlocksizeInBytes' to control)
// func (txpool *TxPool) PackTxsWithBytes(max_bytes int) []*Transaction {
// 	txpool.lock.Lock()
// 	defer txpool.lock.Unlock()

// 	txNum := len(txpool.TxQueue)
// 	currentSize := 0
// 	for tx_idx, tx := range txpool.TxQueue {
// 		currentSize += int(unsafe.Sizeof(*tx))
// 		if currentSize > max_bytes {
// 			txNum = tx_idx
// 			break
// 		}
// 	}

// 	txs_Packed := txpool.TxQueue[:txNum]
// 	txpool.TxQueue = txpool.TxQueue[txNum:]
// 	return txs_Packed
// }

// Relay transactions
func (txpool *TxPool) AddRelayTx(tx *Transaction, shardID uint64) {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	_, ok := txpool.RelayPool[shardID]
	if !ok {
		txpool.RelayPool[shardID] = make([]*Transaction, 0)
	}
	txpool.RelayPool[shardID] = append(txpool.RelayPool[shardID], tx)
}

// txpool get locked
func (txpool *TxPool) GetLocked() {
	txpool.lock.Lock()
}

// txpool get unlocked
func (txpool *TxPool) GetUnlocked() {
	txpool.lock.Unlock()
}

// get the length of tx queue
func (txpool *TxPool) GetTxQueueLen() int {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	return txpool.TxQueue.Len()
}

// get the length of ClearRelayPool
func (txpool *TxPool) ClearRelayPool() {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	txpool.RelayPool = nil
}

// abort ! Pack relay transactions from relay pool
func (txpool *TxPool) PackRelayTxs(shardID, minRelaySize, maxRelaySize uint64) ([]*Transaction, bool) {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	if _, ok := txpool.RelayPool[shardID]; !ok {
		return nil, false
	}
	if len(txpool.RelayPool[shardID]) < int(minRelaySize) {
		return nil, false
	}
	txNum := maxRelaySize
	if uint64(len(txpool.RelayPool[shardID])) < txNum {
		txNum = uint64(len(txpool.RelayPool[shardID]))
	}
	relayTxPacked := txpool.RelayPool[shardID][:txNum]
	txpool.RelayPool[shardID] = txpool.RelayPool[shardID][txNum:]
	return relayTxPacked, true
}

// abort ! Transfer transactions when re-sharding
/*
func (txpool *TxPool) TransferTxs(addr utils.Address) []*Transaction {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	txTransfered := make([]*Transaction, 0)
	newTxQueue := make([]*Transaction, 0)
	for _, tx := range txpool.TxQueue {
		if tx.Sender == addr {
			txTransfered = append(txTransfered, tx)
		} else {
			newTxQueue = append(newTxQueue, tx)
		}
	}
	newRelayPool := make(map[uint64][]*Transaction)
	for shardID, shardPool := range txpool.RelayPool {
		for _, tx := range shardPool {
			if tx.Sender == addr {
				txTransfered = append(txTransfered, tx)
			} else {
				if _, ok := newRelayPool[shardID]; !ok {
					newRelayPool[shardID] = make([]*Transaction, 0)
				}
				newRelayPool[shardID] = append(newRelayPool[shardID], tx)
			}
		}
	}
	txpool.TxQueue = newTxQueue
	txpool.RelayPool = newRelayPool
	return txTransfered
}
*/
