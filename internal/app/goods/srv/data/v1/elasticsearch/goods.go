package elasticsearch

import (
	"context"
	"encoding/json"
	"emshop/internal/app/pkg/code"
	"emshop/pkg/errors"
	"strconv"

	"github.com/olivere/elastic/v7"
	"emshop/internal/app/goods/srv/data/v1/interfaces"
	"emshop/internal/app/goods/srv/domain/do"
)

type goods struct {
	esClient *elastic.Client
}

func newGoods(esf *esSearchFactory) *goods {
	return &goods{esClient: esf.esClient}
}

func (g *goods) Create(ctx context.Context, goods *do.GoodsSearchDO) error {
	_, err := g.esClient.Index().
		Index(goods.GetIndexName()).
		Id(strconv.Itoa(int(goods.ID))).
		BodyJson(&goods).
		Do(ctx)
	return err
}

func (g *goods) Delete(ctx context.Context, ID uint64) error {
	_, err := g.esClient.Delete().
		Index(do.GoodsSearchDO{}.GetIndexName()).
		Id(strconv.Itoa(int(ID))).
		Refresh("true").
		Do(ctx)
	return err
}

func (g *goods) Update(ctx context.Context, goods *do.GoodsSearchDO) error {
	err := g.Delete(ctx, uint64(goods.ID))
	if err != nil {
		// 忽略404错误，因为文档可能不存在
		if !elastic.IsNotFound(err) {
			return err
		}
	}
	err = g.Create(ctx, goods)
	if err != nil {
		return err
	}
	return nil
}

func (g *goods) Search(ctx context.Context, req *interfaces.GoodsFilterRequest) (*do.GoodsSearchDOList, error) {
	// match bool 复合查询
	q := elastic.NewBoolQuery()
	if req.KeyWords != "" {
		q = q.Must(elastic.NewMultiMatchQuery(req.KeyWords, "name", "goods_brief"))
	}
	if req.IsHot {
		q = q.Filter(elastic.NewTermQuery("is_hot", req.IsHot))
	}
	if req.IsNew {
		q = q.Filter(elastic.NewTermQuery("is_new", req.IsNew))
	}

	if req.PriceMin > 0 {
		q = q.Filter(elastic.NewRangeQuery("shop_price").Gte(req.PriceMin))
	}
	if req.PriceMax > 0 {
		q = q.Filter(elastic.NewRangeQuery("shop_price").Lte(req.PriceMax))
	}

	if req.BrandID > 0 {
		q = q.Filter(elastic.NewTermQuery("brands_id", req.BrandID))
	}

	if len(req.CategoryIDs) > 0 {
		q = q.Filter(elastic.NewTermsQuery("category_id", req.CategoryIDs...))
	}

	// 分页
	if req.Pages == 0 {
		req.Pages = 1
	}

	switch {
	case req.PagePerNums > 100:
		req.PagePerNums = 100
	case req.PagePerNums <= 0:
		req.PagePerNums = 10
	}

	res, err := g.esClient.Search().Index(do.GoodsSearchDO{}.
		GetIndexName()).Query(q).
		From(int(req.Pages-1) * int(req.PagePerNums)).
		Size(int(req.PagePerNums)).Do(ctx)

	if err != nil {
		return nil, errors.WithCode(code.ErrEsQuery, "%s", err.Error())
	}

	var ret do.GoodsSearchDOList
	if res.Hits != nil && res.Hits.TotalHits != nil {
		ret.TotalCount = res.Hits.TotalHits.Value
	}
	if res.Hits != nil && res.Hits.Hits != nil {
		for _, value := range res.Hits.Hits {
			goods := do.GoodsSearchDO{}
			err := json.Unmarshal(value.Source, &goods)
			if err != nil {
				return nil, errors.WithCode(code.ErrEsUnmarshal, "%s", err.Error())
			}
			ret.Items = append(ret.Items, &goods)
		}
	}
	return &ret, nil
}

var _ interfaces.GoodsSearchStore = &goods{}