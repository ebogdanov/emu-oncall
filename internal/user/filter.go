package user

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/ebogdanov/emu-oncall/internal/db"
	"github.com/rs/zerolog"

	sq "github.com/Masterminds/squirrel"
)

const (
	itemsOnPage      = 1000
	tableOnCallUsers = "oncall_users"
)

var (
	colsWithPhone  = []string{"id", "user_id", "email", "username", "role", "phone_number"}
	colsWithActive = []string{"id", "user_id", "email", "username", "role", "active"}

	builderUsersSelect = sq.Select(colsWithPhone...).
				From(tableOnCallUsers).
				PlaceholderFormat(sq.Dollar).
				OrderBy("id ASC").
				Limit(itemsOnPage)

	builderUsersCount = sq.Select("count(id)").
				From(tableOnCallUsers).
				PlaceholderFormat(sq.Dollar)

	builderUserGet = sq.Select(colsWithActive...).
			From(tableOnCallUsers).
			PlaceholderFormat(sq.Dollar).
			Limit(1)

	builderUpdateUser = sq.Update(tableOnCallUsers).
				PlaceholderFormat(sq.Dollar)
)

type Storage struct {
	db     *db.DBx
	logger zerolog.Logger
}

func NewStorage(dbx *db.DBx, l zerolog.Logger) *Storage {
	return &Storage{
		db:     dbx,
		logger: l.With().Str("component", "db").Logger(),
	}
}

func (s *Storage) Filter(ctx context.Context, req http.Request) (*List, error) {
	opts := s.fromHTTPRequest(req)

	return s.DBQuery(ctx, *opts)
}

func (s *Storage) DBQuery(ctx context.Context, opts Options) (*List, error) {
	query, queryCnt, args, err := s.dbQuery(opts)
	if err != nil {
		return nil, err
	}

	sqlResult, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("query", query).
			Interface("args", args).
			Msgf("failed execute sql select query from %s", tableOnCallUsers)

		return nil, err
	}

	cntResult, _ := s.db.QueryContext(ctx, queryCnt, args...)

	return s.process(cntResult, sqlResult, opts)
}

func (s *Storage) WithEmail(ctx context.Context, email string) (*Item, error) {
	opts := &Options{Email: email, Limit: 1, Short: false}

	query, _, args, err := s.dbQuery(*opts)
	if err != nil {
		return nil, err
	}

	sqlResult, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		s.logger.Error().
			Err(err).
			Msgf("failed execute sql lookup query from %s", tableOnCallUsers)

		return nil, err
	}

	defer func() { _ = sqlResult.Close() }()
	var (
		id uint64
	)

	item := &Item{}
	if sqlResult.Next() {
		err = sqlResult.Scan(&id, &item.ID, &item.Email, &item.Username, &item.Role, &item.PhoneNumber)

		if err != nil {
			s.logger.Error().
				Err(err).
				Msg("unable to scan user row for query")

			return nil, err
		}

		item.IsPhoneNumberVerified = item.PhoneNumber != ""

		item.Slack = &Slack{
			UserID: item.ID,
			TeamID: item.ID,
		}
	}

	return item, nil
}

func (s *Storage) dbQuery(opts Options) (query, cntQuery string, args []interface{}, err error) {
	cntBuilder := builderUsersCount
	selectQueryBuilder := builderUsersSelect

	if opts.Page > 1 {
		selectQueryBuilder = selectQueryBuilder.Offset(uint64((opts.Page - 1) * itemsOnPage))
	}

	if opts.Limit > 0 {
		selectQueryBuilder = selectQueryBuilder.Limit(uint64(opts.Limit))
		cntBuilder = cntBuilder.Limit(uint64(opts.Limit))
	}

	if opts.UserID != "" {
		whereUserID := sq.Eq{"user_id": opts.UserID}

		selectQueryBuilder = selectQueryBuilder.Where(whereUserID)
		cntBuilder = cntBuilder.Where(whereUserID)
	}

	if opts.Email != "" {
		whereEmail := sq.Eq{"email": opts.Email}

		selectQueryBuilder = selectQueryBuilder.Where(whereEmail)
		cntBuilder = cntBuilder.Where(whereEmail)
	}

	if len(opts.Roles) > 0 {
		whereRoleID := sq.Or{}
		for _, item := range opts.Roles {
			whereRoleID = append(whereRoleID, sq.Eq{"role": item})
		}

		selectQueryBuilder = selectQueryBuilder.Where(whereRoleID)
		cntBuilder = cntBuilder.Where(whereRoleID)
	}

	query, args, err = selectQueryBuilder.ToSql()
	if err != nil {
		s.logger.Error().
			Err(err).
			Msgf("failed build select sql from %s", tableOnCallUsers)

		return "", "", nil, err
	}

	cntQuery, _, _ = cntBuilder.ToSql()

	return query, cntQuery, args, err
}

func (s *Storage) process(cntResult, sqlResult *sql.Rows, opts Options) (*List, error) {
	var (
		id          uint64
		err         error
		cntRows     sql.NullInt64
		phoneNumber sql.NullString
	)

	defer func() {
		_ = cntResult.Close()
		_ = sqlResult.Close()
	}()

	if cntResult.Next() {
		err = cntResult.Scan(&cntRows)

		if err != nil {
			s.logger.Error().
				Err(err).
				Msg("unable to scan count result")
		}
	}

	responseList := &List{
		Count:  cntRows.Int64,
		Result: make([]Item, 0, cntRows.Int64),
	}

	if opts.Page > 1 {
		previous := uint64(opts.Page - 1)
		responseList.Previous = &previous
	}

	for sqlResult.Next() {
		var item Item

		err = sqlResult.Scan(&id, &item.ID, &item.Email, &item.Username, &item.Role, &phoneNumber)

		if err != nil {
			s.logger.Error().
				Err(err).
				Msg("unable to scan user row")
			break
		}
		item.IsPhoneNumberVerified = phoneNumber.String != ""

		// Add "Slack details" + PhoneNumber if requested
		if !opts.Short {
			item.PhoneNumber = phoneNumber.String

			item.Slack = &Slack{
				UserID: item.ID,
				TeamID: item.ID,
			}
		}

		responseList.Result = append(responseList.Result, item)
	}

	if responseList.Count > itemsOnPage*int64(opts.Page) {
		nextPage := uint64(opts.Page + 1)
		responseList.Next = &nextPage
	}

	return responseList, err
}
