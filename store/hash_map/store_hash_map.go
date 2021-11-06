package hash_map

import (
	"encoding/binary"
	"errors"
	"math/rand"
	"pandora-pay/helpers"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type HashMap struct {
	name           string
	Tx             store_db_interface.StoreDBTransactionInterface
	Count          uint64
	CountCommitted uint64
	Changes        map[string]*ChangesMapElement
	Committed      map[string]*CommittedMapElement
	KeyLength      int
	Deserialize    func([]byte, []byte) (helpers.SerializableInterface, error)
	DeletedEvent   func([]byte) error
	StoredEvent    func([]byte, *CommittedMapElement) error
	Indexable      bool
}

//support only for commited data
func (hashMap *HashMap) GetIndexByKey(key string) (uint64, error) {
	if !hashMap.Indexable {
		return 0, errors.New("HashMap is not Indexable")
	}

	//safe to Get because it won't change
	data := hashMap.Tx.Get(hashMap.name + ":listKey:" + key)
	if data == nil {
		return 0, errors.New("Key not found")
	}

	return strconv.ParseUint(string(data), 10, 64)
}

//support only for commited data
func (hashMap *HashMap) GetKeyByIndex(index uint64) ([]byte, error) {
	if !hashMap.Indexable {
		return nil, errors.New("HashMap is not Indexable")
	}

	if index > hashMap.CountCommitted {
		return nil, errors.New("Index exceeds count")
	}

	//Clone require because key might get altered afterwards
	key := hashMap.Tx.Get(hashMap.name + ":list:" + strconv.FormatUint(index, 10))
	if key == nil {
		return nil, errors.New("Not found")
	}

	return key, nil
}

//support only for commited data
func (hashMap *HashMap) GetByIndex(index uint64) (data helpers.SerializableInterface, err error) {

	key, err := hashMap.GetKeyByIndex(index)
	if err != nil {
		return nil, err
	}

	return hashMap.Get(string(key))
}

//support only for commited data
func (hashMap *HashMap) GetRandom() (data helpers.SerializableInterface, err error) {
	if !hashMap.Indexable {
		return nil, errors.New("HashMap is not Indexable")
	}

	index := rand.Uint64() % hashMap.Count
	return hashMap.GetByIndex(index)
}

func (hashMap *HashMap) CloneCommitted() (err error) {

	for key, v := range hashMap.Committed {
		if v.Element != nil {
			if v.Element, err = hashMap.Deserialize([]byte(key), helpers.CloneBytes(v.Element.SerializeToBytes())); err != nil {
				return
			}
		}
	}

	return
}

func (hashMap *HashMap) Get(key string) (out helpers.SerializableInterface, err error) {

	if hashMap.KeyLength != 0 && len(key) != hashMap.KeyLength {
		return nil, errors.New("key length is invalid")
	}
	if exists := hashMap.Changes[key]; exists != nil {
		return exists.Element, nil
	}

	var outData []byte

	if exists2 := hashMap.Committed[key]; exists2 != nil {
		if exists2.Element != nil {
			outData = helpers.CloneBytes(exists2.Element.SerializeToBytes())
		}
	} else {
		//Clone required because data could be altered afterwards
		outData = hashMap.Tx.Get(hashMap.name + ":map:" + key)
	}

	if outData != nil {
		if out, err = hashMap.Deserialize([]byte(key), outData); err != nil {
			return nil, err
		}
	}
	hashMap.Changes[key] = &ChangesMapElement{out, "view", 0}
	return
}

func (hashMap *HashMap) Exists(key string) (bool, error) {

	if hashMap.KeyLength != 0 && len(key) != hashMap.KeyLength {
		return false, errors.New("key length is invalid")
	}
	if exists := hashMap.Changes[key]; exists != nil {
		return exists.Element != nil, nil
	}
	if exists := hashMap.Committed[key]; exists != nil {
		return exists.Element != nil, nil
	}

	return hashMap.Tx.Exists(hashMap.name + ":exists:" + key), nil
}

func (hashMap *HashMap) Update(key string, data helpers.SerializableInterface) error {

	if hashMap.KeyLength != 0 && len(key) != hashMap.KeyLength {
		return errors.New("key length is invalid")
	}
	if data == nil {
		return errors.New("Data is null and it should not be")
	}

	if err := data.Validate(); err != nil {
		return err
	}

	exists := hashMap.Changes[key]

	bEmpty := exists == nil || exists.Element == nil

	if exists == nil {
		exists = new(ChangesMapElement)
		hashMap.Changes[key] = exists
	}
	exists.Status = "update"
	exists.Element = data

	if bEmpty && hashMap.Committed[key] == nil && !hashMap.Tx.Exists(hashMap.name+":exists:"+key) {
		exists.index = hashMap.Count
		hashMap.Count += 1
	}

	return nil
}

func (hashMap *HashMap) Delete(key string) {
	exists := hashMap.Changes[key]

	bEmpty := exists != nil && exists.Element != nil

	if exists == nil {
		exists = new(ChangesMapElement)
		hashMap.Changes[key] = exists
	}
	exists.Status = "del"
	exists.Element = nil

	if bEmpty || hashMap.Committed[key] != nil || hashMap.Tx.Exists(hashMap.name+":exists:"+key) {
		exists.index = hashMap.Count
		hashMap.Count -= 1
	}

	return
}

func (hashMap *HashMap) UpdateOrDelete(key string, data helpers.SerializableInterface) error {
	if data == nil {
		hashMap.Delete(key)
		return nil
	}
	return hashMap.Update(key, data)
}

func (hashMap *HashMap) CommitChanges() (err error) {

	removed := make([]string, len(hashMap.Changes))

	c := 0
	for k, v := range hashMap.Changes {
		if hashMap.KeyLength != 0 && len(k) != hashMap.KeyLength {
			return errors.New("key length is invalid")
		}
		if v.Status == "update" {
			removed[c] = k
			c += 1
		}
	}

	for k, v := range hashMap.Changes {

		if v.Status == "del" {

			committed := hashMap.Committed[k]
			if committed == nil {
				committed = new(CommittedMapElement)
				hashMap.Committed[k] = committed
			}

			v.Status = "view"
			committed.Status = "view"
			committed.Element = nil

			if hashMap.Tx.Exists(hashMap.name + ":exists:" + k) {

				if hashMap.Tx.IsWritable() {

					hashMap.Tx.Delete(hashMap.name + ":map:" + k)
					hashMap.Tx.Delete(hashMap.name + ":exists:" + k)

					if hashMap.Indexable {
						hashMap.Tx.Delete(hashMap.name + ":list:" + strconv.FormatUint(v.index, 10))
						hashMap.Tx.Delete(hashMap.name + ":listKeys:" + k)
					}

				}

				if hashMap.DeletedEvent != nil {
					if err = hashMap.DeletedEvent([]byte(k)); err != nil {
						return
					}
				}

				committed.Stored = "del"
			} else {
				committed.Stored = "view"
			}

			v.index = 0
		}

	}

	for k, v := range hashMap.Changes {

		if v.Status == "update" {

			committed := hashMap.Committed[k]
			if committed == nil {
				committed = new(CommittedMapElement)
				hashMap.Committed[k] = committed
			}

			committed.Element = v.Element

			if !hashMap.Tx.Exists(hashMap.name + ":exists:" + k) {

				if hashMap.Tx.IsWritable() {
					hashMap.Tx.Put(hashMap.name+":exists:"+k, []byte{1})

					if hashMap.Indexable {
						//safe
						hashMap.Tx.Put(hashMap.name+":list:"+strconv.FormatUint(v.index, 10), []byte(k))
						//safe
						hashMap.Tx.Put(hashMap.name+":listKeys:"+k, []byte(strconv.FormatUint(v.index, 10)))
					}

				}

				if hashMap.StoredEvent != nil {
					if err = hashMap.StoredEvent([]byte(k), committed); err != nil {
						return
					}
				}
			}

			if hashMap.Tx.IsWritable() {
				//clone required because the element could change later on
				hashMap.Tx.Put(hashMap.name+":map:"+k, v.Element.SerializeToBytes())
			}

			committed.Status = "view"
			committed.Stored = "update"
			v.index = 0
		}

	}

	for i := 0; i < c; i++ {
		delete(hashMap.Changes, removed[i])
	}

	hashMap.CountCommitted = hashMap.Count

	if hashMap.Tx.IsWritable() {
		buf := make([]byte, binary.MaxVarintLen64)
		n := binary.PutUvarint(buf, hashMap.Count)
		//safe
		hashMap.Tx.Put(hashMap.name+":count", buf[:n])
	}

	return
}

func (hashMap *HashMap) SetTx(dbTx store_db_interface.StoreDBTransactionInterface) {
	hashMap.Tx = dbTx
}

func (hashMap *HashMap) Rollback() {
	hashMap.Changes = make(map[string]*ChangesMapElement)
	hashMap.Count = hashMap.CountCommitted
}

func (hashMap *HashMap) Reset() {
	hashMap.Committed = make(map[string]*CommittedMapElement)
}

func CreateNewHashMap(tx store_db_interface.StoreDBTransactionInterface, name string, keyLength int, indexable bool) (hashMap *HashMap) {

	if len(name) <= 4 {
		panic("Invalid name")
	}

	hashMap = &HashMap{
		name:      name,
		Committed: make(map[string]*CommittedMapElement),
		Changes:   make(map[string]*ChangesMapElement),
		Tx:        tx,
		Count:     0,
		KeyLength: keyLength,
		Indexable: indexable,
	}

	//safe to Get because int doesn't change the data
	buffer := tx.Get(hashMap.name + ":count")
	if buffer != nil {
		count, p := binary.Uvarint(buffer)
		if p <= 0 {
			panic("Error reading")
		}
		hashMap.Count = count
	}
	hashMap.CountCommitted = hashMap.Count

	return
}
