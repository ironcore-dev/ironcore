// Copyright 2022 IronCore authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package attach

import (
	"github.com/ironcore-dev/ironcore/irictl-machine/cmd/irictl-machine/irictlmachine/attach/networkinterface"
	"github.com/ironcore-dev/ironcore/irictl-machine/cmd/irictl-machine/irictlmachine/attach/volume"
	"github.com/ironcore-dev/ironcore/irictl-machine/cmd/irictl-machine/irictlmachine/common"
	clicommon "github.com/ironcore-dev/ironcore/irictl/cmd"
	"github.com/spf13/cobra"
)

func Command(streams clicommon.Streams, clientFactory common.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use: "attach",
	}

	cmd.AddCommand(
		volume.Command(streams, clientFactory),
		networkinterface.Command(streams, clientFactory),
	)

	return cmd
}
