// Copyright 2022 OnMetal authors
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

package get

import (
	"github.com/onmetal/onmetal-api/orictl-machine/cmd/orictl-machine/orictlmachine/common"
	"github.com/onmetal/onmetal-api/orictl-machine/cmd/orictl-machine/orictlmachine/get/machine"
	"github.com/onmetal/onmetal-api/orictl-machine/cmd/orictl-machine/orictlmachine/get/machineclass"
	clicommon "github.com/onmetal/onmetal-api/orictl/cmd"
	"github.com/spf13/cobra"
)

func Command(streams clicommon.Streams, clientFactory common.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use: "get",
	}

	cmd.AddCommand(
		machine.Command(streams, clientFactory),
		machineclass.Command(streams, clientFactory),
	)

	return cmd
}
