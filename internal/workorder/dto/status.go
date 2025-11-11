package dto

type WorkOrderStatus string

const (
	WOStatusWaitingForInspection WorkOrderStatus = "waiting_for_inspection"
	WOStatusInProgress           WorkOrderStatus = "in_progress"
	WOStatusCompleted            WorkOrderStatus = "completed"
	WOStatusFollowUpNeeded       WorkOrderStatus = "follow_up_needed"
	WOStatusAwaitingInfo         WorkOrderStatus = "awaiting_info"
)
