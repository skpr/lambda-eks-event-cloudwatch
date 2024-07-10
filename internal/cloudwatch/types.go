package cloudwatch

// Event used to parse the CloudWatch Alarm event.
type Event struct {
	AlarmARN  string    `json:"alarmArn"`
	AlarmData AlarmData `json:"alarmData"`
}

// AlarmData used to check the previous and current state of the CloudWatch Alarm.
type AlarmData struct {
	AlarmName     string                 `json:"alarmName"`
	State         AlarmDataState         `json:"state"`
	Configuration AlarmDataConfiguration `json:"configuration"`
}

// AlarmDataState used to check the previous and current state of the CloudWatch Alarm.
type AlarmDataState struct {
	Reason string `json:"reason"`
}

// AlarmDataConfiguration used to review the configuration of the CloudWatch Alarm.
type AlarmDataConfiguration struct {
	Description string `json:"description"`
}
