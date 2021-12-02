package api_common

import (
	"encoding/json"
	"errors"
	"github.com/go-pg/urlstruct"
	"net/http"
	"net/url"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type APIAccountTxsRequest struct {
	api_types.APIAccountBaseRequest
	Next uint64 `json:"next,omitempty"`
}

type APIAccountTxsAnswer struct {
	Count uint64             `json:"count,omitempty"`
	Txs   []helpers.HexBytes `json:"txs,omitempty"`
}

func (api *APICommon) AccountTxs(r *http.Request, args *APIAccountTxsRequest, reply *APIAccountTxsAnswer) (err error) {

	publicKey, err := args.GetPublicKey()
	if err != nil {
		return
	}

	publicKeyStr := string(publicKey)

	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		data := reader.Get("addrTxsCount:" + publicKeyStr)
		if data == nil {
			return nil
		}

		if reply.Count, err = strconv.ParseUint(string(data), 10, 64); err != nil {
			return
		}

		if args.Next > reply.Count {
			args.Next = reply.Count
		}

		index := args.Next
		if index < config.API_ACCOUNT_MAX_TXS {
			index = 0
		} else {
			index -= config.API_ACCOUNT_MAX_TXS
		}

		reply.Txs = make([]helpers.HexBytes, args.Next-index)
		for i := index; i < args.Next; i++ {
			hash := reader.Get("addrTx:" + publicKeyStr + ":" + strconv.FormatUint(i, 10))
			if hash == nil {
				return errors.New("Error reading address transaction")
			}
			reply.Txs[args.Next-i-1] = hash
		}

		return
	})
}

func (api *APICommon) GetAccountTxs_http(values url.Values) (interface{}, error) {
	args := &APIAccountTxsRequest{}
	if err := urlstruct.Unmarshal(nil, values, args); err != nil {
		return nil, err
	}
	reply := &APIAccountTxsAnswer{}
	return reply, api.AccountTxs(nil, args, reply)
}

func (api *APICommon) GetAccountTxs_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIAccountTxsRequest{}
	if err := json.Unmarshal(values, &args); err != nil {
		return nil, err
	}
	reply := &APIAccountTxsAnswer{}
	return reply, api.AccountTxs(nil, args, reply)
}