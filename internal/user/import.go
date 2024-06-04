package user

import (
	"context"
	"strconv"

	//nolint:gosec
	"crypto/md5" // md5 here is for uniq value based on username, so this is safe
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

func (s *Storage) Insert(ctx context.Context, id int, name, userName, email string, isAdmin, isActive bool) (bool, error) {
	// Calculate ID
	userHash := hashUserID(id, userName)

	// Select
	selectUserBuilder := builderUserGet.Where(sq.Eq{"user_id": userHash})
	query, args, err := selectUserBuilder.ToSql()

	if err != nil {
		s.logger.Error().
			Err(err).
			Msgf("Failed build SQL select %s query", tableOnCallUsers)

		return false, err
	}

	sqlResult, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		s.logger.Error().
			Err(err).
			Msgf("Failed execute SQL select %s query", tableOnCallUsers)

		return false, err
	}
	defer func() { _ = sqlResult.Close() }()

	if sqlResult.Next() {
		// Compare name - email / admin role / disabled, if not matches - UPDATE
		query, args, err = s.updateQuery(userHash, email, isAdmin, isActive, sqlResult)

		if len(args) == 0 {
			if err != nil {
				s.logger.Error().
					Err(err).
					Str("username", userName).
					Msgf("Failed execute update %s query", tableOnCallUsers)
			}

			return false, err
		}

		s.logger.Info().Str("username", userName).Msg("Update user")
	} else {
		role := roleUser
		if isAdmin {
			role = roleAdmin
		}

		s.logger.Info().Str("username", userName).Msg("Insert user")

		// Not found - insert new record
		query, args, err = sq.Insert("users").
			PlaceholderFormat(sq.Dollar).
			Columns("user_id", "name", "username", "active", "email", "role").
			Values(userHash, name, userName, isActive, email, role).
			ToSql()
	}

	if err != nil {
		s.logger.Error().
			Err(err).
			Str("query", query).
			Interface("args", args).
			Msg("Failed build SQL user query")

		return false, err
	}

	sqlResult2, err := s.db.QueryContext(ctx, query, args...)

	if err != nil {
		s.logger.Error().
			Err(err).
			Str("query", query).
			Interface("args", args).
			Msg("Failed execute SQL query")

		return false, err
	}
	defer func() { _ = sqlResult2.Close() }()

	return true, nil
}

func (s *Storage) updateQuery(userID, email string, isAdmin, isActive bool, sqlResult *sql.Rows) (query string, args []interface{}, err error) {
	var (
		dbID       int
		dbUserID   string
		dbEmail    string
		dbUsername string
		dbRole     string
		dbActive   bool
	)

	updateSQL := builderUpdateUser.Where(sq.Eq{"user_id": userID})

	if err := sqlResult.Scan(&dbID, &dbUserID, &dbEmail, &dbUsername, &dbRole, &dbActive); err != nil {
		return "", nil, err
	}

	if strings.Compare(dbEmail, email) != 0 {
		updateSQL = updateSQL.Set("email", email)
	}

	if dbRole == roleAdmin && !isAdmin {
		updateSQL = updateSQL.Set("role", roleUser)
	}

	if dbRole == roleUser && isAdmin {
		updateSQL = updateSQL.Set("role", roleAdmin)
	}

	if dbActive != isActive {
		updateSQL = updateSQL.Set("active", strconv.FormatBool(isActive))
	}

	return updateSQL.ToSql()
}

func hashUserID(id int, userName string) string {
	// "U" + MD5(substr(userName + Id), 0, 14)

	hashStr := strings.ToLower(userName) + fmt.Sprintf("-%d", id)
	// nolint:gosec
	hash := md5.Sum([]byte(hashStr))
	md5Str := hex.EncodeToString(hash[:])

	userIDHash := "U" + strings.ToUpper(md5Str[:14])

	return userIDHash
}
