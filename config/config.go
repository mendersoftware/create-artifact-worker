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
package config

import (
	"github.com/spf13/viper"
)

const (
	//translate to env vars: CREATE_ARTIFACT_<CAPITALIZED>
	CfgServer         = "server"
	CfgSkipVerify     = "skipverify"
	CfgVerbose        = "verbose"
	CfgWorkDir        = "workdir"
	CfgGatewayUrl     = "gateway_url"
	CfgDeploymentsUrl = "deployments_url"
)

func Init() {
	viper.SetEnvPrefix("CREATE_ARTIFACT")
	viper.AutomaticEnv()

	viper.SetDefault(CfgServer, "")
	viper.SetDefault(CfgSkipVerify, false)
	viper.SetDefault(CfgVerbose, false)
	viper.SetDefault(CfgWorkDir, "/var")
}
