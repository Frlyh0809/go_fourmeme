package client

import (
	"github.com/ethereum/go-ethereum/ethclient"
)

func NewClient(url string) (*ethclient.Client, error) {
	return ethclient.Dial(url)
}
