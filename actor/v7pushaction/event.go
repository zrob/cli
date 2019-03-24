package v7pushaction

type Event struct {
	Type     EventType
	Warnings Warnings
	Err      error
	Data     []PushPlan
}

func NewEvent(eventType EventType, warnings Warnings, err error, data []PushPlan) Event {
	if err != nil {
		return Event{Type: Error, Warnings: warnings, Err: err, Data: data}
	}
	return Event{Type: eventType, Warnings: warnings, Data: data}
}

type EventType string

const (
	Error                           EventType = "Error"
	ApplicationAlreadyExists        EventType = "App already exists"
	ApplyManifest                   EventType = "Applying manifest"
	ApplyManifestComplete           EventType = "Applying manifest Complete"
	BoundRoutes                     EventType = "bound routes"
	BoundServices                   EventType = "bound services"
	ConfiguringServices             EventType = "configuring services"
	CreatedApplication              EventType = "created application"
	CreatedRoutes                   EventType = "created routes"
	CreatingAndMappingRoutes        EventType = "creating and mapping routes"
	CreatingApplication             EventType = "creating application"
	CreatingArchive                 EventType = "creating archive"
	CreatingPackage                 EventType = "creating package"
	PollingBuild                    EventType = "polling build"
	ReadingArchive                  EventType = "reading archive"
	ResourceMatching                EventType = "resource matching"
	RetryUpload                     EventType = "retry upload"
	ScaleWebProcess                 EventType = "scaling the web process"
	ScaleWebProcessComplete         EventType = "scaling the web process complete"
	SetDockerImage                  EventType = "setting docker properties"
	SetDockerImageComplete          EventType = "completed setting docker properties"
	SetDropletComplete              EventType = "set droplet complete"
	SetProcessConfiguration         EventType = "setting configuration on the process"
	SetProcessConfigurationComplete EventType = "completed setting configuration on the process"
	SettingDroplet                  EventType = "setting droplet"
	SettingUpApplication            EventType = "setting up application"
	SkippingApplicationCreation     EventType = "skipping creation"
	StagingComplete                 EventType = "staging complete"
	StartingStaging                 EventType = "starting staging"
	StoppingApplication             EventType = "stopping application"
	StoppingApplicationComplete     EventType = "stopping application complete"
	UnmappingRoutes                 EventType = "unmapping routes"
	UpdatedApplication              EventType = "updated application"
	UploadDropletComplete           EventType = "upload droplet complete"
	UploadingApplication            EventType = "uploading application"
	UploadingApplicationWithArchive EventType = "uploading application with archive"
	UploadingDroplet                EventType = "uploading droplet"
	UploadWithArchiveComplete       EventType = "upload complete"
	Complete                        EventType = "complete"
)
