package db

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"strconv"
	"strings"
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
	//::: GENERAL STMT

	stmt := `select b.id, b.feature_id, b.content, b.is_active, b.created_at, b.updated_at,(SELECT ARRAY_AGG(bt.tag_id) from banners_tags bt  where bt.banner_id = b.id) AS tag_ids from banners b join public.banners_tags bt on b.id = bt.banner_id  `

	//::: CHECK 'where' params
	var args []interface{}
	if feature_id > 0 {
		stmt += ` WHERE bt.feature_id = $1`
		args = append(args, feature_id)
		if tag_id > 0 {

			stmt += ` AND b.tag_id = $2`
			args = append(args, tag_id)

		}
	} else {
		if tag_id > 0 {

			stmt += ` WHERE b.tag_id = $1`
			args = append(args, tag_id)

		}
	}

	stmt += `GROUP BY b.id, b.feature_id, b.content, b.is_active, b.created_at, b.updated_at ORDER BY b.id`

	//	::: CHEcK offset\limit
	if limit > 0 {
		stmt += fmt.Sprintf("LIMIT %d", limit)
	}
	if offset > 0 {
		stmt += fmt.Sprintf("OFFSET %d", offset)

	}

	//	:::: MAKE query!

	rows, err := p.db.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	var banners []*Banner
	for rows.Next() {
		var b *Banner

		var tagIds []int
		err = rows.Scan(&b.Id, b.FeatureID, b.Content, b.IsActive, b.CreatedAt, b.UpdatedAt, tagIds)
		if err != nil {
			return nil, err
		}

		b.TagIds = tagIds

		banners = append(banners, b)

	}
	return banners, nil
}
func (p *Postgres) GetUserBanner(ctx context.Context, feature_id, tag_id int) (string, error) {
	stmt := `SELECT content from banners b join banners_tags bt on bt.banner_id =b.id where b.feature_id = $1 and bt.tag_id = $2 LIMIT 1 `

	var content string

	err := p.db.QueryRow(ctx, stmt, feature_id, tag_id).Scan(&content)
	if err != nil {
		return "", err
	}
	return content, nil
}
func (p *Postgres) PostBanner(ctx context.Context, banner *Banner) error {

	stmt := `INSERT INTO banners(feature_id,content,is_active,created_at,updated_at) values ($1,$2,$3,$4,$5) returning id `

	content, err := json.Marshal(banner.Content)
	if err != nil {
		return err
	}
	var bannerId int
	err = p.db.QueryRow(ctx, stmt, banner.FeatureID, content, banner.IsActive, time.Now(), time.Now()).Scan(&bannerId)
	if err != nil {
		return err
	}
	stmt = `INSERT INTO banners_tags (tag_id,banner_id) values $1`
	//	::: INSERT tags_id
	for _, id := range banner.TagIds {
		_, err = p.db.Exec(ctx, stmt, id, bannerId)
		if err != nil {
			return err
		}
	}
	return nil

}
func (p *Postgres) UpdateBanner(ctx context.Context, banner *Banner, id int, isActive *bool) error {

	//::: MAIN state
	stmt := `update banners set `

	//	:::: PREPARE update params

	var updateParamsStmt []string
	var updateParamsCount = 1
	var args = make([]interface{}, 0)
	if len(banner.Content) > 0 {

		updateParamsStmt = append(updateParamsStmt, "content="+strconv.Itoa(updateParamsCount))
		content, err := json.Marshal(banner.Content)
		if err != nil {
			return err

		}
		args = append(args, content)
		updateParamsCount++

	}
	if banner.FeatureID != 0 {

		updateParamsStmt = append(updateParamsStmt, "content="+strconv.Itoa(updateParamsCount))
		args = append(args, banner.FeatureID)

		updateParamsCount++
	}
	if isActive != nil {

		updateParamsStmt = append(updateParamsStmt, "isActive="+strconv.Itoa(updateParamsCount))
		args = append(args, isActive)

		updateParamsCount++
	}
	updateParamsStmt = append(updateParamsStmt, "content="+strconv.Itoa(updateParamsCount))
	args = append(args, time.Now())

	updateParamsCount++

	//	::: UPdate
	stmt += strings.Join(updateParamsStmt, ", ")

	_, err := p.db.Exec(ctx, stmt, args...)
	if err != nil {
		return err
	}
	stmt = `delete from banners_tags  where banner_id = $1`
	_, err = p.db.Exec(ctx, stmt, id)
	if err != nil {
		return err
	}
	if len(banner.TagIds) > 0 {

		stmt = `insert into banners_tags (banner_id,tag_id)  values `
		var updateParamsStmt []string
		var updateParamsArgs = make([]interface{}, 0)

		//	:::delete tags on banner-id
		//	:::insert new tags on banner-id

		for _, tagId := range banner.TagIds {

			updateParamsStmt = append(updateParamsStmt, "($"+strconv.Itoa(len(updateParamsArgs)+1)+", $"+strconv.Itoa(len(updateParamsArgs)+2)+")")
			updateParamsArgs = append(updateParamsArgs, id, tagId)
		}
		stmt += strings.Join(updateParamsStmt, ", ")
		_, err = p.db.Exec(ctx, stmt, updateParamsArgs...)
		if err != nil {
			return err
		}

	}
	return nil

}
func (p *Postgres) DeleteBanner(ctx context.Context, id int) error {

	//	:::: DELETE banner
	stmt := `DELETE FROM banners_tags where banner_id = $1`
	_, err := p.db.Exec(ctx, stmt, id)
	if err != nil {
		return err

	}

	//	:::: DELETE tags
	stmt = `DELETE FROM banners where banner_id = $1`

	_, err = p.db.Exec(ctx, stmt, id)
	if err != nil {
		return err

	}
	return nil

}
