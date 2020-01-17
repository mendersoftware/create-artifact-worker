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
	"errors"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mendersoftware/create-artifact-worker/config"
	mlog "github.com/mendersoftware/create-artifact-worker/log"
)

const (
	argToken        = "token"
	argArtifactName = "artifact-name"
	argDescription  = "description"
	argDeviceType   = "device-type"
	argArtifactId   = "artifact-id"
	argTenantId     = "tenant-id"
	argArgs         = "args"
)

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
	Server     string
	SkipVerify bool
	Workdir    string

	ArtifactName string
	Description  string
	DeviceType   string
	ArtifactId   string
	FileName     string
	DestDir      string
	TenantId     string
	AuthToken    string
}

func NewSingleFileCmd(cmd *cobra.Command, args []string) (*SingleFileCmd, error) {
	c := &SingleFileCmd{}

	c.Server = viper.GetString(config.CfgServer)
	c.SkipVerify = viper.GetBool(config.CfgSkipVerify)
	c.Workdir = viper.GetString(config.CfgWorkDir)

	// TODO: read other flags/env vars

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *SingleFileCmd) Validate() error {
	if c.Server == "" {
		return errors.New("server address not provided")
	}

	if c.Workdir == "" {
		return errors.New("working directory not provided")
	}

	// TODO other validations, esp. 'args' (json doc)

	return nil
}

func (c *SingleFileCmd) Run() error {
	return nil
}
