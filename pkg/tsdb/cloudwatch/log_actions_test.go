package cloudwatch

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/tsdb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuery_DescribeLogGroups(t *testing.T) {
	origNewCWLogsClient := newCWLogsClient
	t.Cleanup(func() {
		newCWLogsClient = origNewCWLogsClient
	})

	var logs mockedLogs

	newCWLogsClient = func(sess *session.Session) cloudwatchlogsiface.CloudWatchLogsAPI {
		return logs
	}

	t.Run("Empty log group name prefix", func(t *testing.T) {
		logs = mockedLogs{
			logGroups: cloudwatchlogs.DescribeLogGroupsOutput{
				LogGroups: []*cloudwatchlogs.LogGroup{
					{
						LogGroupName: aws.String("group_a"),
					},
					{
						LogGroupName: aws.String("group_b"),
					},
					{
						LogGroupName: aws.String("group_c"),
					},
				},
			},
		}

		executor := &CloudWatchExecutor{}
		resp, err := executor.Query(context.Background(), mockDatasource(), &tsdb.TsdbQuery{
			Queries: []*tsdb.Query{
				{
					Model: simplejson.NewFromAny(map[string]interface{}{
						"type":    "logAction",
						"subtype": "DescribeLogGroups",
						"limit":   50,
					}),
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.Equal(t, &tsdb.Response{
			Results: map[string]*tsdb.QueryResult{
				"": {
					Dataframes: tsdb.NewDecodedDataFrames(data.Frames{
						data.NewFrame("logGroups", data.NewField("logGroupName", nil, []*string{
							aws.String("group_a"), aws.String("group_b"), aws.String("group_c"),
						})),
					}),
				},
			},
		}, resp)
	})

	t.Run("Non-empty log group name prefix", func(t *testing.T) {
		logs = mockedLogs{
			logGroups: cloudwatchlogs.DescribeLogGroupsOutput{
				LogGroups: []*cloudwatchlogs.LogGroup{
					{
						LogGroupName: aws.String("group_a"),
					},
					{
						LogGroupName: aws.String("group_b"),
					},
					{
						LogGroupName: aws.String("group_c"),
					},
				},
			},
		}

		executor := &CloudWatchExecutor{}
		resp, err := executor.Query(context.Background(), mockDatasource(), &tsdb.TsdbQuery{
			Queries: []*tsdb.Query{
				{
					Model: simplejson.NewFromAny(map[string]interface{}{
						"type":               "logAction",
						"subtype":            "DescribeLogGroups",
						"logGroupNamePrefix": "g",
					}),
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.Equal(t, &tsdb.Response{
			Results: map[string]*tsdb.QueryResult{
				"": {
					Dataframes: tsdb.NewDecodedDataFrames(data.Frames{
						data.NewFrame("logGroups", data.NewField("logGroupName", nil, []*string{
							aws.String("group_a"), aws.String("group_b"), aws.String("group_c"),
						})),
					}),
				},
			},
		}, resp)
	})
}

func TestQuery_GetLogGroupFields(t *testing.T) {
	origNewCWLogsClient := newCWLogsClient
	t.Cleanup(func() {
		newCWLogsClient = origNewCWLogsClient
	})

	var logs mockedLogs

	newCWLogsClient = func(sess *session.Session) cloudwatchlogsiface.CloudWatchLogsAPI {
		return logs
	}

	logs = mockedLogs{
		logGroupFields: cloudwatchlogs.GetLogGroupFieldsOutput{
			LogGroupFields: []*cloudwatchlogs.LogGroupField{
				{
					Name:    aws.String("field_a"),
					Percent: aws.Int64(100),
				},
				{
					Name:    aws.String("field_b"),
					Percent: aws.Int64(30),
				},
				{
					Name:    aws.String("field_c"),
					Percent: aws.Int64(55),
				},
			},
		},
	}

	const refID = "A"

	executor := &CloudWatchExecutor{}
	resp, err := executor.Query(context.Background(), mockDatasource(), &tsdb.TsdbQuery{
		Queries: []*tsdb.Query{
			{
				RefId: refID,
				Model: simplejson.NewFromAny(map[string]interface{}{
					"type":         "logAction",
					"subtype":      "GetLogGroupFields",
					"logGroupName": "group_a",
					"limit":        50,
				}),
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	expFrame := data.NewFrame(
		refID,
		data.NewField("name", nil, []*string{
			aws.String("field_a"), aws.String("field_b"), aws.String("field_c"),
		}),
		data.NewField("percent", nil, []*int64{
			aws.Int64(100), aws.Int64(30), aws.Int64(55),
		}),
	)
	expFrame.RefID = refID
	assert.Equal(t, &tsdb.Response{
		Results: map[string]*tsdb.QueryResult{
			refID: {
				Dataframes: tsdb.NewDecodedDataFrames(data.Frames{expFrame}),
				RefId:      refID,
			},
		},
	}, resp)
}

func TestQuery_StartQuery(t *testing.T) {
	origNewCWLogsClient := newCWLogsClient
	t.Cleanup(func() {
		newCWLogsClient = origNewCWLogsClient
	})

	var logs mockedLogs

	newCWLogsClient = func(sess *session.Session) cloudwatchlogsiface.CloudWatchLogsAPI {
		return logs
	}

	t.Run("invalid time range", func(t *testing.T) {
		logs = mockedLogs{
			logGroupFields: cloudwatchlogs.GetLogGroupFieldsOutput{
				LogGroupFields: []*cloudwatchlogs.LogGroupField{
					{
						Name:    aws.String("field_a"),
						Percent: aws.Int64(100),
					},
					{
						Name:    aws.String("field_b"),
						Percent: aws.Int64(30),
					},
					{
						Name:    aws.String("field_c"),
						Percent: aws.Int64(55),
					},
				},
			},
		}

		timeRange := &tsdb.TimeRange{
			From: "1584873443000",
			To:   "1584700643000",
		}

		executor := &CloudWatchExecutor{}
		_, err := executor.Query(context.Background(), mockDatasource(), &tsdb.TsdbQuery{
			TimeRange: timeRange,
			Queries: []*tsdb.Query{
				{
					Model: simplejson.NewFromAny(map[string]interface{}{
						"type":        "logAction",
						"subtype":     "StartQuery",
						"limit":       50,
						"region":      "default",
						"queryString": "fields @message",
					}),
				},
			},
		})
		require.Error(t, err)

		assert.Equal(t, fmt.Errorf("invalid time range: start time must be before end time"), err)
	})

	t.Run("valid time range", func(t *testing.T) {
		const refID = "A"
		logs = mockedLogs{
			logGroupFields: cloudwatchlogs.GetLogGroupFieldsOutput{
				LogGroupFields: []*cloudwatchlogs.LogGroupField{
					{
						Name:    aws.String("field_a"),
						Percent: aws.Int64(100),
					},
					{
						Name:    aws.String("field_b"),
						Percent: aws.Int64(30),
					},
					{
						Name:    aws.String("field_c"),
						Percent: aws.Int64(55),
					},
				},
			},
		}

		timeRange := &tsdb.TimeRange{
			From: "1584700643000",
			To:   "1584873443000",
		}

		executor := &CloudWatchExecutor{}
		resp, err := executor.Query(context.Background(), mockDatasource(), &tsdb.TsdbQuery{
			TimeRange: timeRange,
			Queries: []*tsdb.Query{
				{
					RefId: refID,
					Model: simplejson.NewFromAny(map[string]interface{}{
						"type":        "logAction",
						"subtype":     "StartQuery",
						"limit":       50,
						"region":      "default",
						"queryString": "fields @message",
					}),
				},
			},
		})
		require.NoError(t, err)

		expFrame := data.NewFrame(
			refID,
			data.NewField("queryId", nil, []string{"abcd-efgh-ijkl-mnop"}),
		)
		expFrame.RefID = refID
		expFrame.Meta = &data.FrameMeta{
			Custom: map[string]interface{}{
				"Region": "default",
			},
		}
		assert.Equal(t, &tsdb.Response{
			Results: map[string]*tsdb.QueryResult{
				refID: {
					Dataframes: tsdb.NewDecodedDataFrames(data.Frames{expFrame}),
					RefId:      refID,
				},
			},
		}, resp)
	})
}

func TestQuery_StopQuery(t *testing.T) {
	origNewCWLogsClient := newCWLogsClient
	t.Cleanup(func() {
		newCWLogsClient = origNewCWLogsClient
	})

	var logs mockedLogs

	newCWLogsClient = func(sess *session.Session) cloudwatchlogsiface.CloudWatchLogsAPI {
		return logs
	}

	logs = mockedLogs{
		logGroupFields: cloudwatchlogs.GetLogGroupFieldsOutput{
			LogGroupFields: []*cloudwatchlogs.LogGroupField{
				{
					Name:    aws.String("field_a"),
					Percent: aws.Int64(100),
				},
				{
					Name:    aws.String("field_b"),
					Percent: aws.Int64(30),
				},
				{
					Name:    aws.String("field_c"),
					Percent: aws.Int64(55),
				},
			},
		},
	}

	timeRange := &tsdb.TimeRange{
		From: "1584873443000",
		To:   "1584700643000",
	}

	executor := &CloudWatchExecutor{}
	resp, err := executor.Query(context.Background(), mockDatasource(), &tsdb.TsdbQuery{
		TimeRange: timeRange,
		Queries: []*tsdb.Query{
			{
				Model: simplejson.NewFromAny(map[string]interface{}{
					"type":    "logAction",
					"subtype": "StopQuery",
					"queryId": "abcd-efgh-ijkl-mnop",
				}),
			},
		},
	})
	require.NoError(t, err)

	expFrame := data.NewFrame(
		"StopQueryResponse",
		data.NewField("success", nil, []bool{true}),
	)
	assert.Equal(t, &tsdb.Response{
		Results: map[string]*tsdb.QueryResult{
			"": {
				Dataframes: tsdb.NewDecodedDataFrames(data.Frames{expFrame}),
			},
		},
	}, resp)
}

func TestQuery_GetQueryResults(t *testing.T) {
	origNewCWLogsClient := newCWLogsClient
	t.Cleanup(func() {
		newCWLogsClient = origNewCWLogsClient
	})

	var logs mockedLogs

	newCWLogsClient = func(sess *session.Session) cloudwatchlogsiface.CloudWatchLogsAPI {
		return logs
	}

	const refID = "A"
	logs = mockedLogs{
		queryResults: cloudwatchlogs.GetQueryResultsOutput{
			Results: [][]*cloudwatchlogs.ResultField{
				{
					{
						Field: aws.String("@timestamp"),
						Value: aws.String("2020-03-20 10:37:23.000"),
					},
					{
						Field: aws.String("field_b"),
						Value: aws.String("b_1"),
					},
					{
						Field: aws.String("@ptr"),
						Value: aws.String("abcdefg"),
					},
				},
				{
					{
						Field: aws.String("@timestamp"),
						Value: aws.String("2020-03-20 10:40:43.000"),
					},
					{
						Field: aws.String("field_b"),
						Value: aws.String("b_2"),
					},
					{
						Field: aws.String("@ptr"),
						Value: aws.String("hijklmnop"),
					},
				},
			},
			Statistics: &cloudwatchlogs.QueryStatistics{
				BytesScanned:   aws.Float64(512),
				RecordsMatched: aws.Float64(256),
				RecordsScanned: aws.Float64(1024),
			},
			Status: aws.String("Complete"),
		},
	}

	executor := &CloudWatchExecutor{}
	resp, err := executor.Query(context.Background(), mockDatasource(), &tsdb.TsdbQuery{
		Queries: []*tsdb.Query{
			{
				RefId: refID,
				Model: simplejson.NewFromAny(map[string]interface{}{
					"type":    "logAction",
					"subtype": "GetQueryResults",
					"queryId": "abcd-efgh-ijkl-mnop",
				}),
			},
		},
	})
	require.NoError(t, err)

	time1, err := time.Parse("2006-01-02 15:04:05.000", "2020-03-20 10:37:23.000")
	require.NoError(t, err)
	time2, err := time.Parse("2006-01-02 15:04:05.000", "2020-03-20 10:40:43.000")
	require.NoError(t, err)
	expField1 := data.NewField("@timestamp", nil, []*time.Time{
		aws.Time(time1), aws.Time(time2),
	})
	expField1.SetConfig(&data.FieldConfig{DisplayName: "Time"})
	expField2 := data.NewField("field_b", nil, []*string{
		aws.String("b_1"), aws.String("b_2"),
	})
	expFrame := data.NewFrame(refID, expField1, expField2)
	expFrame.RefID = refID
	expFrame.Meta = &data.FrameMeta{
		Custom: map[string]interface{}{
			"Status": "Complete",
			"Statistics": cloudwatchlogs.QueryStatistics{
				BytesScanned:   aws.Float64(512),
				RecordsMatched: aws.Float64(256),
				RecordsScanned: aws.Float64(1024),
			},
		},
	}

	assert.Equal(t, &tsdb.Response{
		Results: map[string]*tsdb.QueryResult{
			refID: {
				RefId:      refID,
				Dataframes: tsdb.NewDecodedDataFrames(data.Frames{expFrame}),
			},
		},
	}, resp)
}
