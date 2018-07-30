package fusioncdn

import (
	"golang.org/x/net/context"
	"qbox.us/api/fusion/fusion"
)

type TaskResult struct {
	Id string `json:"taskId"`
}

type CreateArgs struct {
	Domain      string             `json:"-"`
	Provider    fusion.CDNProvider `json:"cdnProvider"`
	CallbackURL string             `json:"callbackURL"`
	Config      QiniuConfiguration `json:"config"`
}

type ModifyArgs struct {
	Domain      string             `json:"-"`
	Provider    fusion.CDNProvider `json:"cdnProvider"`
	CallbackURL string             `json:"callbackURL"`
	Config      QiniuConfiguration `json:"config"`
}

type QueryArgs struct {
	Domain      string             `json:"-"`
	CdnProvider fusion.CDNProvider `json:"-"`
}

type DeleteArgs struct {
	Domain      string             `json:"-"`
	Provider    fusion.CDNProvider `json:"cdnProvider"`
	CallbackURL string             `json:"callbackURL"`
}

type Service interface {
	Create(ctx context.Context, args *CreateArgs) (task TaskResult, err error)
	Modify(ctx context.Context, args *ModifyArgs) (task TaskResult, err error)
	Delete(ctx context.Context, args *DeleteArgs) (task TaskResult, err error)
	Query(ctx context.Context, args *QueryArgs) (domainInfo []*DomainInfo, err error)
}
