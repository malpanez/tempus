package cli

import (
	"fmt"
	"strings"
	"time"

	"tempus/internal/constants"
	tzpkg "tempus/internal/timezone"

	"github.com/spf13/cobra"
)

func NewTimezoneCmd(app *App) *cobra.Command {
	root := &cobra.Command{
		Use:   "timezone",
		Short: "Timezone information and conversion",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List timezones (filterable)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTZList(app, cmd, args)
		},
	}
	listCmd.Flags().String("search", "", "Filter by text (matches IANA, display name, or country)")
	listCmd.Flags().String("country", "", "Filter by country (case-insensitive contains)")
	listCmd.Flags().String("region", "", "Filter by region (supported: europe)")
	listCmd.Flags().Bool("all", false, "Show all known zones (ignores region)")

	infoCmd := &cobra.Command{
		Use:   "info <name-or-IANA>",
		Short: "Show details for a specific timezone",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTZInfo(app, cmd, args)
		},
	}

	root.AddCommand(listCmd, infoCmd)
	return root
}

func runTZList(app *App, cmd *cobra.Command, _ []string) error {
	w := stdoutWriter(app)
	search, _ := cmd.Flags().GetString("search")
	country, _ := cmd.Flags().GetString("country")
	region, _ := cmd.Flags().GetString("region")
	showAll, _ := cmd.Flags().GetBool("all")

	tm := tzpkg.NewTimezoneManager()

	var zones []*tzpkg.TimezoneInfo
	switch {
	case showAll:
		zones = tm.ListTimezones()
	case strings.EqualFold(strings.TrimSpace(region), "europe"):
		zones = tm.GetEuropeanTimezones()
	default:
		zones = tm.ListTimezones()
	}

	search = strings.ToLower(strings.TrimSpace(search))
	country = strings.ToLower(strings.TrimSpace(country))

	filtered := make([]*tzpkg.TimezoneInfo, 0, len(zones))
	for _, z := range zones {
		match := true
		if search != "" {
			if !strings.Contains(strings.ToLower(z.IANA), search) &&
				!strings.Contains(strings.ToLower(z.DisplayName), search) &&
				!strings.Contains(strings.ToLower(z.Country), search) {
				match = false
			}
		}
		if match && country != "" {
			if !strings.Contains(strings.ToLower(z.Country), country) {
				match = false
			}
		}
		if match {
			filtered = append(filtered, z)
		}
	}

	fmt.Fprintf(w, "%-32s  %-7s  %-3s  %-28s  %s\n", "IANA", "Offset", "DST", "Display", "Country")
	for _, z := range filtered {
		dst := "no"
		if z.DST {
			dst = "yes"
		}
		name := CleanDisplay(z.DisplayName)
		fmt.Fprintf(w, "%-32s  %-7s  %-3s  %-28s  %s\n",
			z.IANA, z.Offset, dst, name, z.Country)
	}
	return nil
}

func runTZInfo(app *App, _ *cobra.Command, args []string) error {
	w := stdoutWriter(app)
	query := strings.TrimSpace(strings.Join(args, " "))
	if query == "" {
		return fmt.Errorf("please provide a timezone name or IANA identifier")
	}

	tm := tzpkg.NewTimezoneManager()

	zone, err := tm.GetTimezone(query)
	if err != nil {
		if mapped, mapErr := cityToIANA(query); mapErr == nil && mapped != "" {
			if z2, err2 := tm.GetTimezone(mapped); err2 == nil {
				zone = z2
			}
		}
	}

	if zone == nil {
		sugs := tm.SuggestTimezone(query)
		if len(sugs) == 0 {
			fmt.Fprintln(w, "Timezone not found.")
			return nil
		}
		fmt.Fprintln(w, "Timezone not found. Did you mean:")
		for _, s := range sugs {
			fmt.Fprintf(w, "  - %s (%s) [%s]\n", CleanDisplay(s.DisplayName), s.Country, s.IANA)
		}
		return nil
	}

	loc, err := time.LoadLocation(zone.IANA)
	if err != nil {
		printZoneInfo(w, zone, "", "")
		return nil
	}

	now := time.Now().In(loc)
	printZoneInfo(w, zone, now.Format(constants.DateTimeFormatISOSeconds), now.Format(constants.DateTimeFormatRFC1123))
	return nil
}

func printZoneInfo(w interface{ Write([]byte) (int, error) }, z *tzpkg.TimezoneInfo, local1, local2 string) {
	name := CleanDisplay(z.DisplayName)
	fmt.Fprintf(w, "IANA:       %s\n", z.IANA)
	fmt.Fprintf(w, "Display:    %s\n", name)
	fmt.Fprintf(w, "Country:    %s\n", z.Country)
	fmt.Fprintf(w, "Offset:     %s\n", z.Offset)
	if z.DST {
		fmt.Fprintf(w, "DST:        yes\n")
	} else {
		fmt.Fprintf(w, "DST:        no\n")
	}
	if local1 != "" {
		fmt.Fprintf(w, "Now:        %s\n", local1)
	}
	if local2 != "" {
		fmt.Fprintf(w, "Readable:   %s\n", local2)
	}
}

func cityToIANA(s string) (string, error) {
	x := strings.ToLower(strings.TrimSpace(s))
	if x == "" {
		return "", fmt.Errorf("Unknown city '%s'. Use 'tempus timezone list --search %s' to find the IANA identifier", s, s)
	}

	if x == "melilla" || x == "ceuta" {
		return "Africa/Ceuta", nil
	}
	if x == "canarias" || x == "gran canaria" || x == "tenerife" || x == "las palmas" {
		return "Atlantic/Canary", nil
	}

	switch x {
	case "pelotas", "porto alegre", "porto-alegre", "florianopolis", "florianópolis":
		return "America/Sao_Paulo", nil
	case "campo grande", "campo-grande", "ponta pora", "ponta-porã", "ponta-pora", "dourados":
		return "America/Campo_Grande", nil
	case "cuiaba", "cuiabá":
		return "America/Cuiaba", nil
	case "manaus":
		return "America/Manaus", nil
	case "recife":
		return "America/Recife", nil
	case "belem", "belém":
		return "America/Belem", nil
	case "fortaleza":
		return "America/Fortaleza", nil
	case "salvador":
		return "America/Bahia", nil
	case "rio", "rio de janeiro", "rio-de-janeiro", "niteroi", "niterói":
		return "America/Sao_Paulo", nil
	case "sao paulo", "são paulo", "sao-paulo", "campinas":
		return "America/Sao_Paulo", nil
	}

	if x == "dublin" {
		return "Europe/Dublin", nil
	}
	if x == "london" {
		return "Europe/London", nil
	}

	if x == "madrid" || x == "barcelona" || x == "valencia" || x == "bilbao" || x == "sevilla" {
		return "Europe/Madrid", nil
	}

	return "", fmt.Errorf("Unknown city '%s'. Use 'tempus timezone list --search %s' to find the IANA identifier", s, s)
}
