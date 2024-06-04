package user

// nolint:gosec
import (
	"context"
	"database/sql"
	"net/http"

	"github.com/ebogdanov/emu-oncall/internal/logger"

	sq "github.com/Masterminds/squirrel"
)

const itemsOnPage = 100

var (
	colsWithPhone  = []string{"id", "user_id", "email", "username", "role", "phone_number"}
	colsWithActive = []string{"id", "user_id", "email", "username", "role", "active"}

	builderUsersSelect = sq.Select(colsWithPhone...).
				From("users").
				PlaceholderFormat(sq.Dollar).
				OrderBy("id ASC").
				Limit(itemsOnPage)

	builderUsersCount = sq.Select("count(id)").
				From("users").
				PlaceholderFormat(sq.Dollar)

	builderUserGet = sq.Select(colsWithActive...).
			From("users").
			PlaceholderFormat(sq.Dollar).
			Limit(1)

	builderUpdateUser = sq.Update("users").
				PlaceholderFormat(sq.Dollar)
)

type Storage struct {
	db     *sql.DB
	logger *logger.Instance
}

func NewStorage(db *sql.DB, l *logger.Instance) *Storage {
	return &Storage{
		db:     db,
		logger: l,
	}
}

func (s *Storage) ByFilter(ctx context.Context, req http.Request) (*List, error) {
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
		s.logger.Error().Str("component", "user_storage").
			Err(err).
			Interface("args", args).
			Msg("failed execute sql select users query")

		return nil, err
	}

	cntResult, _ := s.db.QueryContext(ctx, queryCnt, args...)

	return s.serializeList(cntResult, sqlResult, opts)
}

func (s *Storage) ByEmail(ctx context.Context, email string) (*Item, error) {
	opts := &Options{Email: email, Limit: 1, Short: false}

	query, _, args, err := s.dbQuery(*opts)
	if err != nil {
		return nil, err
	}

	sqlResult, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		s.logger.Error().Str("component", "user_storage").
			Err(err).
			Msg("failed execute sql lookup user query")

		return nil, err
	}

	var id uint64
	defer func() { _ = sqlResult.Close() }()

	if sqlResult.Next() {
		item := &Item{}
		err = sqlResult.Scan(&id, &item.ID, &item.Email, &item.Username, &item.Role, &item.PhoneNumber)

		if err == nil {
			item.IsPhoneNumberVerified = item.PhoneNumber != ""

			item.Slack = &Slack{
				UserID: item.ID,
				TeamID: item.ID,
			}

			return item, nil
		}

		s.logger.Error().Str("component", "user_storage").
			Err(err).
			Msg("unable to scan user row")
	}

	return nil, err
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
		s.logger.Error().Str("component", "user_storage").
			Err(err).
			Msg("failed build select users sql")

		return "", "", nil, err
	}

	cntQuery, _, _ = cntBuilder.ToSql()

	return query, cntQuery, args, err
}

func (s *Storage) serializeList(cntResult, sqlResult *sql.Rows, opts Options) (*List, error) {
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
			s.logger.Error().Str("component", "user_storage").
				Err(err).
				Msg("unable to scan cnt result")
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
			s.logger.Error().Str("component", "user_storage").
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
