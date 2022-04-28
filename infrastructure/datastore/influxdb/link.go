package influxdb

import (
	"context"
	"encoding/json"
	"fmt"

	influxdb "github.com/influxdata/influxdb/client/v2"

	"zntr.io/hexagonal-bazel/domain/urlshortener/link"
)

// LinkAnalytics returns an influxdb implementation of the analytic repository.
func LinkAnalytics(client influxdb.Client, database string) link.Analytic {
	return &linkAnalyticRepository{
		client:   client,
		database: database,
	}
}

// -----------------------------------------------------------------------------

type linkAnalyticRepository struct {
	client   influxdb.Client
	database string
}

// -----------------------------------------------------------------------------

// Resolved create a link resolution point in the analytic database.
func (ra *linkAnalyticRepository) Resolved(ctx context.Context, id link.ID) error {
	batchPoints, err := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
		Database:  ra.database,
		Precision: "s",
	})
	if err != nil {
		return fmt.Errorf("influxdb: unable to create BatchPoints instance: %w", err)
	}

	// Prepare tag
	tag := map[string]string{
		"id": string(id),
	}

	// Prepare fields
	field := map[string]interface{}{
		"resolved": true,
	}

	// Create point
	point, err := influxdb.NewPoint("resolved", tag, field)
	if err != nil {
		return fmt.Errorf("influxdb: unable to create a point: %w", err)
	}

	// Append point to batch
	batchPoints.AddPoint(point)

	// Send to the server
	if err := ra.client.WriteCtx(ctx, batchPoints); err != nil {
		return fmt.Errorf("influxdb: unable to push batchpoints to the server: %w", err)
	}

	// No error
	return nil
}

// GetResolutionCountPerID returns resolution hits for a given period in days.
func (ra *linkAnalyticRepository) GetResolutionCountPerID(ctx context.Context, id link.ID, dayPeriod int) (int64, error) {
	// Prepare the query
	q := "SELECT COUNT(resolved) FROM resolved WHERE id = $id"
	params := map[string]interface{}{
		"id": id,
	}

	// Has day period defined
	if dayPeriod != 0 {
		q += "AND time > now() - $dayPeriod"
		params["dayPeriod"] = fmt.Sprintf("%dd", dayPeriod)
	}

	// Query the database
	response, err := ra.client.QueryCtx(ctx, influxdb.Query{
		Command:    q,
		Parameters: params,
		Database:   ra.database,
	})
	if err != nil {
		return 0, fmt.Errorf("influxdb: unable to query the database: %w", err)
	}

	// No series returned?
	if len(response.Results[0].Series) == 0 {
		return 0, link.ErrLinkNotFound
	}

	// Extract the value
	resolved, err := response.Results[0].Series[0].Values[0][1].(json.Number).Int64()
	if err != nil {
		return 0, fmt.Errorf("influxdb: can't retrieve metric value: %w", err)
	}

	return resolved, nil
}
