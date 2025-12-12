package dto

type SubmitWorkRequest struct {
	AssignmentID string `form:"assignment_id" binding:"required,uuid"`
	StudentID    string `form:"student_id" binding:"required,uuid"`
}

type GetReportRequest struct {
	WorkID string `uri:"work_id" binding:"required,uuid"`
}

type HealthCheckResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Storage  string `json:"storage"`
}
