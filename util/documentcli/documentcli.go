// Copyright 2023 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package documentcli

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kingpin/v2"
)

func GenerateMarkdown(model *kingpin.ApplicationModel) error {
	file := os.Stdout
	defer file.Close()

	if err := writeHeader(file, model); err != nil {
		return err
	}

	if model.FlagGroupModel != nil && len(model.FlagGroupModel.Flags) > 0 {
		flagRows := [][]string{
			{"Flag", "Description", "Default"},
		}

		for _, flag := range model.FlagGroupModel.Flags {
			if flag.Hidden {
				continue
			}
			flagRows = append(flagRows, createFlagRow(flag))
		}

		if err := writeTable(file, "## Flags", flagRows); err != nil {
			return err
		}
	}

	if model.ArgGroupModel != nil && len(model.ArgGroupModel.Args) > 0 {
		argRows := [][]string{
			{"Argument", "Description", "Default", "Required"},
		}

		for _, arg := range model.ArgGroupModel.Args {
			argRows = append(argRows, createArgRow(arg))
		}

		if err := writeTable(file, "## Arguments", argRows); err != nil {
			return err
		}
	}

	if model.CmdGroupModel != nil {
		cmdRows := [][]string{
			{"Command", "Description"},
		}

		for _, cmd := range model.CmdGroupModel.Commands {
			if cmd.Hidden {
				continue
			}
			cmdRows = append(cmdRows, createCmdRow(cmd))
		}

		if err := writeTable(file, "## Commands", cmdRows); err != nil {
			return err
		}

		if err := writeSubcommands(file, model.Name, model.CmdGroupModel.Commands); err != nil {
			return err
		}
	}

	return nil
}

func writeHeader(file *os.File, model *kingpin.ApplicationModel) error {
	header := fmt.Sprintf("---\ntitle: %s\n\n---\n\n# %s\n\n%s\n\n", model.Name, model.Name, model.Help)
	if _, err := file.WriteString(header); err != nil {
		return err
	}
	return nil
}

func createFlagRow(flag *kingpin.FlagModel) []string {
	var defaultValue string
	if len(flag.Default) > 0 {
		defaultValue = fmt.Sprintf("`%s`", flag.Default[0])
	}

	var flagRow []string
	if flag.Short == '\x00' {
		flagRow = []string{fmt.Sprintf("`--%s`", flag.Name), flag.Help, defaultValue}
	} else {
		flagRow = []string{fmt.Sprintf("`-%c`, `--%s`", flag.Short, flag.Name), flag.Help, defaultValue}
	}

	return flagRow
}

func createArgRow(arg *kingpin.ArgModel) []string {
	var defaultValue string
	if len(arg.Default) > 0 {
		defaultValue = fmt.Sprintf("`%s`", arg.Default[0])
	}

	required := ""
	if arg.Required {
		required = "Yes"
	}

	argRow := []string{arg.Name, arg.Help, defaultValue, required}
	return argRow
}

func createCmdRow(cmd *kingpin.CmdModel) []string {
	if cmd.Hidden {
		return []string{}
	}
	cmdRow := []string{cmd.FullCommand, cmd.Help}
	return cmdRow
}

func writeTable(file *os.File, header string, data [][]string) error {
	if len(data) == 1 {
		return nil
	}

	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("\n\n%s\n\n", header))

	columnsToRender := determineColumnsToRender(data)

	headers := data[0]
	buf.WriteString("|")
	for _, j := range columnsToRender {
		buf.WriteString(fmt.Sprintf(" %s |", headers[j]))
	}
	buf.WriteString("\n")

	buf.WriteString("|")
	for range columnsToRender {
		buf.WriteString(" --- |")
	}
	buf.WriteString("\n")

	for i := 1; i < len(data); i++ {
		row := data[i]
		buf.WriteString("|")
		for _, j := range columnsToRender {
			buf.WriteString(fmt.Sprintf(" %s |", row[j]))
		}
		buf.WriteString("\n")
	}

	if _, err := file.WriteString(strings.TrimSpace(buf.String())); err != nil {
		return err
	}

	return nil
}

func determineColumnsToRender(data [][]string) []int {
	columnsToRender := []int{}
	if len(data) == 0 {
		return columnsToRender
	}
	for j := 0; j < len(data[0]); j++ {
		renderColumn := false
		for i := 1; i < len(data); i++ {
			if data[i][j] != "" {
				renderColumn = true
				break
			}
		}
		if renderColumn {
			columnsToRender = append(columnsToRender, j)
		}
	}
	return columnsToRender
}

func writeSubcommands(file *os.File, modelName string, commands []*kingpin.CmdModel) error {
	for _, cmd := range commands {
		if cmd.Hidden {
			continue
		}

		help := cmd.Help
		if cmd.HelpLong != "" {
			help = cmd.HelpLong
		}
		if _, err := file.WriteString(fmt.Sprintf("\n\n### `%s %s`\n\n%s\n\n", modelName, cmd.FullCommand, help)); err != nil {
			return err
		}

		if cmd.FlagGroupModel != nil && len(cmd.FlagGroupModel.Flags) > 0 {
			flagRows := [][]string{
				{"Flag", "Description", "Default"},
			}

			for _, flag := range cmd.FlagGroupModel.Flags {
				if flag.Hidden {
					continue
				}
				flagRows = append(flagRows, createFlagRow(flag))
			}

			if err := writeTable(file, "### Flags", flagRows); err != nil {
				return err
			}
		}

		if cmd.ArgGroupModel != nil && len(cmd.ArgGroupModel.Args) > 0 {
			argRows := [][]string{
				{"Argument", "Description", "Default", "Required"},
			}

			for _, arg := range cmd.ArgGroupModel.Args {
				argRows = append(argRows, createArgRow(arg))
			}

			if err := writeTable(file, "### Arguments", argRows); err != nil {
				return err
			}
		}

		if cmd.CmdGroupModel != nil && len(cmd.CmdGroupModel.Commands) > 0 {
			if err := writeSubcommands(file, modelName, cmd.CmdGroupModel.Commands); err != nil {
				return err
			}
		}
	}
	return nil
}
