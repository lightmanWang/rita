package commands

import (
	"encoding/csv"
	"os"

	"github.com/activecm/rita/analysis/beacon"
	beaconData "github.com/activecm/rita/datatypes/beacon"
	"github.com/activecm/rita/resources"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

func init() {
	command := cli.Command{
		Name:      "show-beacons",
		Usage:     "Print hosts which show signs of C2 software",
		ArgsUsage: "<database>",
		Flags: []cli.Flag{
			humanFlag,
			configFlag,
		},
		Action: showBeacons,
	}

	bootstrapCommands(command)
}

func showBeacons(c *cli.Context) error {
	db := c.Args().Get(0)
	if db == "" {
		return cli.NewExitError("Specify a database", -1)
	}
	res := resources.InitResources(c.String("config"))
	res.DB.SelectDB(db)

	var data []beaconData.AnalysisView

	ssn := res.DB.Session.Copy()
	resultsView := beacon.GetBeaconResultsView(res, ssn, 0)
	if resultsView == nil {
		return cli.NewExitError("No results were found for "+db, -1)
	}
	resultsView.All(&data)
	ssn.Close()

	if c.Bool("human-readable") {
		err := showBeaconReport(data)
		if err != nil {
			return cli.NewExitError(err.Error(), -1)
		}
		return nil
	}

	err := showBeaconCsv(data)
	if err != nil {
		return cli.NewExitError(err.Error(), -1)
	}
	return nil
}

func showBeaconReport(data []beaconData.AnalysisView) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Score", "Source IP", "Destination IP",
		"Connections", "Avg. Bytes", "Intvl Range", "Size Range", "Top Intvl",
		"Top Size", "Top Intvl Count", "Top Size Count", "Intvl Skew",
		"Size Skew", "Intvl Dispersion", "Size Dispersion", "Intvl Duration"})

	for _, d := range data {
		table.Append(
			[]string{
				f(d.Score), d.Src, d.Dst, i(d.Connections), f(d.AvgBytes),
				i(d.TSIRange), i(d.DSRange), i(d.TSIMode), i(d.DSMode),
				i(d.TSIModeCount), i(d.DSModeCount), f(d.TSISkew), f(d.DSSkew),
				i(d.TSIDispersion), i(d.DSDispersion), f(d.TSDuration),
			},
		)
	}
	table.Render()
	return nil
}

func showBeaconCsv(data []beaconData.AnalysisView) error {
	csvWriter := csv.NewWriter(os.Stdout)
	headers := []string{
		"Score", "Source", "Destination", "Connections",
		"Avg Bytes", "TS Range", "DS Range", "TS Mode", "DS Mode", "TS Mode Count",
		"DS Mode Count", "TS Skew", "DS Skew", "TS Dispersion", "DS Dispersion",
		"TS Duration",
	}
	csvWriter.Write(headers)

	for _, d := range data {
		csvWriter.Write(
			[]string{
				f(d.Score), d.Src, d.Dst, i(d.Connections), f(d.AvgBytes),
				i(d.TSIRange), i(d.DSRange), i(d.TSIMode), i(d.DSMode),
				i(d.TSIModeCount), i(d.DSModeCount), f(d.TSISkew), f(d.DSSkew),
				i(d.TSIDispersion), i(d.DSDispersion), f(d.TSDuration),
			},
		)
	}
	csvWriter.Flush()
	return nil
}
