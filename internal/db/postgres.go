package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"time"
)

type Postgres struct {
	db *pgx.Conn
}

type Banner struct {
	Id        int                    `json:"banner_id"`
	TagIds    []int                  `json:"tag_ids"`
	FeatureID int                    `json:"feature_id"`
	Content   map[string]interface{} `json:"content"`
	IsActive  bool                   `json:"is_active"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

func NewPostgres(ctx context.Context, connStr string) (*Postgres, error) {
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, err
	}
	return &Postgres{conn}, nil
}

func (p *Postgres) GetAdminBanner(ctx context.Context, feature_id, tag_id, limit, offset int) ([]*Banner, error) {

}
func (p *Postgres) GetUserBanner(ctx context.Context, feature_id, tag_id int) (string, error) {

}
func (p *Postgres) PostBanner(ctx context.Context, banner *Banner) error {

}
func (p *Postgres) UpdateBanner(ctx context.Context, banner *Banner, id int) error {

}
func (p *Postgres) DeleteBanner(ctx context.Context, id int) error {

}
