package version

var Version = "DEV"

func IsDevelopMode() bool {
	v := Version

	return v == "DEV" || v == ""
}
