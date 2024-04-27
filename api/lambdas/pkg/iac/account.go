package iac

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/flostadler/festus/api/pkg/types"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var pulumiCommand auto.PulumiCommand = nil

const pulumiURL = "https://github.com/pulumi/pulumi/releases/download/v3.113.3/pulumi-v3.113.3-linux-x64.tar.gz"

func CreateAccount(ctx context.Context, account *types.Account, org *types.Organization) (string, error) {
	// Create a new Pulumi stack
	if pulumiCommand == nil {
		if err := installPulumiCLI(ctx); err != nil {
			println("Failed to install pulumi CLI: %s", err.Error())
			return "", err
		}
		println("Successfully installed pulumi CLI")
	} else {
		println("Reusing pre-initialized pulumi CLI installation")
	}

	// TODO: his is just a stand-in for the stack that should create an account. But this is good enough to prove the automation API works in a lambda
	deployFunc := func(ctx *pulumi.Context) error {
		// create private S3 bucket
		_, err := s3.NewBucket(ctx, "s3-website-bucket", &s3.BucketArgs{})
		if err != nil {
			return err
		}

		return nil
	}

	workdir, err := os.MkdirTemp("", "pulumi")
	if err != nil {
		return "", err
	}
	println("Created temporary directory: " + workdir)

	s, err := auto.UpsertStackInlineSource(ctx, account.AccountName, org.OrgName, deployFunc, auto.EnvVars(map[string]string{
		"PULUMI_ACCESS_TOKEN": org.PulumiAccessToken,
		// "PULUMI_BACKEND_URL": "https://app.pulumi.com/flostadler",
	}), auto.Pulumi(pulumiCommand), auto.WorkDir(workdir), auto.PulumiHome("/tmp/.pulumi"))
	s.SetConfig(ctx, "aws:region", auto.ConfigValue{Value: "us-west-2"})
	s.SetConfig(ctx, "aws:accessKey", auto.ConfigValue{Value: account.AwsAccessKey})
	s.SetConfig(ctx, "aws:secretKey", auto.ConfigValue{Value: account.AwsSecretKey})
	s.SetConfig(ctx, "aws:token", auto.ConfigValue{Value: account.AwsSessionToken})
	if err != nil {
		return "", err
	}

	w := s.Workspace()

	err = w.InstallPlugin(ctx, "aws", "v6.32.0")
	if err != nil {
		return "", err
	}

	res, err := s.Up(ctx, optup.SuppressProgress(), optup.ProgressStreams(os.Stdout))
	if err != nil {
		return "", err
	}

	return res.StdOut, nil
}

func downloadFile(filepath string, url string) (err error) {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func installPulumiCLI(ctx context.Context) error {
	err := downloadFile("/tmp/pulumi.tar.gz", pulumiURL)
	if err != nil {
		return err
	}
	defer os.Remove("/tmp/pulumi.tar.gz")
	tempdir, err := os.MkdirTemp("", "pulumi")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempdir)

	err = ExtractTarGz("/tmp/pulumi.tar.gz", tempdir)
	if err != nil {
		return err
	}

	err = os.Rename(filepath.Join(tempdir, "pulumi"), "/tmp/bin")
	if err != nil {
		return err
	}

	pulumiCommand, err = auto.InstallPulumiCommand(ctx, &auto.PulumiCommandOptions{
		Root: "/tmp",
	})
	if err != nil {
		return err
	}
	err = os.Mkdir("/tmp/.pulumi", os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
