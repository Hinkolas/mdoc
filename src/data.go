package src

import "time"

type SystemData struct {
	Date    string
	Time    string
	Version string
}

type TemplateData struct {
	System SystemData
	Body   string
}

// This functions collects system data that will be available to the templates
// e.g. current timestamp, tool version etc.
func CollectSystemData() (*SystemData, error) {

	now := time.Now()

	// TODO: Collect real system data
	return &SystemData{
		Date:    now.Format("02 January 2006"),
		Time:    now.Format("15:04:05"),
		Version: "unknown",
	}, nil

}
