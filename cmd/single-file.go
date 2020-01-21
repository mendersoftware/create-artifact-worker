// Copyright 2020 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
package cmd

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mendersoftware/create-artifact-worker/client"
	"github.com/mendersoftware/create-artifact-worker/config"
	mlog "github.com/mendersoftware/create-artifact-worker/log"
	"github.com/pkg/errors"
)

const (
	argToken          = "token"
	argArtifactName   = "artifact-name"
	argDescription    = "description"
	argDeviceType     = "device-type"
	argArtifactId     = "artifact-id"
	argGetArtifactUri = "get-artifact-uri"
	argDelArtifactUri = "delete-artifact-uri"
	argTenantId       = "tenant-id"
	argArgs           = "args"
)

type args struct {
	Filename string `json:"filename"`
	DestDir  string `json:"dest_dir"`
}

var singleFileCmd = &cobra.Command{
	Use:   "single-file",
	Short: "Generate an update using a single-file update module.",
	Long: "\nBesides command line args, supports the following env vars:\n\n" +
		"CREATE_ARTIFACT_SERVER root server url (required)\n" +
		"CREATE_ARTIFACT_SKIPVERIFY skip ssl verification (default: false)\n" +
		"CREATE_ARTIFACT_WORKDIR working dir for processing (default: /var)\n" +
		"CREATE_ARTIFACT_GATEWAY_URL public-facing gateway url\n" +
		"CREATE_ARTIFACT_DEPLOYMENTS_URL internal deployments service url\n",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := NewSingleFileCmd(cmd, args)
		if err != nil {
			mlog.Error(err.Error())
			os.Exit(1)
		}

		err = c.Run()
		if err != nil {
			mlog.Error(err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	singleFileCmd.Flags().String(argToken, "", "auth token")
	singleFileCmd.MarkFlagRequired(argToken)

	singleFileCmd.Flags().String(argArtifactName, "", "artifact name")
	singleFileCmd.MarkFlagRequired(argArtifactName)

	singleFileCmd.Flags().String(argArtifactId, "", "artifact id")
	singleFileCmd.MarkFlagRequired(argArtifactId)

	singleFileCmd.Flags().String(argGetArtifactUri, "", "pre-signed s3 url to uploaded temp artifact (GET)")
	singleFileCmd.MarkFlagRequired(argGetArtifactUri)

	singleFileCmd.Flags().String(argDelArtifactUri, "", "pre-signed s3 url to uploaded temp artifact (DELETE)")
	singleFileCmd.MarkFlagRequired(argDelArtifactUri)

	singleFileCmd.Flags().String(argTenantId, "", "tenant id")
	singleFileCmd.MarkFlagRequired(argTenantId)

	singleFileCmd.Flags().String(argDeviceType, "", "device type")
	singleFileCmd.MarkFlagRequired(argDeviceType)

	// json string of specific args: dest dir, file name
	singleFileCmd.Flags().String(argArgs, "", "specific args in json form: {\"file\":<DESTINATION_FILE_NAME_ON_DEVICE>, \"dest_dir\":<DESTINATION_DIR_ON_DEVICE>}")
	singleFileCmd.MarkFlagRequired(argArgs)

	singleFileCmd.Flags().String(argDescription, "", "artifact description")
}

type SingleFileCmd struct {
	ServerUrl      string
	DeploymentsUrl string
	SkipVerify     bool
	Workdir        string

	ArtifactName   string
	Description    string
	DeviceType     string
	ArtifactId     string
	GetArtifactUri string
	DelArtifactUri string
	Args           string
	FileName       string
	DestDir        string
	TenantId       string
	AuthToken      string
}

func NewSingleFileCmd(cmd *cobra.Command, args []string) (*SingleFileCmd, error) {
	c := &SingleFileCmd{}
	c.init(cmd)

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *SingleFileCmd) init(cmd *cobra.Command) error {
	c.ServerUrl = viper.GetString(config.CfgServer)
	c.DeploymentsUrl = viper.GetString(config.CfgDeploymentsUrl)
	c.SkipVerify = viper.GetBool(config.CfgSkipVerify)
	c.Workdir = viper.GetString(config.CfgWorkDir)
	c.SkipVerify = viper.GetBool(config.CfgSkipVerify)

	var arg string
	arg, err := cmd.Flags().GetString(argArtifactName)
	c.ArtifactName = arg
	if err != nil {
		return err
	}

	arg, err = cmd.Flags().GetString(argDescription)
	c.Description = arg
	if err != nil {
		return err
	}

	arg, err = cmd.Flags().GetString(argDeviceType)
	c.DeviceType = arg
	if err != nil {
		return err
	}

	arg, err = cmd.Flags().GetString(argArtifactId)
	c.ArtifactId = arg
	if err != nil {
		return err
	}

	arg, err = cmd.Flags().GetString(argGetArtifactUri)
	c.GetArtifactUri = arg
	if err != nil {
		return err
	}

	arg, err = cmd.Flags().GetString(argDelArtifactUri)
	c.DelArtifactUri = arg
	if err != nil {
		return err
	}

	arg, err = cmd.Flags().GetString(argTenantId)
	c.TenantId = arg
	if err != nil {
		return err
	}

	arg, err = cmd.Flags().GetString(argToken)
	c.AuthToken = arg
	if err != nil {
		return err
	}

	arg, err = cmd.Flags().GetString(argArgs)
	c.Args = arg
	if err != nil {
		return err
	}

	return nil
}

func (c *SingleFileCmd) Validate() error {
	if err := config.ValidUrl(c.ServerUrl); err != nil {
		return errors.Wrap(err, "invalid gateway address")
	}

	if err := config.ValidAbsPath(c.Workdir); err != nil {
		return errors.Wrap(err, "invalid workdir")
	}

	var args args

	err := json.Unmarshal([]byte(c.Args), &args)
	if err != nil {
		return errors.Wrap(err, "can't parse 'args'")
	}

	c.FileName = args.Filename
	c.DestDir = args.DestDir

	if c.FileName == "" {
		return errors.New("destination filename can't be empty")
	}

	if err := config.ValidAbsPath(c.DestDir); err != nil {
		return errors.Wrap(err, "invalid artifact destination dir")
	}

	return nil
}

func (c *SingleFileCmd) Run() error {
	mlog.Info("running single-file update module generation")

	cd, err := client.NewDeployments(c.ServerUrl, c.DeploymentsUrl, c.SkipVerify)
	if err != nil {
		return errors.New("failed to configure 'deployments' client")
	}

	cs3 := client.NewStorage()

	ctx := context.Background()

	downloadDir, err := ioutil.TempDir(c.Workdir, "single-file")

	//gotcha: must download under the correct name (destination name on the device)
	//artifact generator will not allow renaming it
	downloadFile := filepath.Join(downloadDir, c.FileName)

	err = cs3.Download(ctx, c.GetArtifactUri, downloadFile)
	if err != nil {
		return errors.Wrapf(err, "failed to download input file at %s", c.GetArtifactUri)
	}

	// make the filename unique by naming it after the artifact
	outfile := c.ArtifactId + "-generated"
	outfile = filepath.Join(downloadDir, outfile)

	// run gen script
	cmd := exec.Command(
		"/usr/bin/single-file-artifact-gen",
		"-n", c.ArtifactName,
		"-t", c.DeviceType,
		"-d", c.DestDir,
		"-o", outfile,
		downloadFile,
	)

	std, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "single-file-artifact-gen exited with error %s", std)
	}

	err = cs3.Delete(ctx, c.DelArtifactUri)
	if err != nil {
		return errors.Wrapf(err, "failed to delete artifact at %s", c.DelArtifactUri)
	}

	err = cd.UploadArtifactInternal(ctx, outfile, c.ArtifactId, c.TenantId, c.Description)
	if err != nil {
		return errors.Wrapf(err, "failed to upload generated artifact")
	}

	err = os.RemoveAll(downloadDir)
	if err != nil {
		mlog.Error("failed to remove temp working dir %s: %v", downloadDir, err.Error())
	}

	return nil
}
