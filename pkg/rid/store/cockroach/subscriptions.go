package cockroach

import (
	"context"
	"fmt"
	"time"

	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	repos "github.com/interuss/dss/pkg/rid/repos"
	dssql "github.com/interuss/dss/pkg/sql"
	"github.com/interuss/stacktrace"

	"github.com/coreos/go-semver/semver"
	"github.com/golang/geo/s2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jonboulle/clockwork"

	"go.uber.org/zap"
)

const (
	subscriptionFields       = "id, owner, url, notification_index, cells, starts_at, ends_at, writer, updated_at"
	updateSubscriptionFields = "id, url, notification_index, cells, starts_at, ends_at, writer, updated_at"
)

func NewISASubscriptionRepo(ctx context.Context, db dssql.Queryable, dbVersion semver.Version, logger *zap.Logger, clock clockwork.Clock) repos.Subscription {
	if dbVersion.Compare(v400) >= 0 {
		return &subscriptionRepo{
			Queryable: db,
			logger:    logger,
			clock:     clock,
		}
	}
	return &subscriptionRepoV3{
		Queryable: db,
		logger:    logger,
		clock:     clock,
	}
}

// subscriptions is an implementation of the SubscriptionRepo for CRDB.
type subscriptionRepo struct {
	dssql.Queryable

	clock  clockwork.Clock
	logger *zap.Logger
}

// process a query that should return one or many subscriptions.
func (c *subscriptionRepo) process(ctx context.Context, query string, args ...interface{}) ([]*ridmodels.Subscription, error) {
	rows, err := c.Query(ctx, query, args...)
	if err != nil {
		return nil, stacktrace.Propagate(err, fmt.Sprintf("Error in query: %s", query))
	}
	defer rows.Close()

	var payload []*ridmodels.Subscription
	var cids []int64

	var writer pgtype.Text
	for rows.Next() {
		s := new(ridmodels.Subscription)

		var updateTime time.Time

		err := rows.Scan(
			&s.ID,
			&s.Owner,
			&s.URL,
			&s.NotificationIndex,
			&cids,
			&s.StartTime,
			&s.EndTime,
			&writer,
			&updateTime,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error scanning Subscription row")
		}
		s.Writer = writer.String

		s.SetCells(cids)
		s.Version = dssmodels.VersionFromTime(updateTime)
		payload = append(payload, s)
	}
	if err := rows.Err(); err != nil {
		return nil, stacktrace.Propagate(err, "Error in rows query result")
	}
	return payload, nil
}

// processOne processes a query that should return exactly a single subscription.
func (c *subscriptionRepo) processOne(ctx context.Context, query string, args ...interface{}) (*ridmodels.Subscription, error) {
	subs, err := c.process(ctx, query, args...)
	if err != nil {
		return nil, err // No need to Propagate this error as this stack layer does not add useful information
	}
	if len(subs) > 1 {
		return nil, stacktrace.NewError("Query returned %d subscriptions when only 0 or 1 was expected", len(subs))
	}
	if len(subs) == 0 {
		return nil, nil
	}
	return subs[0], nil
}

// MaxSubscriptionCountInCellsByOwner counts how many subscriptions the
// owner has in each one of these cells, and returns the number of subscriptions
// in the cell with the highest number of subscriptions.
func (c *subscriptionRepo) MaxSubscriptionCountInCellsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) (int, error) {
	// TODO:steeling this query is expensive. The standard defines the max sub
	// per "area", but area is loosely defined. Since we may not have to be so
	// strict we could keep this count in memory, (or in some other storage).
	var query = `
    SELECT
      IFNULL(MAX(subscriptions_per_cell_id), 0)
    FROM (
      SELECT
        COUNT(*) AS subscriptions_per_cell_id
      FROM (
      	SELECT unnest(cells) as cell_id
      	FROM subscriptions
      	WHERE owner = $1
      		AND ends_at >= $2
      )
      WHERE
        cell_id = ANY($3)
      GROUP BY cell_id
    )`

	row := c.QueryRow(ctx, query, owner, c.clock.Now(), dssql.CellUnionToCellIds(cells))
	var ret int
	err := row.Scan(&ret)
	return ret, stacktrace.Propagate(err, "Error scanning subscription count row")
}

// GetSubscription returns the subscription identified by "id".
// Returns nil, nil if not found
func (c *subscriptionRepo) GetSubscription(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error) {
	// TODO(steeling) we should enforce startTime and endTime to not be null at the DB level.
	var query = fmt.Sprintf(`
		SELECT %s FROM subscriptions
		WHERE id = $1`, subscriptionFields)
	uid, err := id.PgUUID()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	return c.processOne(ctx, query, uid)
}

// UpdateSubscription updates the Subscription.. not yet implemented.
// Returns nil, nil if ID, version not found
func (c *subscriptionRepo) UpdateSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	var (
		updateQuery = fmt.Sprintf(`
		UPDATE
		  subscriptions
		SET (%s) = ($1, $2, $3, $4, $5, $6, $7, transaction_timestamp())
		WHERE id = $1 AND updated_at = $8
		RETURNING
			%s`, updateSubscriptionFields, subscriptionFields)
	)

	cids, err := dssql.CellUnionToCellIdsWithValidation(s.Cells)

	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert array to jackc/pgtype")
	}

	id, err := s.ID.PgUUID()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	return c.processOne(ctx, updateQuery,
		id,
		s.URL,
		s.NotificationIndex,
		cids,
		s.StartTime,
		s.EndTime,
		s.Writer,
		s.Version.ToTimestamp())
}

// InsertSubscription inserts subscription into the store and returns
// the resulting subscription including its ID.
func (c *subscriptionRepo) InsertSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	var (
		insertQuery = fmt.Sprintf(`
		INSERT INTO
		  subscriptions
		  (%s)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, transaction_timestamp())
		RETURNING
			%s`, subscriptionFields, subscriptionFields)
	)

	cids, err := dssql.CellUnionToCellIdsWithValidation(s.Cells)

	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert array to jackc/pgtype")
	}

	id, err := s.ID.PgUUID()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	return c.processOne(ctx, insertQuery,
		id,
		s.Owner,
		s.URL,
		s.NotificationIndex,
		cids,
		s.StartTime,
		s.EndTime,
		s.Writer)
}

// DeleteSubscription deletes the subscription identified by ID.
// It must be done in a txn and the version verified.
// Returns nil, nil if ID, version not found
func (c *subscriptionRepo) DeleteSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	var (
		query = fmt.Sprintf(`
		DELETE FROM
			subscriptions
		WHERE
			id = $1
			AND updated_at = $2
		RETURNING %s`, subscriptionFields)
	)
	id, err := s.ID.PgUUID()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert id to PgUUID")
	}
	return c.processOne(ctx, query, id, s.Version.ToTimestamp())
}

// UpdateNotificationIdxsInCells incremement the notification for each sub in the given cells.
func (c *subscriptionRepo) UpdateNotificationIdxsInCells(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	var updateQuery = fmt.Sprintf(`
			UPDATE subscriptions
			SET notification_index = notification_index + 1
			WHERE
				cells && $1
				AND ends_at >= $2
			RETURNING %s`, subscriptionFields)

	return c.process(
		ctx, updateQuery, dssql.CellUnionToCellIds(cells), c.clock.Now())
}

// SearchSubscriptions returns all subscriptions in "cells".
func (c *subscriptionRepo) SearchSubscriptions(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	var (
		query = fmt.Sprintf(`
			SELECT
				%s
			FROM
				subscriptions
			WHERE
				cells && $1
			AND
				ends_at >= $2
			LIMIT $3`, subscriptionFields)
	)

	if len(cells) == 0 {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "no location provided")
	}

	return c.process(ctx, query, dssql.CellUnionToCellIds(cells), c.clock.Now(), dssmodels.MaxResultLimit)
}

// SearchSubscriptionsByOwner returns all subscriptions in "cells".
func (c *subscriptionRepo) SearchSubscriptionsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error) {
	var (
		query = fmt.Sprintf(`
			SELECT
				%s
			FROM
				subscriptions
			WHERE
				cells && $1
			AND
				subscriptions.owner = $2
			AND
				ends_at >= $3
			LIMIT $4`, subscriptionFields)
	)

	if len(cells) == 0 {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "no location provided")
	}

	return c.process(ctx, query, dssql.CellUnionToCellIds(cells), owner, c.clock.Now(), dssmodels.MaxResultLimit)
}

// ListExpiredSubscriptions lists all expired Subscriptions based on writer.
// Records expire if current time is <expiredDurationInMin> minutes more than records' endTime.
// The function queries both empty writer and null writer when passing empty string as a writer.
func (c *subscriptionRepo) ListExpiredSubscriptions(ctx context.Context, writer string) ([]*ridmodels.Subscription, error) {
	writerQuery := "'" + writer + "'"
	if len(writer) == 0 {
		writerQuery = "'' OR writer = NULL"
	}

	var (
		query = fmt.Sprintf(`
	SELECT
		%s
	FROM
		subscriptions
	WHERE
		ends_at + INTERVAL '%d' MINUTE <= CURRENT_TIMESTAMP
	AND
		(writer = %s)`, subscriptionFields, expiredDurationInMin, writerQuery)
	)

	return c.process(ctx, query)
}
