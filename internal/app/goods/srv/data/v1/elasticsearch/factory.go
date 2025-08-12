package elasticsearch

import (
	"github.com/olivere/elastic/v7"
	"emshop/internal/app/goods/srv/data/v1/interfaces"
	"emshop/internal/app/pkg/options"
	"emshop/pkg/db"
	"emshop/pkg/errors"
	"sync"
)

// SearchFactory 搜索工厂接口
type SearchFactory interface {
	Goods() interfaces.GoodsSearchStore
}

var (
	searchFactory SearchFactory
	once          sync.Once
)

type esSearchFactory struct {
	esClient *elastic.Client
}

func (esf *esSearchFactory) Goods() interfaces.GoodsSearchStore {
	return newGoods(esf)
}

var _ SearchFactory = &esSearchFactory{}

// NewElasticsearchFactory 创建Elasticsearch搜索工厂
func NewElasticsearchFactory(opts *options.EsOptions) (SearchFactory, error) {
	if opts == nil && searchFactory == nil {
		return nil, errors.New("failed to get es client")
	}

	var err error
	once.Do(func() {
		esOpt := db.EsOptions{
			Host: opts.Host,
			Port: opts.Port,
		}
		esClient, err := db.NewEsClient(&esOpt)
		if err != nil {
			return
		}
		searchFactory = &esSearchFactory{esClient: esClient}
	})
	if searchFactory == nil || err != nil {
		return nil, errors.New("failed to get es client")
	}
	return searchFactory, nil
}