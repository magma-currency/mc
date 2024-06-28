package chain_network

import (
	"mc/blockchain"
	"mc/config"
	"mc/helpers/recovery"
	"mc/mempool"
	"mc/network/api_implementation/api_websockets/consensus"
	"mc/network/server/node_http"
	"mc/network/websocks"
	"mc/network/websocks/connection"
	"time"
)

func continuouslyDownloadChain() {
	recovery.SafeGo(func() {

		for {

			list := websocks.Websockets.GetAllSockets()
			for _, conn := range list {
				if conn.Handshake.Consensus == config.NODE_CONSENSUS_TYPE_FULL {
					data, err := connection.SendJSONAwaitAnswer[consensus.ChainUpdateNotification](conn, []byte("get-chain"), nil, nil, 0)
					if err == nil {
						node_http.HttpServer.ApiWebsockets.Consensus.ChainUpdateProcess(conn, data)
					}
					time.Sleep(1 * time.Millisecond)
				}
			}

			time.Sleep(2000 * time.Millisecond)
		}

	})
}

func continuouslyDownloadMempool() {

	recovery.SafeGo(func() {

		for {

			list := websocks.Websockets.GetAllSockets()
			for _, conn := range list {
				if config.NODE_CONSENSUS == config.NODE_CONSENSUS_TYPE_FULL && conn.Handshake.Consensus == config.NODE_CONSENSUS_TYPE_FULL {
					DownloadMempool(conn)
					time.Sleep(1 * time.Millisecond)
				}
			}

			time.Sleep(2000 * time.Millisecond)
		}

	})

}

func syncBlockchainNewConnections() {
	recovery.SafeGo(func() {

		cn := websocks.Websockets.UpdateNewConnectionMulticast.AddListener()
		defer websocks.Websockets.UpdateNewConnectionMulticast.RemoveChannel(cn)

		for {

			conn, ok := <-cn
			if !ok {
				return
			}

			//making it async
			recovery.SafeGo(func() {

				data, err := connection.SendJSONAwaitAnswer[consensus.ChainUpdateNotification](conn, []byte("get-chain"), nil, nil, 0)
				if err == nil {
					node_http.HttpServer.ApiWebsockets.Consensus.ChainUpdateProcess(conn, data)
				}

			})

		}
	})
}

func InitChainNetwork(chain *blockchain.Blockchain, mempool *mempool.Mempool) {

	continuouslyDownloadChain()

	if config.NODE_CONSENSUS == config.NODE_CONSENSUS_TYPE_FULL {
		continuouslyDownloadMempool()
	}

	syncBlockchainNewConnections()

	initializeConsensus(chain, mempool)

}
