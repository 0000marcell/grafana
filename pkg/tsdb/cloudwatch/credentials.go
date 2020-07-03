package cloudwatch

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/credentials/endpointcreds"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
)

// Session factory.
// Stubbable by tests.
var newSession = session.NewSession

// STS credentials factory.
// Stubbable by tests.
var newSTSCredentials = stscreds.NewCredentials

// EC2Metadata service factory.
// Stubbable by tests.
var newEC2Metadata = ec2metadata.New

func remoteCredProvider(sess *session.Session) credentials.Provider {
	ecsCredURI := os.Getenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI")

	if len(ecsCredURI) > 0 {
		return ecsCredProvider(sess, ecsCredURI)
	}
	return ec2RoleProvider(sess)
}

func ecsCredProvider(sess *session.Session, uri string) credentials.Provider {
	const host = `169.254.170.2`

	d := defaults.Get()
	return endpointcreds.NewProviderClient(
		*d.Config,
		d.Handlers,
		fmt.Sprintf("http://%s%s", host, uri),
		func(p *endpointcreds.Provider) { p.ExpiryWindow = 5 * time.Minute })
}

func ec2RoleProvider(sess *session.Session) credentials.Provider {
	return &ec2rolecreds.EC2RoleProvider{Client: newEC2Metadata(sess), ExpiryWindow: 5 * time.Minute}
}

func (e *CloudWatchExecutor) getDsInfo(region string) *DatasourceInfo {
	return retrieveDsInfo(e.DataSource, region)
}

func retrieveDsInfo(datasource *models.DataSource, region string) *DatasourceInfo {
	defaultRegion := datasource.JsonData.Get("defaultRegion").MustString()
	if region == "default" {
		region = defaultRegion
	}

	authType := datasource.JsonData.Get("authType").MustString()
	assumeRoleArn := datasource.JsonData.Get("assumeRoleArn").MustString()
	externalID := datasource.JsonData.Get("externalId").MustString()
	decrypted := datasource.DecryptedValues()
	accessKey := decrypted["accessKey"]
	secretKey := decrypted["secretKey"]

	datasourceInfo := &DatasourceInfo{
		Region:        region,
		Profile:       datasource.Database,
		AuthType:      authType,
		AssumeRoleArn: assumeRoleArn,
		ExternalID:    externalID,
		AccessKey:     accessKey,
		SecretKey:     secretKey,
	}

	return datasourceInfo
}

func newAWSSession(dsInfo *DatasourceInfo) (*session.Session, error) {
	regionCfg := &aws.Config{Region: aws.String(dsInfo.Region)}
	cfgs := []*aws.Config{
		regionCfg,
	}
	// Choose authentication scheme based on the type chosen for the data source
	// Basically, we support the following methods:
	// Shared credentials: Providing access key pair sourced from user's AWS credentials file
	// Static credentials: Providing access key pair directly
	// SDK: Leave it to SDK to decide
	switch dsInfo.AuthType {
	case "credentials":
		plog.Debug("Authenticating towards AWS with shared credentials", "profile", dsInfo.Profile,
			"region", dsInfo.Region)
		cfgs = append(cfgs, &aws.Config{
			Credentials: credentials.NewSharedCredentials("", dsInfo.Profile),
		})
	case "keys":
		plog.Debug("Authenticating towards AWS with an access key pair", "region", dsInfo.Region)
		cfgs = append(cfgs, &aws.Config{
			Credentials: credentials.NewStaticCredentials(dsInfo.AccessKey, dsInfo.SecretKey, ""),
		})
	case "sdk":
		plog.Debug("Authenticating towards AWS with default SDK method", "region", dsInfo.Region)
	default:
		return nil, fmt.Errorf(`%q is not a valid authentication type - expected "credentials", "keys" or "sdk"`,
			dsInfo.AuthType)
	}
	sess, err := newSession(cfgs...)
	if err != nil {
		return nil, err
	}

	// We should assume a role in AWS
	if dsInfo.AssumeRoleArn != "" {
		plog.Debug("Trying to assume role in AWS", "arn", dsInfo.AssumeRoleArn)

		sess, err = newSession(regionCfg, &aws.Config{
			Credentials: newSTSCredentials(sess, dsInfo.AssumeRoleArn, func(p *stscreds.AssumeRoleProvider) {
				if dsInfo.ExternalID != "" {
					p.ExternalID = aws.String(dsInfo.ExternalID)
				}
			}),
		})
		if err != nil {
			return nil, err
		}
	}

	plog.Debug("Successfully authenticated towards AWS")
	return sess, nil
}

func (e *CloudWatchExecutor) getClient(region string) (*cloudwatch.CloudWatch, error) {
	datasourceInfo := e.getDsInfo(region)
	sess, err := newAWSSession(datasourceInfo)
	if err != nil {
		return nil, err
	}

	client := cloudwatch.New(sess)

	client.Handlers.Send.PushFront(func(r *request.Request) {
		r.HTTPRequest.Header.Set("User-Agent", fmt.Sprintf("Grafana/%s", setting.BuildVersion))
	})

	return client, nil
}

func retrieveLogsClient(datasourceInfo *DatasourceInfo) (*cloudwatchlogs.CloudWatchLogs, error) {
	plog.Debug("Creating CloudWatch logs client")
	sess, err := newAWSSession(datasourceInfo)
	if err != nil {
		return nil, err
	}

	client := cloudwatchlogs.New(sess)

	client.Handlers.Send.PushFront(func(r *request.Request) {
		r.HTTPRequest.Header.Set("User-Agent", fmt.Sprintf("Grafana/%s", setting.BuildVersion))
	})

	return client, nil
}
