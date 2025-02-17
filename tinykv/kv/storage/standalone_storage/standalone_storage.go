package standalone_storage

import (
	"github.com/Connor1996/badger"
	"github.com/pingcap-incubator/tinykv/kv/config"
	"github.com/pingcap-incubator/tinykv/kv/storage"
	"github.com/pingcap-incubator/tinykv/kv/util/engine_util"
	"github.com/pingcap-incubator/tinykv/proto/pkg/kvrpcpb"
)

// StandAloneStorage is an implementation of `Storage` for a single-node TinyKV instance. It does not
// communicate with other nodes and all data is stored locally.
type StandAloneStorage struct {
	db *badger.DB
}

func NewStandAloneStorage(conf *config.Config) *StandAloneStorage {
	db := engine_util.CreateDB("kv", conf)
	return &StandAloneStorage{
		db: db,
	}
}

func (s *StandAloneStorage) Stop() error {
	return s.db.Close()
}

func (s *StandAloneStorage) Reader(ctx *kvrpcpb.Context) (storage.StorageReader, error) {
	// YOUR CODE HERE (lab1).
	txn := s.db.NewTransaction(false)
	return NewBadgerReader(txn), nil
}

// Write 将storage.Modify内容通过列族cf中的keys进行修改
func (s *StandAloneStorage) Write(ctx *kvrpcpb.Context, batch []storage.Modify) error {
	// YOUR CODE HERE (lab1).
	// Try to check the definition of `storage.Modify` and txn interface of `badger`.
	// As the column family is not supported by `badger`, a wrapper is used to simulate it.
	//txn := s.db.NewTransaction(true)
	//for _, m := range batch {
	//	// Modify中支持两种操作，put和delete
	//	switch m.Data.(type) {
	//	case storage.Put:
	//		put := m.Data.(storage.Put)
	//		// set(key, vals) 事务txn将键值对放入storage，在put的列族中进行put
	//		if err := txn.Set(engine_util.KeyWithCF(put.Cf, put.Key), put.Value); err != nil {
	//			return err
	//		}
	//	case storage.Delete:
	//		del := m.Data.(storage.Delete)
	//		// 在del的列族中进行删除
	//		if err := txn.Delete(engine_util.KeyWithCF(del.Cf, del.Key)); err != nil {
	//			return err
	//		}
	//	}
	//
	//	err := txn.Commit()
	//	if err != nil {
	//		return err
	//	}
	//}
	//
	//return nil
	writeBatch := engine_util.WriteBatch{}
	for _, m := range batch {
		writeBatch.SetCF(m.Cf(), m.Key(), m.Value())
	}
	return writeBatch.WriteToDB(s.db)

}

type BadgerReader struct {
	txn *badger.Txn
}

func NewBadgerReader(txn *badger.Txn) *BadgerReader {
	return &BadgerReader{txn}
}

func (b *BadgerReader) GetCF(cf string, key []byte) ([]byte, error) {
	val, err := engine_util.GetCFFromTxn(b.txn, cf, key)
	if err == badger.ErrKeyNotFound {
		return nil, nil
	}
	return val, err
}

func (b *BadgerReader) IterCF(cf string) engine_util.DBIterator {
	return engine_util.NewCFIterator(cf, b.txn)
}

func (b *BadgerReader) Close() {
	b.txn.Discard()
}
