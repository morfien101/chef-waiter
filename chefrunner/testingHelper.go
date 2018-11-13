package chefrunner

// This is a basic implementation of the chef worker that can assit in testing in other package.

//FakeChefRunnerWorker used for testing
// Fake out the things we need to isolate the web package form the rest of chefwaiter.
type FakeChefRunnerWorker struct {
	maintenance bool
}

// OnDemandRun will return a static string with onde to identify that it was a on demand job.
// The string will statify the regex for guids
func (c *FakeChefRunnerWorker) OnDemandRun() string {
	return `onde-1234-1234-1234-1234`
}

// PeriodicRun will return a static string with onde to identify that it was a periodic job.
// The string will statify the regex for guids
func (c *FakeChefRunnerWorker) PeriodicRun() string {
	return `peri-1234-1234-1234-1234`
}

// CustomRun will return a static string with onde to identify that it was a periodic job.
// The string will statify the regex for guids
func (c *FakeChefRunnerWorker) CustomRun(jobDetails string) string {
	return `cust-1234-1234-1234-1234`
}

// InMaintenanceMode will return the maintenace value
func (c *FakeChefRunnerWorker) InMaintenanceMode() bool {
	return c.maintenance
}

// NewFakeChefRunnerWorker will return a chef run worker with a constant maintence window
func NewFakeChefRunnerWorker(inMaintenanceMode bool) *FakeChefRunnerWorker {
	return &FakeChefRunnerWorker{maintenance: inMaintenanceMode}
}
