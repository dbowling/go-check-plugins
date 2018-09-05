package checkcloudwatchlogs

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"

	"github.com/mackerelio/checkers"
)

type logOpts struct {
	Region          string `long:"region" value-name:"REGION" description:"AWS Region"`
	AccessKeyID     string `long:"access-key-id" value-name:"ACCESS-KEY-ID" description:"AWS Access Key ID"`
	SecretAccessKey string `long:"secret-access-key" value-name:"SECRET-ACCESS-KEY" description:"AWS Secret Access Key"`
	LogGroupName    string `long:"log-group-name" value-name:"LOG-GROUP-NAME" description:"Log group name"`
}

// Do the plugin
func Do() {
	ckr := run(os.Args[1:])
	ckr.Name = "CloudWatch Logs"
	ckr.Exit()
}

type cloudwatchLogsPlugin struct {
	Service      *cloudwatchlogs.CloudWatchLogs
	LogGroupName string
}

func newCloudwatchLogsPlugin(args []string) (*cloudwatchLogsPlugin, error) {
	opts := &logOpts{}
	_, err := flags.ParseArgs(opts, args)
	if err != nil {
		return nil, err
	}
	service, err := createService(opts)
	if err != nil {
		return nil, err
	}
	return &cloudwatchLogsPlugin{
		Service:      service,
		LogGroupName: opts.LogGroupName,
	}, nil
}

func createService(opts *logOpts) (*cloudwatchlogs.CloudWatchLogs, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	config := aws.NewConfig()
	if opts.AccessKeyID != "" && opts.SecretAccessKey != "" {
		config = config.WithCredentials(
			credentials.NewStaticCredentials(opts.AccessKeyID, opts.SecretAccessKey, ""),
		)
	}
	if opts.Region != "" {
		config = config.WithRegion(opts.Region)
	}
	return cloudwatchlogs.New(sess, config), nil
}

func (p *cloudwatchLogsPlugin) run() error {
	if p.LogGroupName == "" {
		return errors.New("specify log group name")
	}
	var nextToken *string
	for {
		startTime := time.Now().Add(-5 * time.Minute)
		output, err := p.Service.FilterLogEvents(&cloudwatchlogs.FilterLogEventsInput{
			StartTime:    aws.Int64(startTime.Unix() * 1000),
			LogGroupName: aws.String(p.LogGroupName),
			NextToken:    nextToken,
		})
		if err != nil {
			return err
		}
		fmt.Printf("%#v\n", err)
		fmt.Printf("%#v\n", output)
		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}

func run(args []string) *checkers.Checker {
	p, err := newCloudwatchLogsPlugin(args)
	if err != nil {
		return checkers.NewChecker(checkers.UNKNOWN, fmt.Sprint(err))
	}
	err = p.run()
	if err != nil {
		return checkers.NewChecker(checkers.UNKNOWN, fmt.Sprint(err))
	}
	return checkers.NewChecker(checkers.OK, "ok")
}
